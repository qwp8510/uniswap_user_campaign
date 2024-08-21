package cmd

import (
	"database/sql"
	"os"
	"time"
	"tradingAce/pkg/core/db"
	"tradingAce/pkg/utils"

	"github.com/spf13/cobra"
)

// SchedulerCmd 是此程式的Service入口點
var MigrateCmd = &cobra.Command{
	Run: runMigrate,
	Use: "migrate",
}

var (
	sqlPath string
)

func runMigrate(_ *cobra.Command, _ []string) {

	d, err := db.SetupDB()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	if err := db.Upgrade(d, "migrations"); err != nil {
		panic(err)
	}

	if err := initFirstTask(d); err != nil {
		panic(err)
	}
}

func initFirstTask(d *sql.DB) error {
	startAt, parseErr := time.Parse("2006-01-02", os.Getenv("FIRST_TASK_START"))
	if parseErr != nil {
		return parseErr
	}

	if _, err := d.Exec(
		`INSERT INTO task("id", "createdAt", "name", "pairAddress", "startAt")
		SELECT $1, $2, $3, $4, $5
		WHERE NOT EXISTS (SELECT 1 FROM task WHERE name = 'onboarding');`,
		utils.GenDBID(), time.Now(), "onboarding", nil, startAt,
	); err != nil {
		return err
	}

	_, err := d.Exec(
		`INSERT INTO task("id", "createdAt", "name", "pairAddress", "startAt") VALUES ($1, $2, $3, $4, $5) ON CONFLICT ("pairAddress") DO NOTHING;`,
		utils.GenDBID(), time.Now(), "share_pool", "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc", startAt,
	)

	return err
}
