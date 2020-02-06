package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/uninstall"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall from this machine",
	Annotations: map[string]string{
		ANNOTATION_REQUIRES_ROOT: "true",
	},
	Run: func(cmd *cobra.Command, args []string) {
		uninstall.Uninstall()
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
