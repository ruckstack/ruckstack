package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server"
	"github.com/spf13/cobra"
)

func init() {
	var force bool
	var cmd = &cobra.Command{
		Use: "stop",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},
		Short: "Shuts down " + environment.PackageConfig.Name,
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Stop(force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force a non-graceful stop of the server")

	rootCmd.AddCommand(cmd)
}
