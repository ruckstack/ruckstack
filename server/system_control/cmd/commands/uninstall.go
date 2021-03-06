package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/uninstall"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall from this machine",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstall.Uninstall()
		},
	})
}
