package main

import (
	"fmt"
	"os"
	"tradingAce/cmd"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "server cli"}

func main() {
	godotenv.Load(".env/.env")

	rootCmd.AddCommand(cmd.MigrateCmd, cmd.TaskListenerCmd, cmd.DownCmd, cmd.ServerCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
