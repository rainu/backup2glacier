package backup

import (
	"backup2glacier/database"
	"backup2glacier/database/model"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/pkg/errors"
	"io"
	"os"
	"regexp"
	"sync"
	"time"
)

type BackupResult struct {
	Vault       string
	ArchiveDesc string
	ArchiveInfo *glacier.ArchiveCreationOutput
	UploadId    *string
	PartSize    int
	TotalSize   int64
	Error       error
}

type BackupCreater interface {
	io.Closer

	Create(files []string, blacklist []*regexp.Regexp, description, vaultName string) *BackupResult
}

type BackupGetter interface {
	io.Closer

	Download(backupId uint, target string) error
}

type BackupDeleter interface {
	Delete(backupId uint) error
}

type BackupManager interface {
	io.Closer

	Create(files []string, blacklist []*regexp.Regexp, description, vaultName string) *BackupResult
	Download(backupId uint, target string) error
	Delete(backupId uint) error
}

type backupManager struct {
	dbRepository database.Repository
	glacier      AWSGlacier

	partSize     int
	savePassword bool
	password     *string
	tier         string
	pollInterval time.Duration
}

func NewBackupCreater(pw string, savePw bool, partSize int, dbUrl string) (BackupCreater, error) {
	return NewBackupManager(&pw, savePw, partSize, time.Millisecond, "", database.NewRepository(dbUrl))
}

func NewBackupGetter(pw *string, tier string, pollInterval time.Duration, dbUrl string) (BackupGetter, error) {
	return NewBackupManager(pw, false, 0, pollInterval, tier, database.NewRepository(dbUrl))
}

func NewBackupDeleter(dbUrl string) (BackupDeleter, error) {
	return NewBackupManager(nil, false, 0, 0, "", database.NewRepository(dbUrl))
}

func NewBackupDeleterForRepository(dbRepo database.Repository) (BackupDeleter, error) {
	return NewBackupManager(nil, false, 0, 0, "", dbRepo)
}

func NewBackupManager(pw *string, savePw bool, partSize int, pollInterval time.Duration, tier string, dbRepo database.Repository) (BackupManager, error) {
	g, err := NewAWSGlacier()
	if err != nil {
		return nil, err
	}

	return &backupManager{
		dbRepository: dbRepo,
		glacier:      g,
		partSize:     partSize,
		pollInterval: pollInterval,
		tier:         tier,
		savePassword: savePw,
		password:     pw,
	}, nil
}

func (b *backupManager) Close() error {
	return b.dbRepository.Close()
}

func (b *backupManager) Create(files []string, blacklist []*regexp.Regexp, description, vaultName string) *BackupResult {
	// folder/file -> zip -> encrypt -> glacier
	srcZip, dstZip := io.Pipe()
	srcCrypt, dstCrypt := io.Pipe()

	wg := sync.WaitGroup{}
	wg.Add(3)

	//save backup intent
	dbBackupEntity := b.saveBackupIntent(description, vaultName)

	//zipping
	contentChan := make(chan *ZipContent, 50)

	go func() {
		defer wg.Done()
		defer dstZip.Close()

		Zip(files, blacklist, dstZip, contentChan)
	}()
	go func() {
		for {
			content, open := <-contentChan
			if !open {
				return
			}

			//store content direct into db
			b.dbRepository.AddContent(dbBackupEntity, &model.Content{
				Path:    content.Realpath,
				Length:  content.Length,
				ModTime: content.FileInfo.ModTime(),
			})
		}
	}()

	//encryption
	go func() {
		defer wg.Done()
		defer dstCrypt.Close()

		crypt := NewCryptModule(*b.password)
		crypt.Encrypt(srcZip, dstCrypt)
	}()

	var err error
	var uploadResult *AWSGlacierUploadResult
	var uploadId *string

	//uploading
	go func() {
		defer wg.Done()
		defer srcCrypt.Close()

		uploadResult, uploadId, err = b.glacier.Upload(AWSGlacierUpload{
			Source:      srcCrypt,
			VaultName:   vaultName,
			ArchiveDesc: description,
			PartSize:    b.partSize,
		})
	}()

	//wait for all to finish
	wg.Wait()

	if uploadResult == nil {
		//avoid nil-pointer if upload fails
		uploadResult = &AWSGlacierUploadResult{}
	}

	result := &BackupResult{
		Vault:       vaultName,
		UploadId:    uploadId,
		ArchiveInfo: uploadResult.CreationResult,
		TotalSize:   uploadResult.TotalSize,
		PartSize:    uploadResult.PartSize,
		ArchiveDesc: uploadResult.ArchiveDesc,
		Error:       err,
	}

	//save to db
	b.updateBackup(result, dbBackupEntity)

	return result
}

