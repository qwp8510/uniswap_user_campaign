package cmd

import (
	"context"
	"log"
	"tradingAce/pkg/core/db"
	"tradingAce/pkg/service"

	"github.com/spf13/cobra"
)

// SchedulerCmd 是此程式的Service入口點
var CheckSharePoolTaskCmd = &cobra.Command{
	Run: runCheckSharePoolTaskCmd,
	Use: "checkSharePoolTask",
}

func runCheckSharePoolTaskCmd(_ *cobra.Command, _ []string) {
	d, err := db.SetupDB()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	if err := db.Upgrade(d, "migrations"); err != nil {
		panic(err)
	}

	s := service.NewService(d)
	if err := s.UserTask.CheckSharePoolTasks(context.TODO()); err != nil {
		log.Panicln(err)
	}
}
