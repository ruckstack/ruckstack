package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Checks system status",
}

func init() {
	rootCmd.AddCommand(statusCmd)

}
