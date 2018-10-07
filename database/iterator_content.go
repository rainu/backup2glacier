package database

import (
	"backup2glacier/database/model"
	"database/sql"
	"github.com/jinzhu/gorm"
	"io"
)

type ContentIterator interface {
	io.Closer

	Next() (*model.Content, bool)
}

type contentIterator struct {
	sqlRows *sql.Rows
	db      *gorm.DB
}

func newContentIterator(sqlRows *sql.Rows, db *gorm.DB) ContentIterator {
	return &contentIterator{
		sqlRows: sqlRows,
		db:      db,
	}
}

func (b *contentIterator) Close() error {
	return b.sqlRows.Close()
}

func (b *contentIterator) Next() (*model.Content, bool) {
	next := b.sqlRows.Next()
	if !next {
		return nil, false
	}

	var result model.Content
	b.db.ScanRows(b.sqlRows, &result)

	return &result, true
}
