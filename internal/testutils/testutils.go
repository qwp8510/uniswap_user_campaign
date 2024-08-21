package testutils

import (
	"database/sql"
	"fmt"
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

	rows, err := d.Query("SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}
	fmt.Println(tables)

	return d, nil
}
