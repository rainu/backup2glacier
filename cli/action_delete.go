package cli

import (
	"backup2glacier/backup"
	"backup2glacier/config"
	. "backup2glacier/log"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type actionDelete struct {
}

func NewDeleteAction() CliAction {
	return &actionDelete{}
}

func (a *actionDelete) Do(cfg *config.Config) {
	if !cfg.Delete.DontAsk {
		if !askToBeSure() {
			LogFatal("Backup deletion cancelled!")
			return
		}
	}

	b, err := backup.NewBackupDeleter(cfg.Delete.Database)
	if err != nil {
		LogFatal("Could not delete backup. Error: %v", err)
	}

	err = b.Delete(cfg.Delete.BackupId)
	if err != nil {
		LogFatal("Error while delete backup. Error: %v", err)
	}
}

func (a *actionDelete) Validate(cfg *config.Config) {
	ValidateDatabase(&cfg.Delete.DatabaseConfig)
	ValidateAWS(&cfg.Delete.AwsGeneralConfig)
}

func askToBeSure() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Are you sure to delete the backup? (y/n): ")
	answer, _ := reader.ReadString('\n')

	switch strings.ToLower(strings.Trim(answer, "\n")) {
	case "y":
		fallthrough
	case "yes":
		return true
	}

	return false
}
