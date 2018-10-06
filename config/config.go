package config

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"os"
)

const (
	ActionCreate = "CREATE"
	ActionList   = "LIST"
)

const DefaultDatabase = "~/.aws/backup2glacier/database.db"

// Config - Collection of all configuration options for this application.
type Config struct {
	Action       string `arg:"positional,required,env:ACTION,help:The action to process. Possible values: CREATE;LIST. Default: CREATE."`
	File         string `arg:"positional,env:FILE,help:The file or folder to backup."`
	AWSVaultName string `arg:"positional,env:AWS_VAULT_NAME,help:The name of the glacier vault."`

	LogLevel string `arg:"-l,env:LOG_LEVEL,help:The log level."`

	Password     string `arg:"-p,env:PASSWORD,help:The password for encryption."`
	SavePassword bool   `arg:"--save-password,env:SAVE_PASSWORD,help:Should the password save into the database (plain)? Default: false"`

	Database string `arg:"--database,env:DATABASE,help:The path to the database. Default is ~/.aws/backup2glacier/database.db"`

	AWSProfile            string `arg:"--aws-profile,env:AWS_PROFILE,help:If you want to use a other AWS profile"`
	AWSPartSize           int    `arg:"--aws-part-size,env:AWS_PART_SIZE,help:The size of each part (except the last) in MiB."`
	AWSArchiveDescription string `arg:"-d,env:AWS_ARCHIVE_DESC,help:The description of the archive."`

	argParser *arg.Parser `arg:"-"`
}

// NewConfig - Constructor for Config objects
func NewConfig() Config {
	cfg := Config{
		LogLevel:     "INFO",
		Action:       ActionCreate,
		Database:     DefaultDatabase,
		SavePassword: false,
		AWSPartSize:  1024 * 1024, //1MB chunk
	}
	cfg.argParser = arg.MustParse(&cfg)

	if !isValidAction(cfg.Action) {
		cfg.Fail("The action is not valid.")
	}

	return cfg
}

func (c *Config) Fail(format string, args ...interface{}) {
	fmt.Printf(format+"\n\n", args...)
	c.argParser.WriteHelp(os.Stdout)
	os.Exit(1)
}

func isValidAction(action string) bool {
	switch action {
	case ActionCreate:
		fallthrough
	case ActionList:
		return true
	default:
		return false
	}
}
