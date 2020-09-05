package commands

import (
	"github.com/ruckstack/ruckstack/internal/system_control/server"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"github.com/spf13/cobra"
)

func init() {
	packageConfig := util.GetPackageConfig()

	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts " + packageConfig.Name,
		Annotations: map[string]string{
			REQUIRES_ROOT: "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			server.Start()
			return nil
		},
	})
}