func (b *backupManager) saveBackupIntent(description string, vaultName string) *model.Backup {
	dbBackupEntity := &model.Backup{
		Description: description,
		Vault:       vaultName,
	}
	if b.savePassword {
		dbBackupEntity.Password = *b.password
	}
	b.dbRepository.SaveBackup(dbBackupEntity)
	return dbBackupEntity
}

func (b *backupManager) updateBackup(result *BackupResult, dbBackupEntity *model.Backup) {
	if result.Error != nil {
		dbBackupEntity.Error = result.Error.Error()
	}
	dbBackupEntity.UploadId = result.UploadId
	dbBackupEntity.Length = result.TotalSize

	if result.ArchiveInfo != nil {
		dbBackupEntity.ArchiveId = result.ArchiveInfo.ArchiveId
		dbBackupEntity.Checksum = result.ArchiveInfo.Checksum
		dbBackupEntity.Location = result.ArchiveInfo.Location
	}

	b.dbRepository.UpdateBackup(dbBackupEntity)
}

func (b *backupManager) Download(backupId uint, target string) error {
	fTarget, err := os.Create(target)
	if err != nil {
		return errors.Wrap(err, "Could not create target file")
	}
	defer fTarget.Close()

	if err != nil {
		return errors.Wrap(err, "Error while initialise download")
	}

	toDownload := b.dbRepository.GetBackupById(backupId)
	if b.password != nil {
		toDownload.Password = *b.password
	}

	// glacier -> decrypt -> save as zip
	srcCrypt, dstCrypt := io.Pipe()

	wg := sync.WaitGroup{}
	wg.Add(2)

	var downloadErr error
	go func() {
		defer wg.Done()
		defer dstCrypt.Close()

		downloadErr = b.glacier.Download(AWSGlacierDownload{
			VaultName:    toDownload.Vault,
			ArchiveId:    *toDownload.ArchiveId,
			Target:       dstCrypt,
			Tier:         b.tier,
			PollInterval: b.pollInterval,
		})
	}()

	var decryptErr error
	go func() {
		defer wg.Done()

		crypt := NewCryptModule(toDownload.Password)
		decryptErr = crypt.Decrypt(srcCrypt, fTarget)
	}()

	//wait for all to finish
	wg.Wait()

	if downloadErr != nil {
		return errors.Wrap(downloadErr, "Error while downloading from glacier")
	}
	if decryptErr != nil {
		return errors.Wrap(decryptErr, "Error while decrypt content")
	}

	return nil
}

func (b *backupManager) ensureTarget(target string) error {
	handle, err := os.Stat(target)

	if err != nil {
		if os.IsNotExist(err) {
			//if not exists -> create directory
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return errors.Wrap(err, "Could not create directory")
			}
		} else {
			return errors.Wrap(err, "Could not determine target")
		}
	} else {
		if !handle.IsDir() {
			return errors.New("Target '" + target + "' is not a directory")
		}
	}

	return nil
}

func (b *backupManager) Delete(backupId uint) error {
	toDelete := b.dbRepository.GetBackupById(backupId)
	if toDelete.ArchiveId == nil {
		return errors.New("The backup has no archiveId. Was the upload successful?")
	}

	err := b.glacier.Delete(AWSGlacierDelete{
		VaultName: toDelete.Vault,
		ArchiveId: *toDelete.ArchiveId,
	})
	if err != nil {
		return errors.Wrap(err, "Could not delete archive")
	}

	b.dbRepository.DeleteBackupById(backupId)
	return nil
}
