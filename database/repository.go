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
	"time"
)

type Repository interface {
	io.Closer

	SaveBackup(backup *model.Backup)
	UpdateBackup(backup *model.Backup)
	AddContent(backup *model.Backup, content *model.Content)

	Count() int64
	List() BackupIterator
	GetBackupById(uint) *model.Backup
	GetBackupContentsById(uint) (*model.Backup, ContentIterator)
	GetOlderThan(string, time.Time) BackupIterator
	DeleteBackupById(uint)
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

func (r *repository) List() BackupIterator {
	sqlRows, err := r.db.Model(&model.Backup{}).Rows()
	if err != nil {
		panic(errors.Wrap(err, "Error while creating rows"))
	}

	return newBackupIterator(sqlRows, r.db)
}

func (r *repository) GetOlderThan(vault string, time time.Time) BackupIterator {
	sqlRows, err := r.db.Model(&model.Backup{}).
		Where(model.ColumnBackupVault+" = ? AND "+model.ColumnCreatedAt+" < ?", vault, time).
		Rows()

	if err != nil {
		panic(errors.Wrap(err, "Error while creating rows"))
	}

	return newBackupIterator(sqlRows, r.db)
}

func (r *repository) GetBackupById(id uint) *model.Backup {
	var backup model.Backup
	r.db.First(&backup, id)

	return &backup
}

func (r *repository) GetBackupContentsById(id uint) (*model.Backup, ContentIterator) {
	var backup model.Backup
	r.db.First(&backup, id)

	sqlRows, err := r.db.Model(&model.Content{}).Where(&model.Content{BackupID: id}).Rows()
	if err != nil {
		panic(errors.Wrap(err, "Error while creating rows"))
	}

	return &backup, newContentIterator(sqlRows, r.db)
}

func (r *repository) DeleteBackupById(id uint) {
	backup := r.GetBackupById(id)
	if backup != nil {
		r.db.Where(&model.Content{BackupID: id}).Delete(&model.Content{})
		r.db.Delete(backup)
	}
}
