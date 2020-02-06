package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/server"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use: "start",
	Annotations: map[string]string{
		ANNOTATION_REQUIRES_ROOT: "true",
	},
	Run: func(cmd *cobra.Command, args []string) {
		server.Start()
	},
}

func init() {
	packageConfig := util.GetPackageConfig()

	startCmd.Short = "Starts " + packageConfig.Name

	rootCmd.AddCommand(startCmd)
}
