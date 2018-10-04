package backup

import (
	"backup2glacier/config"
	"backup2glacier/database"
	"backup2glacier/database/model"
	"github.com/aws/aws-sdk-go/service/glacier"
	"io"
	"sync"
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

type BackupManager interface {
	io.Closer

	Create(file, description, vaultName string) *BackupResult
}

type backupManager struct {
	dbRepository database.Repository
	glacier      AWSGlacier
	crypt        CryptModule
	config       *config.Config
}

func NewBackupManager(config *config.Config) (BackupManager, error) {
	g, err := NewAWSGlacier(config)
	if err != nil {
		return nil, err
	}

	return &backupManager{
		dbRepository: database.NewRepository(config.Database),
		glacier:      g,
		crypt:        NewCryptModule(config.Password),
		config:       config,
	}, nil
}

func (b *backupManager) Close() error {
	return b.dbRepository.Close()
}

func (b *backupManager) Create(file, description, vaultName string) *BackupResult {
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

		Zip(file, dstZip, contentChan)
	}()
	go func() {
		for {
			content, open := <-contentChan
			if !open {
				return
			}

			//store content direct into db
			b.dbRepository.AddContent(dbBackupEntity, &model.Content{
				ZipPath:  content.Zippath,
				RealPath: content.Realpath,
				Length:   content.Length,
			})
		}
	}()

	//encryption
	go func() {
		defer wg.Done()
		defer dstCrypt.Close()

		b.crypt.Encrypt(srcZip, dstCrypt)
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
			PartSize:    b.config.AWSPartSize,
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
	if b.config.SavePassword {
		dbBackupEntity.Password = b.config.Password
	}
	b.dbRepository.SaveBackup(dbBackupEntity)
	return dbBackupEntity
}

func (b *backupManager) updateBackup(result *BackupResult, dbBackupEntity *model.Backup) {
	if result.Error != nil {
		dbBackupEntity.Error = result.Error.Error()
	}
	dbBackupEntity.ArchiveId = result.ArchiveInfo.ArchiveId
	dbBackupEntity.UploadId = result.UploadId
	dbBackupEntity.Checksum = result.ArchiveInfo.Checksum
	dbBackupEntity.Location = result.ArchiveInfo.Location
	dbBackupEntity.Length = result.TotalSize

	b.dbRepository.UpdateBackup(dbBackupEntity)
}
