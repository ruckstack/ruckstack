package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/config"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server"
	"github.com/spf13/cobra"
)

var (
	installOptions server.InstallOptions
)

func init() {
	var cmd = &cobra.Command{
		Use: "setup",
		//Annotations: map[string]string{
		//	RequiresRoot: "true",
		//},
		Short: "Setup " + config.PackageConfig.Name,
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Setup(installOptions)
		},
	}

	rootCmd.AddCommand(cmd)
}
