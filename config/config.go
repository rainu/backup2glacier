package config

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"os"
)

const (
	ActionCreate = "CREATE"
	ActionList   = "LIST"
	ActionShow   = "SHOW"
)

const DefaultDatabase = "~/.aws/backup2glacier/database.db"

type Config struct {
	General *GeneralConfig
	Create  *CreateConfig
	Show    *ShowConfig
	List    *ListConfig
}

type GeneralConfig struct {
	Action string   `arg:"positional,required,env:ACTION,help:The action to process. Possible values: CREATE;LIST;SHOW. Default: CREATE."`
	N      []string `arg:"positional"`

	LogLevel string `arg:"-l,env:LOG_LEVEL,help:The log level."`

	argParser *arg.Parser `arg:"-"`
}

type CreateConfig struct {
	DatabaseConfig
	AwsConfig

	Action       string `arg:"positional,required"`
	File         string `arg:"positional,env:FILE,help:The file or folder to backup."`
	AWSVaultName string `arg:"positional,env:AWS_VAULT_NAME,help:The name of the glacier vault."`

	Password     string `arg:"-p,env:PASSWORD,help:The password for encryption."`
	SavePassword bool   `arg:"--save-password,env:SAVE_PASSWORD,help:Should the password save into the database (plain)? Default: false"`

	argParser *arg.Parser `arg:"-"`
}

type ListConfig struct {
	DatabaseConfig

	Action string `arg:"positional,required"`

	argParser *arg.Parser `arg:"-"`
}

type ShowConfig struct {
	DatabaseConfig

	Action   string `arg:"positional,required"`
	BackupId uint   `arg:"positional,required,env:BACKUP_ID,help:The id of the backup to Show."`

	argParser *arg.Parser `arg:"-"`
}

type AwsConfig struct {
	AWSProfile            string `arg:"--aws-profile,env:AWS_PROFILE,help:If you want to use a other AWS profile"`
	AWSPartSize           int    `arg:"--aws-part-size,env:AWS_PART_SIZE,help:The size of each part (except the last) in MiB."`
	AWSArchiveDescription string `arg:"-d,env:AWS_ARCHIVE_DESC,help:The description of the archive."`
}

type DatabaseConfig struct {
	Database string `arg:"--database,env:DATABASE,help:The path to the database. Default is ~/.aws/backup2glacier/database.db"`
}

// NewConfig - Constructor for Config objects
func NewConfig() *Config {
	cfg := &Config{}

	cfg.General = &GeneralConfig{
		LogLevel: "INFO",
		Action:   ActionCreate,
	}

	cfg.General.argParser = arg.MustParse(cfg.General)

	if !isValidAction(cfg.General.Action) {
		cfg.General.Fail("The action is not valid.")
	}

	switch cfg.General.Action {
	case ActionCreate:
		cfg.Create = &CreateConfig{
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
			AwsConfig: AwsConfig{
				AWSPartSize: 1024 * 1024, //1MB chunk
			},
			SavePassword: false,
		}

		cfg.Create.argParser = arg.MustParse(cfg.Create)
	case ActionList:
		cfg.List = &ListConfig{
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
		}

		cfg.List.argParser = arg.MustParse(cfg.List)
	case ActionShow:
		cfg.Show = &ShowConfig{
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
		}

		cfg.Show.argParser = arg.MustParse(cfg.Show)
	}

	return cfg
}

func (c *GeneralConfig) Fail(format string, args ...interface{}) {
	failInternal(c.argParser, format, args...)
}
func (c *CreateConfig) Fail(format string, args ...interface{}) {
	failInternal(c.argParser, format, args...)
}
func (c *ListConfig) Fail(format string, args ...interface{}) {
	failInternal(c.argParser, format, args...)
}
func (c *ShowConfig) Fail(format string, args ...interface{}) {
	failInternal(c.argParser, format, args...)
}
func failInternal(argParser *arg.Parser, format string, args ...interface{}) {
	fmt.Printf(format+"\n\n", args...)
	argParser.WriteHelp(os.Stdout)
	os.Exit(1)
}

func isValidAction(action string) bool {
	switch action {
	case ActionCreate:
		fallthrough
	case ActionShow:
		fallthrough
	case ActionList:
		return true
	default:
		return false
	}
}
