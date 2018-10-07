package database

import (
	"backup2glacier/database/model"
	"database/sql"
	"github.com/jinzhu/gorm"
	"io"
)

type BackupIterator interface {
	io.Closer

	Next() (*model.Backup, bool)
}

type backupIterator struct {
	sqlRows *sql.Rows
	db      *gorm.DB
}

func newBackupIterator(sqlRows *sql.Rows, db *gorm.DB) BackupIterator {
	return &backupIterator{
		sqlRows: sqlRows,
		db:      db,
	}
}

func (b *backupIterator) Close() error {
	return b.sqlRows.Close()
}

func (b *backupIterator) Next() (*model.Backup, bool) {
	next := b.sqlRows.Next()
	if !next {
		return nil, false
	}

	var result model.Backup
	b.db.ScanRows(b.sqlRows, &result)

	return &result, true
}
