package cli

import (
	"backup2glacier/backup"
	"backup2glacier/config"
	. "backup2glacier/log"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

var validPartSizes = []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 4096}

type actionCreate struct {
}

func NewCreateAction() CliAction {
	return &actionCreate{}
}

func (a *actionCreate) Do(cfg *config.Config) {
	b, err := backup.NewBackupManager(cfg)

	if err != nil {
		LogFatal("Could not init backup. Error: %v", err)
	}
	defer b.Close()

	b.Create(cfg.File, cfg.AWSArchiveDescription, cfg.AWSVaultName)
}

func (a *actionCreate) Validate(cfg *config.Config) {
	if cfg.File == "" {
		cfg.Fail("No file given!")
	}

	if !isValidPartSize(cfg.AWSPartSize) {
		cfg.Fail("The part size is not valid. Valid sizes are: %+v", validPartSizes)
	}

	cfg.AWSPartSize = 1024 * 1024 * cfg.AWSPartSize

	if cfg.Password == "" {
		cfg.Password = askForPassword()
	}

	if cfg.AWSArchiveDescription == "" {
		cfg.AWSArchiveDescription = "Backup " + cfg.File + " to " + cfg.AWSVaultName
	}

	ValidateDatabase(cfg)
}

func askForPassword() string {
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		LogFatal("Could not read password. Error: %v", err)
	}
	fmt.Println()

	fmt.Print("Repeat Password: ")
	bytePassword2, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		LogFatal("Could not read password. Error: %v", err)
	}
	fmt.Println()

	pw1 := string(bytePassword)
	pw2 := string(bytePassword2)

	if pw1 != pw2 {
		LogFatal("Passwords doesn't match each other!")
	}

	return pw1
}

func isValidPartSize(size int) bool {
	for _, valid := range validPartSizes {
		if valid == size {
			return true
		}
	}

	return false
}
