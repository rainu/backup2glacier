package config

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"os"
	"time"
)

const (
	ActionCreate = "CREATE"
	ActionList   = "LIST"
	ActionShow   = "SHOW"
	ActionGet    = "GET"
	ActionDelete = "DELETE"
)

const DefaultDatabase = "~/.aws/backup2glacier/database.db"

type Config struct {
	Action string

	Create *CreateConfig
	Get    *GetConfig
	Delete *DeleteConfig
	Show   *ShowConfig
	List   *ListConfig
}

type CreateConfig struct {
	GeneralConfig
	DatabaseConfig
	AwsGeneralConfig

	AWSVaultName string   `arg:"positional,env:AWS_VAULT_NAME,help:The name of the glacier vault."`
	Files        []string `arg:"positional,env:FILE,help:The file or folder to backup."`

	AWSPartSize           int    `arg:"--aws-part-size,env:AWS_PART_SIZE,help:The size of each part (except the last) in MiB."`
	AWSArchiveDescription string `arg:"-d,env:AWS_ARCHIVE_DESC,help:The description of the archive."`

	Password     string `arg:"-p,env:PASSWORD,help:The password for encryption."`
	SavePassword bool   `arg:"--save-password,env:SAVE_PASSWORD,help:Should the password save into the database (plain)? Default: false"`

	argParser *arg.Parser `arg:"-"`
}

type ListConfig struct {
	GeneralConfig
	DatabaseConfig

	argParser *arg.Parser `arg:"-"`
}

type ShowConfig struct {
	GeneralConfig
	DatabaseConfig

	BackupId uint `arg:"positional,required,env:BACKUP_ID,help:The id of the backup to Show."`

	argParser *arg.Parser `arg:"-"`
}

type GetConfig struct {
	GeneralConfig
	DatabaseConfig
	AwsGeneralConfig

	BackupId uint   `arg:"positional,required,env:BACKUP_ID,help:The id of the backup to get."`
	File     string `arg:"positional,env:FILE,help:The target zip path."`

	AWSTier         string        `arg:"--aws-tier,env:AWS_TIER,help:The tier to use for the archive retrieval job. Default: Standard. Possible: Expedited;Standard;Bulk"`
	AWSPollInterval time.Duration `arg:"--aws-poll-interval,env:AWS_POLL_INTERVAL,help:The interval to poll job status. Default: 30min."`
	Password        *string       `arg:"-p,env:PASSWORD,help:The password for decryption. If no password is given it will use the one in the database"`

	argParser *arg.Parser `arg:"-"`
}

type DeleteConfig struct {
	GeneralConfig
	DatabaseConfig
	AwsGeneralConfig

	BackupId uint `arg:"positional,required,env:BACKUP_ID,help:The id of the backup to get."`
	DontAsk  bool `arg:"-y,env:DONT_ASK,help:Dont ask if you be sure to delete backup."`

	argParser *arg.Parser `arg:"-"`
}

type GeneralConfig struct {
	LogLevel string `arg:"-l,env:LOG_LEVEL,help:The log level."`
}

type AwsGeneralConfig struct {
	AWSProfile string `arg:"--aws-profile,env:AWS_PROFILE,help:If you want to use a other AWS profile"`
}

type DatabaseConfig struct {
	Database string `arg:"--database,env:DATABASE,help:The path to the database. Default is ~/.aws/backup2glacier/database.db"`
}

// NewConfig - Constructor for Config objects
func NewConfig() *Config {
	cfg := &Config{}

	if len(os.Args) <= 1 {
		fmt.Printf("You have to specify a subcommand: %v\n", []string{ActionCreate, ActionGet, ActionDelete, ActionList, ActionShow})
		os.Exit(2)
	}
	cfg.Action = os.Args[1]

	if !isValidAction(cfg.Action) {
		fmt.Printf("You have to specify a valid subcommand: %v\n", []string{ActionCreate, ActionGet, ActionDelete, ActionList, ActionShow})
		os.Exit(2)
	}

	switch cfg.Action {
	case ActionCreate:
		cfg.Create = &CreateConfig{
			GeneralConfig: GeneralConfig{
				LogLevel: "INFO",
			},
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
			AWSPartSize:  1024 * 1024, //1MB chunk
			SavePassword: false,
		}

		cfg.Create.argParser, _ = arg.NewParser(arg.Config{}, cfg.Create)
		cfg.Create.argParser.Parse(os.Args[2:])
	case ActionGet:
		cfg.Get = &GetConfig{
			GeneralConfig: GeneralConfig{
				LogLevel: "INFO",
			},
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
			AWSPollInterval: 30 * time.Minute,
			AWSTier:         "Standard",
		}

		cfg.Get.argParser, _ = arg.NewParser(arg.Config{}, cfg.Get)
		cfg.Get.argParser.Parse(os.Args[2:])
	case ActionList:
		cfg.List = &ListConfig{
			GeneralConfig: GeneralConfig{
				LogLevel: "INFO",
			},
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
		}

		cfg.List.argParser, _ = arg.NewParser(arg.Config{}, cfg.List)
		cfg.List.argParser.Parse(os.Args[2:])
	case ActionShow:
		cfg.Show = &ShowConfig{
			GeneralConfig: GeneralConfig{
				LogLevel: "INFO",
			},
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
		}

		cfg.Show.argParser, _ = arg.NewParser(arg.Config{}, cfg.Show)
		cfg.Show.argParser.Parse(os.Args[2:])
	case ActionDelete:
		cfg.Delete = &DeleteConfig{
			GeneralConfig: GeneralConfig{
				LogLevel: "INFO",
			},
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},

			DontAsk: false,
		}

		cfg.Delete.argParser, _ = arg.NewParser(arg.Config{}, cfg.Delete)
		cfg.Delete.argParser.Parse(os.Args[2:])
	}

	return cfg
}

func (c *CreateConfig) Fail(format string, args ...interface{}) {
	failInternal(c.argParser, format, args...)
}
func (c *GetConfig) Fail(format string, args ...interface{}) {
	failInternal(c.argParser, format, args...)
}
func (c *DeleteConfig) Fail(format string, args ...interface{}) {
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
	case ActionGet:
		fallthrough
	case ActionDelete:
		fallthrough
	case ActionShow:
		fallthrough
	case ActionList:
		return true
	default:
		return false
	}
}
