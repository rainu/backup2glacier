package config

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"os"
	"regexp"
	"time"
)

const (
	ActionCreate  = "CREATE"
	ActionList    = "LIST"
	ActionShow    = "SHOW"
	ActionGet     = "GET"
	ActionDelete  = "DELETE"
	ActionCurator = "CURATOR"
)

const DefaultDatabase = "~/.aws/backup2glacier/database.db"

type Config struct {
	Action string

	Create  *CreateConfig
	Get     *GetConfig
	Delete  *DeleteConfig
	Show    *ShowConfig
	List    *ListConfig
	Curator *CuratorConfig
}

type CreateConfig struct {
	GeneralConfig
	DatabaseConfig
	AwsGeneralConfig

	AWSVaultName string   `arg:"positional,env:AWS_VAULT_NAME,help:The name of the glacier vault."`
	Files        []string `arg:"positional,env:FILE,help:The file or folder to backup."`
	Blacklist    []string `arg:"-b,separate,env:BLACKLIST,help:Regular expressions of files that should be excluded."`
	Whitelist    []string `arg:"-w,separate,env:WHITELIST,help:Regular expressions of files that should be included even if their would be excluded by blacklist."`

	AWSPartSize           int    `arg:"--aws-part-size,env:AWS_PART_SIZE,help:The size of each part (except the last) in MiB."`
	AWSArchiveDescription string `arg:"-d,env:AWS_ARCHIVE_DESC,help:The description of the archive."`

	Password     string `arg:"-p,env:PASSWORD,help:The password for encryption."`
	SavePassword bool   `arg:"--save-password,env:SAVE_PASSWORD,help:Should the password save into the database (plain)? Default: false"`

	argParser *arg.Parser `arg:"-"`
}

type ListConfig struct {
	GeneralConfig
	DatabaseConfig

	Factor int  `arg:"--factor,env:FACTOR,help:Conversion factor for the size specification. For example: Use 1024 for KiB.Default:1"`
	Kb     bool `arg:"--kb,env:FACTOR_KB,help:Use KB as conversion factor."`
	Kib    bool `arg:"--kib,env:FACTOR_KIB,help:Use KiB as conversion factor."`
	Mb     bool `arg:"--mb,env:FACTOR_MB,help:Use MB as conversion factor."`
	Mib    bool `arg:"--mib,env:FACTOR_MIB,help:Use MiB as conversion factor."`
	Gb     bool `arg:"--gb,env:FACTOR_MB,help:Use GB as conversion factor."`
	Gib    bool `arg:"--gib,env:FACTOR_MIB,help:Use GiB as conversion factor."`

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

type CuratorConfig struct {
	GeneralConfig
	DatabaseConfig
	AwsGeneralConfig

	AWSVaultName string `arg:"positional,env:AWS_VAULT_NAME,help:The name of the glacier vault."`
	DontAsk      bool   `arg:"-y,env:DONT_ASK,help:Dont ask if you be sure to delete backup."`

	OlderThanTime *time.Time `arg:"--older-than,env:OLDER_THAN,help:Backups which are older than this time."`
	MaxAgeDays    int        `arg:"--max-age,env:MAX_AGE,help:Backups which are older than x days."`
	KeepN         int        `arg:"--keep,env:LAST,help:Keep only the last n backups."`

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
		fmt.Printf("You have to specify a subcommand: %v\n", []string{ActionCreate, ActionGet, ActionDelete, ActionList, ActionShow, ActionCurator})
		os.Exit(2)
	}
	cfg.Action = os.Args[1]

	if !isValidAction(cfg.Action) {
		fmt.Printf("You have to specify a valid subcommand: %v\n", []string{ActionCreate, ActionGet, ActionDelete, ActionList, ActionShow, ActionCurator})
		os.Exit(2)
	}

	var err error
	var argParser *arg.Parser

	switch cfg.Action {
	case ActionCreate:
		cfg.Create = &CreateConfig{
			GeneralConfig: GeneralConfig{
				LogLevel: "INFO",
			},
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
			AWSPartSize:  1, //1MB chunk
			SavePassword: false,
		}

		cfg.Create.argParser, _ = arg.NewParser(arg.Config{}, cfg.Create)
		argParser = cfg.Create.argParser
		err = cfg.Create.argParser.Parse(os.Args[2:])
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
		argParser = cfg.Get.argParser
		err = cfg.Get.argParser.Parse(os.Args[2:])
	case ActionList:
		cfg.List = &ListConfig{
			GeneralConfig: GeneralConfig{
				LogLevel: "INFO",
			},
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},
		}

		cfg.List.argParser, err = arg.NewParser(arg.Config{}, cfg.List)
		argParser = cfg.List.argParser
		err = cfg.List.argParser.Parse(os.Args[2:])
		if err == nil {
			if cfg.List.Kb {
				cfg.List.Factor = 1000
			} else if cfg.List.Kib {
				cfg.List.Factor = 1024
			} else if cfg.List.Mb {
				cfg.List.Factor = 1000 * 1000
			} else if cfg.List.Mib {
				cfg.List.Factor = 1024 * 1024
			} else if cfg.List.Gb {
				cfg.List.Factor = 1000 * 1000 * 1000
			} else if cfg.List.Gib {
				cfg.List.Factor = 1024 * 1024 * 1024
			}
		}
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
		argParser = cfg.Show.argParser
		err = cfg.Show.argParser.Parse(os.Args[2:])
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
		argParser = cfg.Delete.argParser
		err = cfg.Delete.argParser.Parse(os.Args[2:])
	case ActionCurator:
		cfg.Curator = &CuratorConfig{
			GeneralConfig: GeneralConfig{
				LogLevel: "INFO",
			},
			DatabaseConfig: DatabaseConfig{
				Database: DefaultDatabase,
			},

			DontAsk: false,
		}

		cfg.Curator.argParser, _ = arg.NewParser(arg.Config{}, cfg.Curator)
		argParser = cfg.Curator.argParser
		err = cfg.Curator.argParser.Parse(os.Args[2:])
	}

	if err != nil {
		if err == arg.ErrHelp {
			argParser.WriteHelp(os.Stdout)
			os.Exit(0)
		}

		fmt.Printf("Error while parsing arguments: %s", err.Error())
		os.Exit(3)
	}

	return cfg
}

func (c *CreateConfig) GetBlacklist() []*regexp.Regexp {
	var result []*regexp.Regexp

	for _, curEntry := range c.Blacklist {
		result = append(result, regexp.MustCompile(curEntry))
	}

	return result
}

func (c *CreateConfig) GetWhitelist() []*regexp.Regexp {
	var result []*regexp.Regexp

	for _, curEntry := range c.Whitelist {
		result = append(result, regexp.MustCompile(curEntry))
	}

	return result
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
func (c *CuratorConfig) Fail(format string, args ...interface{}) {
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
		fallthrough
	case ActionCurator:
		return true
	default:
		return false
	}
}
