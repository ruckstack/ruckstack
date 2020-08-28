package commands

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "ruckstack",
	Short:   "Ruckstack CLI",
	Long:    `Ruckstack CLI`,
	Version: "0.8.3",
}

func init() {
}

func Execute(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
