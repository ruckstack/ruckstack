package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/daemon/internal"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/spf13/cobra"
)

func init() {

	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts " + environment.PackageConfig.Name,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !environment.IsRunningAsRoot {
				return fmt.Errorf("command %s must be run as sudo or root", cmd.Name())
			}

			return internal.Start()
		},
	})
}
