package cli

import (
	"backup2glacier/backup"
	"backup2glacier/config"
	"backup2glacier/database"
	. "backup2glacier/log"
	"time"
)

type actionCurator struct {
}

func NewCuratorAction() CliAction {
	return &actionCurator{}
}

func (a *actionCurator) Do(cfg *config.Config) {
	dbRepository := database.NewRepository(cfg.Curator.Database)

	var backupIter database.BackupIterator

	if cfg.Curator.OlderThanTime != nil {
		t := *cfg.Curator.OlderThanTime
		backupIter = dbRepository.GetOlderThan(cfg.Curator.AWSVaultName, t)
	} else if cfg.Curator.MaxAgeDays != 0 {
		t := time.Now().Add(time.Hour * 24 * time.Duration(cfg.Curator.MaxAgeDays) * -1)
		backupIter = dbRepository.GetOlderThan(cfg.Curator.AWSVaultName, t)
	} else {
		backupIter = dbRepository.GetLast(cfg.Curator.AWSVaultName, cfg.Curator.KeepN)
	}

	backupIds := printBackups(backupIter, 1)

	if len(backupIds) == 0 {
		LogInfo("Nothing to do.")
		return
	}

	if !cfg.Curator.DontAsk {
		if !askToBeSure() {
			LogFatal("Curator cancelled!")
			return
		}
	}

	//delete all
	b, err := backup.NewBackupDeleterForRepository(dbRepository)
	if err != nil {
		LogFatal("Could not delete backup. Error: %v", err)
	}

	for _, backupId := range backupIds {
		err = b.Delete(backupId)
		if err != nil {
			LogFatal("Error while delete backup. Error: %v", err)
		}
	}
}

func (a *actionCurator) Validate(cfg *config.Config) {
	ValidateDatabase(&cfg.Curator.DatabaseConfig)
	ValidateAWS(&cfg.Curator.AwsGeneralConfig)

	if cfg.Curator.OlderThanTime == nil && cfg.Curator.MaxAgeDays == 0 && cfg.Curator.KeepN == 0 {
		cfg.Curator.Fail("Even OlderThan, MaxAge or Keep must be given!")
	}
}
