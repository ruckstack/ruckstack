package commands

import (
	"github.com/ruckstack/ruckstack/internal/system-control/uninstall"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall from this machine",
		Annotations: map[string]string{
			REQUIRES_ROOT: "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			uninstall.Uninstall()
			return nil
		},
	})
}
