package database

import (
	"testing"
)

func TestName(t *testing.T) {
	r := NewRepository("/home/rainu/.aws/backup2glacier/database.db")
	r.DeleteBackupById(7)
}
