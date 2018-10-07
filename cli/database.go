package cli

import (
	"backup2glacier/config"
	"os"
	"os/user"
	"strings"
)

func ValidateDatabase(cfg *config.DatabaseConfig) {
	if cfg.Database == config.DefaultDatabase {
		usr, _ := user.Current()
		os.MkdirAll(usr.HomeDir+"/.aws/backup2glacier/", os.ModePerm)
	}

	if strings.HasPrefix(cfg.Database, "~/") {
		usr, _ := user.Current()

		cfg.Database = usr.HomeDir + "/" + cfg.Database[2:]
	}
}
