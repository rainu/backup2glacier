package cli

import (
	"backup2glacier/backup"
	"backup2glacier/config"
	. "backup2glacier/log"
)

var validTiers = []string{"Expedited", "Standard", "Bulk"}

type actionGet struct {
}

func NewGetAction() CliAction {
	return &actionGet{}
}

func (a *actionGet) Do(cfg *config.Config) {
	b, err := backup.NewBackupGetter(
		cfg.Get.Password,
		cfg.Get.AWSTier,
		cfg.Get.AWSPollInterval,
		cfg.Get.Database)

	if err != nil {
		LogFatal("Could not init backup. Error: %v", err)
	}
	defer b.Close()

	b.Download(cfg.Get.BackupId, cfg.Get.File)
}

func (a *actionGet) Validate(cfg *config.Config) {
	if cfg.Get.File == "" {
		cfg.Get.Fail("No file given!")
	}

	if !isValidTier(cfg.Get.AWSTier) {
		cfg.Get.Fail("The part size is not valid. Valid sizes are: %+v", validPartSizes)
	}

	ValidateDatabase(&cfg.Get.DatabaseConfig)
	ValidateAWS(&cfg.Get.AwsGeneralConfig)
}

func isValidTier(tier string) bool {
	for _, valid := range validTiers {
		if valid == tier {
			return true
		}
	}

	return false
}
