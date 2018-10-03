package main

import (
	"backup2glacier/backup"
	"backup2glacier/config"
	. "backup2glacier/log"
)

func main() {
	cfg := config.NewConfig()
	b, err := backup.NewBackup(&cfg)

	if err != nil {
		LogFatal("Could not init backup. Error: %v", err)
	}

	b.Create(cfg.File, cfg.AWSVaultName)
}
