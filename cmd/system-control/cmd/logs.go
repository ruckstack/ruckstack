package cmd

import (
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Display logs",
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
