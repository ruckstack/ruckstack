package cmd

import (
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Low-Level System Commands",
}

func init() {
	rootCmd.AddCommand(systemCmd)

}
