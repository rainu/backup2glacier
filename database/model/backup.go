package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

const (
	ColumnID         = "id"
	ColumnCreatedAt  = "created_at"
	ColumnUpdateddAt = "updated_at"

	ColumnBackupVault       = "vault"
	ColumnBackupDescription = "description"
	ColumnBackupUploadId    = "upload_id"
	ColumnBackupArchiveId   = "archive_id"
	ColumnBackupLocation    = "location"
	ColumnBackupChecksum    = "checksum"
	ColumnBackupLength      = "length"
	ColumnBackupPassword    = "password"
	ColumnBackupError       = "error"

	ColumnContentZipPath  = "zip_path"
	ColumnContentRealPath = "real_path"
	ColumnContentLength   = "length"
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
	Path     string    `db:"path" gorm:"type:TEXT"`
	Length   int64     `db:"length"`
	ModTime  time.Time `db:"mod"`
}
