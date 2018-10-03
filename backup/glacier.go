package backup

import (
	"backup2glacier/config"
	. "backup2glacier/log"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

type AWSGlacierUpload struct {
	Source      io.Reader
	VaultName   string
	ArchiveDesc string
	PartSize    int
}

type AWSGlacier interface {
	Upload(AWSGlacierUpload) (*glacier.ArchiveCreationOutput, error)
}

type awsGlacier struct {
	session *session.Session
	glacier *glacier.Glacier
}

func NewAWSGlacier(config *config.Config) (AWSGlacier, error) {
	s, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "Error while create new AWS session")
	}

	return &awsGlacier{
		s,
		glacier.New(s),
	}, nil
}

func (a *awsGlacier) Upload(upload AWSGlacierUpload) (*glacier.ArchiveCreationOutput, error) {
	initResult, err := a.initUpload(upload.VaultName, upload.ArchiveDesc, upload.PartSize)
	if err != nil {
		return nil, errors.Wrap(err, "Could not init multipart upload")
	}
	LogInfo("Initialise multipart upload: %+v", initResult)

	//closure for reusing purposes
	abortMultipartUpload := func() {
		request := &glacier.AbortMultipartUploadInput{
			AccountId: aws.String("-"),
			VaultName: aws.String(upload.VaultName),
			UploadId:  initResult.UploadId,
		}
		LogDebug("Send AbortMultipartUpload: %+v", request)
		a.glacier.AbortMultipartUpload(request)
	}

	//if any panic occurred: try to abort the upload
	defer func() {
		if r := recover(); r != nil {
			abortMultipartUpload()
			panic(r)
		}
	}()

	totalBytes, hashes, err := a.uploadParts(upload.Source, upload.VaultName, initResult.UploadId, upload.PartSize)
	if err != nil {
		abortMultipartUpload()
		return nil, errors.Wrap(err, "Failed to upload to glacier. Upload aborted")
	}

	completeResult, err := a.completeUpload(upload.VaultName, initResult.UploadId, hashes, totalBytes)
	if err != nil {
		abortMultipartUpload()
		return nil, errors.Wrap(err, "Error while completing multipart upload")
	}
	LogInfo("Complete multipart upload: %+v", completeResult)

	return completeResult, nil
}

func (a *awsGlacier) initUpload(vaultName, archiveDesc string, partSize int) (*glacier.InitiateMultipartUploadOutput, error) {
	request := &glacier.InitiateMultipartUploadInput{
		AccountId:          aws.String("-"),
		PartSize:           aws.String(strconv.Itoa(partSize)),
		VaultName:          aws.String(vaultName),
		ArchiveDescription: aws.String(archiveDesc),
	}
	LogDebug("Send InitiateMultipartUpload: %+v", request)
	result, err := a.glacier.InitiateMultipartUpload(request)

	return result, err
}

func (a *awsGlacier) uploadParts(src io.Reader, vaultName string, uploadId *string, partSize int) (int64, [][]byte, error) {
	var hashes [][]byte
	var readBytes int64
	readBytes = 0

	for {
		written, result, err := a.uploadPart(src, vaultName, uploadId, partSize, readBytes)
		if err != nil {
			if err != io.EOF {
				return -1, nil, errors.Wrap(err, "Failed upload part to glacier")
			}
			break
		}

		readBytes += int64(written)

		hash, _ := hex.DecodeString(*result.Checksum)
		hashes = append(hashes, hash)
	}

	return readBytes, hashes, nil
}

func (a *awsGlacier) uploadPart(src io.Reader, vaultName string, uploadId *string, partSize int, totalWritten int64) (int64, *glacier.UploadMultipartPartOutput, error) {
	tmpFile, err := ioutil.TempFile("", "aws-glacier-part")
	if err != nil {
		return 0, nil, errors.Wrap(err, "Could not create temporary file for part upload")
	}
	defer os.Remove(tmpFile.Name())

	written, err := io.CopyN(tmpFile, src, int64(partSize))
	if written == 0 {
		//maybe we have reached the EOF
		return written, nil, err
	}

	tmpFile.Seek(0, 0)

	request := &glacier.UploadMultipartPartInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(vaultName),
		UploadId:  uploadId,
		Range:     aws.String(fmt.Sprintf("bytes %d-%d/*", totalWritten, totalWritten+written-1)),
		Body:      tmpFile,
	}

	LogDebug("Send UploadMultipartPart: %+v", request)
	result, err := a.glacier.UploadMultipartPart(request)

	if err != nil {
		return 0, nil, errors.Wrap(err, "Error while uploading part")
	}
	LogInfo("Uploaded multipart part: %+v", result)

	return written, result, nil
}

func (a *awsGlacier) completeUpload(vaultName string, uploadId *string, hashes [][]byte, totalBytes int64) (*glacier.ArchiveCreationOutput, error) {
	treeHash := glacier.ComputeTreeHash(hashes)
	request := &glacier.CompleteMultipartUploadInput{
		AccountId:   aws.String("-"),
		VaultName:   aws.String(vaultName),
		UploadId:    uploadId,
		ArchiveSize: aws.String(fmt.Sprintf("%d", totalBytes)),
		Checksum:    aws.String(fmt.Sprintf("%x", treeHash)),
	}
	LogDebug("Send CompleteMultipartUpload: %+v", request)
	result, err := a.glacier.CompleteMultipartUpload(request)

	return result, err
}
