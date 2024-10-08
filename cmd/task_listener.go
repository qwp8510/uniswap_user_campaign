package cmd

import (
	"tradingAce/internal/listener"
	"tradingAce/pkg/core/db"
	"tradingAce/pkg/service"

	"github.com/spf13/cobra"
)

// SchedulerCmd 是此程式的Service入口點
var TaskListenerCmd = &cobra.Command{
	Run: runTaskListener,
	Use: "taskListener",
}

func runTaskListener(_ *cobra.Command, _ []string) {

	d, err := db.SetupDB()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	s := service.NewService(d)

	taskListener := listener.NewTaskListener(s.Task, s.Transaction, s.UserTask)
	taskListener.Listen()
}
