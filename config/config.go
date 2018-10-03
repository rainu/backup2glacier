package config

import (
	. "backup2glacier/log"
	"fmt"
	"github.com/alexflint/go-arg"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

var validPartSizes = []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 4096}

// Config - Collection of all configuration options for this application.
type Config struct {
	awsConfig

	LogLevel string `arg:"-l,env:LOG_LEVEL,help:The log level."`

	File string `arg:"-f,env:FILE,help:The file or folder to backup."`

	Password string `arg:"-p,env:PASSWORD,help:The password for encryption."`
}

type awsConfig struct {
	AWSProfile string `arg:"--aws-profile,env:AWS_PROFILE,help:If you want to use a other AWS profile"`

	AWSPartSize           int    `arg:"--aws-part-size,env:AWS_PART_SIZE,help:The size of each part (except the last) in MiB."`
	AWSVaultName          string `arg:"-v,env:AWS_VAULT_NAME,help:The name of the vault."`
	AWSArchiveDescription string `arg:"-d,env:AWS_ARCHIVE_DESC,help:The description of the archive."`
}

// NewConfig - Constructor for Config objects
func NewConfig() Config {
	cfg := Config{
		LogLevel: "INFO",

		awsConfig: awsConfig{
			AWSPartSize: 1024 * 1024, //1MB chunk
		},
	}
	err := arg.Parse(&cfg)
	if err != nil {
		LogFatal("Could not read config. Error: %v", err)
	}

	if cfg.File == "" {
		LogFatal("No file given!")
	}

	if !isValidPartSize(cfg.AWSPartSize) {
		LogFatal("The part size is not valid. Valid sizes are: %+v", validPartSizes)
	}

	cfg.AWSPartSize = 1024 * 1024 * cfg.AWSPartSize

	if cfg.Password == "" {
		cfg.Password = askForPassword()
	}

	return cfg
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
