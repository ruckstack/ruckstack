package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server"
	"github.com/spf13/cobra"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "server",
		Short: "Controls the server",
	}

	initServerStart(cmd)

	rootCmd.AddCommand(cmd)
}

func initServerStart(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use: "start",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},
		Short: "Starts " + environment.PackageConfig.Name,
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Start()
		},
	})
}
