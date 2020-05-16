package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/restart"
	"github.com/spf13/cobra"
)

var systemService bool

var restartServiceCmd = &cobra.Command{
	Use:   "service [service-id]",
	Short: "Restarts all containers in a service",
	Args:  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		restart.Service(systemService, args[0])
	},
}

func init() {
	restartServiceCmd.Flags().BoolVar(&systemService, "system", false, "Set this flag if the service is a system service")

	restartCmd.AddCommand(restartServiceCmd)
}
