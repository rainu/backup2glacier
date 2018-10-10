package cli

import (
	"backup2glacier/config"
	"os"
)

func ValidateAWS(cfg *config.AwsGeneralConfig) {
	if cfg.AWSProfile != "" {
		os.Setenv("AWS_PROFILE", cfg.AWSProfile)
	}
}
