package cli

import (
	"backup2glacier/config"
	"backup2glacier/database"
	"encoding/csv"
	"fmt"
	"os"
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

	w := csv.NewWriter(os.Stdout)
	w.UseCRLF = true
	w.Comma = ';'

	err := w.Write([]string{"PATH", "LENGTH", "MODIFY"})
	if err != nil {
		panic(err)
	}

	for {
		content, next := contentIter.Next()
		if !next {
			break
		}

		w.Write([]string{
			content.Path,
			fmt.Sprintf("%d", content.Length),
			content.ModTime.Format(time.RFC3339),
		})

		if err != nil {
			panic(err)
		}
	}
	w.Flush()
}

func (a *actionShow) Validate(cfg *config.Config) {
	ValidateDatabase(&cfg.Show.DatabaseConfig)
}
