package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/config"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server"
	"github.com/spf13/cobra"
)

func init() {
	var cmd = &cobra.Command{
		Use: "stop",
		//Annotations: map[string]string{
		//	RequiresRoot: "true",
		//},
		Short: "Stop " + config.PackageConfig.Name,
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Stop()
		},
	}

	rootCmd.AddCommand(cmd)
}
