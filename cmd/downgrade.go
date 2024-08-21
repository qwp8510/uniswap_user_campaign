package cmd

import (
	"tradingAce/pkg/core/db"

	"github.com/spf13/cobra"
)

// SchedulerCmd 是此程式的Service入口點
var DownCmd = &cobra.Command{
	Run: runDowngrade,
	Use: "downgrade",
}

func runDowngrade(_ *cobra.Command, _ []string) {

	d, err := db.SetupDB()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	if err := db.Downgrade(d, "migrations"); err != nil {
		panic(err)
	}
}
