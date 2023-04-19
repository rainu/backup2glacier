package cli

import (
	"backup2glacier/backup"
	"backup2glacier/config"
	. "backup2glacier/log"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"regexp"
	"syscall"
)

var validPartSizes = []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 4096}

type actionCreate struct {
}

func NewCreateAction() CliAction {
	return &actionCreate{}
}

func (a *actionCreate) Do(cfg *config.Config) {
	b, err := backup.NewBackupCreater(
		cfg.Create.Password,
		cfg.Create.SavePassword,
		cfg.Create.AWSPartSize,
		cfg.Create.Database)

	if err != nil {
		LogFatal("Could not init backup. Error: %v", err)
	}
	defer b.Close()

	result := b.Create(cfg.Create.Files, cfg.Create.GetBlacklist(), cfg.Create.GetWhitelist(), cfg.Create.AWSArchiveDescription, cfg.Create.AWSVaultName)

	if result.Error != nil {
		LogError("Could not upload backup. Error: %v", result.Error)
	} else {
		LogInfo("Successfully upload backup. Result: %+v", result)
	}
}

func (a *actionCreate) Validate(cfg *config.Config) {
	if len(cfg.Create.Files) == 0 {
		cfg.Create.Fail("No file given!")
	}

	for _, curExpr := range cfg.Create.Blacklist {
		if _, err := regexp.Compile(curExpr); err != nil {
			cfg.Create.Fail(fmt.Sprintf(`Blacklist expression is invalid: "%s" Cause: %v`, curExpr, err))
		}
	}

	for _, curExpr := range cfg.Create.Whitelist {
		if _, err := regexp.Compile(curExpr); err != nil {
			cfg.Create.Fail(fmt.Sprintf(`Whitelist expression is invalid: "%s" Cause: %v`, curExpr, err))
		}
	}

	if !isValidPartSize(cfg.Create.AWSPartSize) {
		cfg.Create.Fail("The part size is not valid. Valid sizes are: %+v", validPartSizes)
	}

	cfg.Create.AWSPartSize = 1024 * 1024 * cfg.Create.AWSPartSize

	if cfg.Create.Password == "" {
		cfg.Create.Password = askForPassword()
	}

	if cfg.Create.AWSArchiveDescription == "" {
		cfg.Create.AWSArchiveDescription = fmt.Sprintf("Backup %v to %s", cfg.Create.Files, cfg.Create.AWSVaultName)
	}

	ValidateDatabase(&cfg.Create.DatabaseConfig)
	ValidateAWS(&cfg.Create.AwsGeneralConfig)
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
