package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server"
	"github.com/spf13/cobra"
)

func init() {

	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts " + environment.PackageConfig.Name,
		Annotations: map[string]string{
			RequiresRoot: "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Start()
		},
	})
}
