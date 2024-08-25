package cmd

import (
	"tradingAce/internal/rest"
	"tradingAce/pkg/core/db"
	"tradingAce/pkg/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Run: runServer,
	Use: "server",
}

func runServer(_ *cobra.Command, _ []string) {
	d, err := db.SetupDB()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	s := service.NewService(d)
	server := rest.NewRestServer(s.Task, s.UserPoint, s.UserTask)

	r := gin.Default()
	r.GET("/userTasks/:address", server.GetUserTasks)
	r.GET("/userPoints/*taskId", server.GetUserPoints)
	r.POST("/sharePoolTask", server.CreateSharePoolTask)

	r.Run(":8080")
}
