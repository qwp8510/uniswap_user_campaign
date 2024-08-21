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

	if err := db.Upgrade(d, "migrations"); err != nil {
		panic(err)
	}

	s := service.NewService(d)

	taskListener := listener.SwapEventTask{
		TaskMgr:        s.Task,
		TransactionMgr: s.Transaction,
		UserTaskMgr:    s.UserTask,
	}

	taskListener.Listen()
}
