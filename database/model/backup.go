package model

import (
	"github.com/jinzhu/gorm"
)

type Backup struct {
	gorm.Model

	Vault       string    `db:"vault"`
	Description string    `db:"description" gorm:"type:TEXT"`
	UploadId    *string   `db:"upload_id"`
	ArchiveId   *string   `db:"archive_id"`
	Location    *string   `db:"location"`
	Checksum    *string   `db:"checksum"`
	Length      int64     `db:"length"`
	Password    string    `db:"password"`
	Error       string    `db:"error"`
	FileList    []Content `gorm:"foreignkey:BackupID"`
}

type Content struct {
	ID       uint `gorm:"primary_key"`
	BackupID uint
	ZipPath  string `db:"zip_path" gorm:"type:TEXT"`
	RealPath string `db:"real_path" gorm:"type:TEXT"`
	Length   int64  `db:"length"`
}
