package cli

import "backup2glacier/config"

type CliAction interface {
	Do(config *config.Config)
	Validate(config *config.Config)
}
