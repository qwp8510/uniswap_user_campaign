package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func SetupDB() (*sql.DB, error) {
	POSTGRES_HOST := os.Getenv("POSTGRES_HOST")
	POSTGRES_PORT := os.Getenv("POSTGRES_PORT")
	POSTGRES_DB := os.Getenv("POSTGRES_DB")
	POSTGRES_USER := os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD := os.Getenv("POSTGRES_PASSWORD")

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB,
	)

	d, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	d.SetConnMaxLifetime(time.Minute * 3)
	d.SetConnMaxIdleTime(15 * time.Minute)
	d.SetMaxOpenConns(10)
	d.SetMaxIdleConns(10)
	err = d.Ping()

	return d, err
}

func Downgrade(c *sql.DB, path string) error {
	driver, err := postgres.WithInstance(c, &postgres.Config{})
	if err != nil {
		return err
	}
	mi, err := migrate.NewWithDatabaseInstance("file://"+path, "mysql", driver)
	if err != nil {
		return err
	}
	err = mi.Down()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}

func Upgrade(c *sql.DB, path string) error {
	driver, err := postgres.WithInstance(c, &postgres.Config{})
	if err != nil {
		return err
	}
	mi, err := migrate.NewWithDatabaseInstance("file://"+path, "postgres", driver)
	if err != nil {
		return err
	}
	err = mi.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}
