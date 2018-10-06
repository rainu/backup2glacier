package database

import (
	"backup2glacier/database/model"
	_ "github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/github"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"io"
	"reflect"
)

type Repository interface {
	io.Closer

	SaveBackup(backup *model.Backup)
	UpdateBackup(backup *model.Backup)
	AddContent(backup *model.Backup, content *model.Content)

	Count() int64
	List() []model.Backup
}

type repository struct {
	db *gorm.DB
}

func NewRepository(dbFile string) Repository {
	db, err := gorm.Open("sqlite3", dbFile)
	if err != nil {
		panic(errors.Wrap(err, "failed to connect database"))
	}

	// Migrate the schema
	db.AutoMigrate(&model.Content{})
	db.AutoMigrate(&model.Backup{})

	return &repository{
		db,
	}
}

func (r *repository) Close() error {
	return r.db.Close()
}

func (r *repository) SaveBackup(backup *model.Backup) {
	r.db.Create(backup)
}

func (r *repository) UpdateBackup(backup *model.Backup) {
	r.db.Save(backup)
}

func (r *repository) AddContent(backup *model.Backup, content *model.Content) {
	content.BackupID = backup.ID

	r.db.Create(content)
}

func (r *repository) Count() int64 {
	var count int64
	r.db.Table(reflect.TypeOf(&model.Backup{}).Name()).Count(&count)

	return count
}

func (r *repository) List() []model.Backup {
	var backups []model.Backup
	r.db.Find(&backups)

	return backups
}
