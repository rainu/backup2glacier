package backup

import (
	"backup2glacier/config"
	"fmt"
	"io"
	"sync"
)

type Backup interface {
	Create(file, vaultName string)
}

type backup struct {
	glacier AWSGlacier
	crypt   CryptModule
	config  *config.Config
}

func NewBackup(config *config.Config) (Backup, error) {
	g, err := NewAWSGlacier(config)
	if err != nil {
		return nil, err
	}

	return &backup{
		crypt:   NewCryptModule(config.Password),
		glacier: g,
		config:  config,
	}, nil
}

func (b *backup) Create(file, vaultName string) {
	// folder/file -> zip -> encrypt -> glacier
	srcZip, dstZip := io.Pipe()
	srcCrypt, dstCrypt := io.Pipe()

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		defer dstZip.Close()

		Zip(file, dstZip)
	}()

	go func() {
		defer wg.Done()
		defer dstCrypt.Close()

		b.crypt.Encrypt(srcZip, dstCrypt)
	}()

	go func() {
		defer wg.Done()
		defer srcCrypt.Close()

		result, err := b.glacier.Upload(AWSGlacierUpload{
			Source:    srcCrypt,
			VaultName: vaultName,
			PartSize:  b.config.AWSPartSize,
		})

		fmt.Printf("Result: %+v\nError: %v\n", result, err)
	}()

	wg.Wait()
}
