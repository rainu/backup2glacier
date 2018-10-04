package main

import (
	"backup2glacier/backup"
	"backup2glacier/config"
	. "backup2glacier/log"
)

func main() {
	cfg := config.NewConfig()
	b, err := backup.NewBackupManager(&cfg)

	if err != nil {
		LogFatal("Could not init backup. Error: %v", err)
	}
	defer b.Close()

	b.Create(cfg.File, cfg.AWSArchiveDescription, cfg.AWSVaultName)
}
