package cli

import (
	"backup2glacier/config"
	"backup2glacier/database"
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

type actionList struct {
}

func NewListAction() CliAction {
	return &actionList{}
}

func (a *actionList) Do(cfg *config.Config) {
	dbRepository := database.NewRepository(cfg.List.Database)

	backupIter := dbRepository.List()
	printBackups(backupIter, cfg.List.Factor)
}

func printBackups(iter database.BackupIterator, factor int) []uint {
	defer iter.Close()
	backupIds := make([]uint, 0, 10)

	w := csv.NewWriter(os.Stdout)
	w.UseCRLF = true
	w.Comma = ';'

	err := w.Write([]string{"ID", "CREATED", "VAULT", "DESCRIPTION", "LENGTH", "ARCHIVE_ID"})
	if err != nil {
		panic(err)
	}

	for {
		backup, next := iter.Next()
		if !next {
			break
		}

		backupIds = append(backupIds, backup.ID)

		var sLength string

		if factor == 1 {
			sLength = fmt.Sprintf("%d", backup.Length)
		} else {
			sLength = fmt.Sprintf("%.2f", float64(2684354560)/float64(factor))
		}

		err = w.Write([]string{
			fmt.Sprintf("%d", backup.ID),
			backup.CreatedAt.Format(time.RFC3339),
			backup.Vault,
			backup.Description,
			sLength,
			sValue(backup.ArchiveId),
		})
		if err != nil {
			panic(err)
		}
	}
	w.Flush()

	return backupIds
}

func (a *actionList) Validate(cfg *config.Config) {
	ValidateDatabase(&cfg.List.DatabaseConfig)

	if cfg.List.Factor < 1 {
		cfg.List.Factor = 1
	}
}

func sValue(value *string) string {
	if value != nil {
		return *value
	}

	return ""
}
