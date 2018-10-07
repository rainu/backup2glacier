package cli

import (
	"backup2glacier/config"
	"backup2glacier/database"
	"github.com/tatsushid/go-prettytable"
	"time"
)

type actionList struct {
}

func NewListAction() CliAction {
	return &actionList{}
}

func (a *actionList) Do(cfg *config.Config) {
	dbRepository := database.NewRepository(cfg.List.Database)

	tbl, err := prettytable.NewTable(
		prettytable.Column{Header: "ID"},
		prettytable.Column{Header: "CREATED"},
		prettytable.Column{Header: "VAULT"},
		prettytable.Column{Header: "DESCRIPTION"},
		prettytable.Column{Header: "LENGTH"},
		prettytable.Column{Header: "ARCHIVE_ID"},
	)
	if err != nil {
		panic(err)
	}
	tbl.Separator = " | "

	backupIter := dbRepository.List()
	defer backupIter.Close()

	for {
		backup, next := backupIter.Next()
		if !next {
			break
		}

		err := tbl.AddRow(
			backup.ID,
			backup.CreatedAt.Format(time.RFC3339),
			backup.Vault,
			backup.Description,
			backup.Length,
			sValue(backup.ArchiveId))
		if err != nil {
			panic(err)
		}
	}

	tbl.Print()
}

func (a *actionList) Validate(cfg *config.Config) {
	ValidateDatabase(&cfg.List.DatabaseConfig)
}

func sValue(value *string) string {
	if value != nil {
		return *value
	}

	return ""
}
