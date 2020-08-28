package commands

import (
	"github.com/ruckstack/ruckstack/internal/system-control/restart"
	"github.com/spf13/cobra"
)

func init() {
	var restartCmd = &cobra.Command{
		Use:   "restart",
		Short: "Restarts parts of the system",
	}

	initRestartContainer(restartCmd)
	initServiceLogs(restartCmd)

	rootCmd.AddCommand(restartCmd)
}

func initRestartContainer(parent *cobra.Command) {
	var systemContainer bool

	var cmd = &cobra.Command{
		Use:   "container [container-id]",
		Short: "Restart a container",
		Args:  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restart.Container(systemContainer, args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&systemContainer, "system", false, "Set this flag if the container is a system container")

	parent.AddCommand(cmd)
}

func initRestartService(parent *cobra.Command) {
	var systemService bool

	var cmd = &cobra.Command{
		Use:   "service [service-id]",
		Short: "Restarts all containers in a service",
		Args:  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restart.Service(systemService, args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&systemService, "system", false, "Set this flag if the service is a system service")

	parent.AddCommand(cmd)

}
