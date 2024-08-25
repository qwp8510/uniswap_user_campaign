package testutils

import (
	"database/sql"
	"testing"
	"tradingAce/pkg/core/db"
)

func GetTestDb(_ *testing.T, migrationPath string) (*sql.DB, error) {
	d, err := db.SetupDB()
	if err != nil {
		return d, err
	}

	if err := db.Downgrade(d, migrationPath); err != nil {
		return d, err
	}
	if err := db.Upgrade(d, migrationPath); err != nil {
		return d, err
	}

	return d, nil
}
