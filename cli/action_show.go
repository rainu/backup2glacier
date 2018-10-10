package cli

import (
	"backup2glacier/config"
	"backup2glacier/database"
	"fmt"
	"github.com/tatsushid/go-prettytable"
	"time"
)

type actionShow struct {
}

func NewShowAction() CliAction {
	return &actionShow{}
}

func (a *actionShow) Do(cfg *config.Config) {
	dbRepository := database.NewRepository(cfg.Show.Database)

	backup, contentIter := dbRepository.GetBackupContentsById(cfg.Show.BackupId)
	defer contentIter.Close()

	if backup.ID == 0 {
		//Not found!
		return
	}

	tbl, err := prettytable.NewTable(
		prettytable.Column{Header: "PATH"},
		prettytable.Column{Header: "LENGTH"},
	)
	if err != nil {
		panic(err)
	}
	tbl.Separator = " | "

	for {
		content, next := contentIter.Next()
		if !next {
			break
		}

		err := tbl.AddRow(
			content.Path,
			content.Length)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf(`Id: %d
Vault: %s
Description: %s
Length: %d
Created at: %s
Archive Id: %s
Upload Id: %s
Location: %s
Password: %s
Error: %s
Content:

`, backup.ID,
		backup.Vault,
		backup.Description,
		backup.Length,
		backup.CreatedAt.Format(time.RFC3339),
		sValue(backup.ArchiveId),
		sValue(backup.UploadId),
		sValue(backup.Location),
		backup.Password,
		backup.Error)

	tbl.Print()
}

func (a *actionShow) Validate(cfg *config.Config) {
	ValidateDatabase(&cfg.Show.DatabaseConfig)
}
