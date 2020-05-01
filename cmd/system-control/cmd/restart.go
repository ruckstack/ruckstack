package cmd

import (
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restarts parts of the system",
}

func init() {
	rootCmd.AddCommand(restartCmd)

}
