package db

import (
	"testing"

	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func Test_SetupDB(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := SetupDB()
	if err != nil {
		t.Errorf("setup db err: %v", err)
	}
	defer d.Close()
}

func Test_Migrate(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := SetupDB()
	if err != nil {
		t.Errorf("setup db err: %v", err)
	}
	defer d.Close()

	if err := Upgrade(d, "../../../migrations"); err != nil {
		t.Errorf("Upgrade err: %v", err)
	}

	if err := Downgrade(d, "../../../migrations"); err != nil {
		t.Errorf("Downgrade err: %v", err)
	}
}
