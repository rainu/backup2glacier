package config

import (
	. "backup2glacier/log"
	"fmt"
	"github.com/alexflint/go-arg"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"os/user"
	"strings"
	"syscall"
)

var validPartSizes = []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 4096}

const defaultDatabase = "~/.aws/backup2glacier/database.db"

// Config - Collection of all configuration options for this application.
type Config struct {
	File         string `arg:"positional,required,env:FILE,help:The file or folder to backup."`
	AWSVaultName string `arg:"positional,required,env:AWS_VAULT_NAME,help:The name of the glacier vault."`

	LogLevel string `arg:"-l,env:LOG_LEVEL,help:The log level."`

	Password     string `arg:"-p,env:PASSWORD,help:The password for encryption."`
	SavePassword bool   `arg:"--save-password,env:SAVE_PASSWORD,help:Should the password save into the database (plain)? Default: false"`

	Database string `arg:"--database,env:DATABASE,help:The path to the database. Default is ~/.aws/backup2glacier/database.db"`

	AWSProfile            string `arg:"--aws-profile,env:AWS_PROFILE,help:If you want to use a other AWS profile"`
	AWSPartSize           int    `arg:"--aws-part-size,env:AWS_PART_SIZE,help:The size of each part (except the last) in MiB."`
	AWSArchiveDescription string `arg:"-d,env:AWS_ARCHIVE_DESC,help:The description of the archive."`
}

// NewConfig - Constructor for Config objects
func NewConfig() Config {
	cfg := Config{
		LogLevel:     "INFO",
		Database:     defaultDatabase,
		SavePassword: false,
		AWSPartSize:  1024 * 1024, //1MB chunk
	}
	parser := arg.MustParse(&cfg)

	if cfg.File == "" {
		fail(parser, "No file given!")
	}

	if !isValidPartSize(cfg.AWSPartSize) {
		fail(parser, "The part size is not valid. Valid sizes are: %+v", validPartSizes)
	}

	cfg.AWSPartSize = 1024 * 1024 * cfg.AWSPartSize

	if cfg.Password == "" {
		cfg.Password = askForPassword()
	}

	if cfg.AWSArchiveDescription == "" {
		cfg.AWSArchiveDescription = "Backup " + cfg.File + " to " + cfg.AWSVaultName
	}

	if cfg.Database == defaultDatabase {
		usr, _ := user.Current()
		os.MkdirAll(usr.HomeDir+"/.aws/backup2glacier/", os.ModePerm)
	}

	if strings.HasPrefix(cfg.Database, "~/") {
		usr, _ := user.Current()

		cfg.Database = usr.HomeDir + "/" + cfg.Database[2:]
	}

	return cfg
}

func fail(parser *arg.Parser, format string, args ...interface{}) {
	fmt.Printf(format+"\n\n", args...)
	parser.WriteHelp(os.Stdout)
	os.Exit(1)
}

func isValidPartSize(size int) bool {
	for _, valid := range validPartSizes {
		if valid == size {
			return true
		}
	}

	return false
}

func askForPassword() string {
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		LogFatal("Could not read password. Error: %v", err)
	}
	fmt.Println()

	fmt.Print("Repeat Password: ")
	bytePassword2, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		LogFatal("Could not read password. Error: %v", err)
	}
	fmt.Println()

	pw1 := string(bytePassword)
	pw2 := string(bytePassword2)

	if pw1 != pw2 {
		LogFatal("Passwords doesn't match each other!")
	}

	return pw1
}
