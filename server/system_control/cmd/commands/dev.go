package commands

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/dev"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/spf13/cobra"
)

func init() {
	var devCmd = &cobra.Command{
		Use:   "dev",
		Short: "Enable Development Mode",
	}

	if environment.ClusterConfig.DevModeEnabled {
		initDevModeDisable(devCmd)
		initDevModeReroute(devCmd)
		initDevModeShowRoutes(devCmd)
		initDevModeRemoveRoute(devCmd)
	} else {
		initDevModeEnable(devCmd)
	}

	rootCmd.AddCommand(devCmd)

}

func initDevModeEnable(parent *cobra.Command) {
	var cmd = &cobra.Command{
		Use:   "enable",
		Short: "Enables 'development mode'",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.Enable()
		},
	}

	parent.AddCommand(cmd)
}

func initDevModeDisable(parent *cobra.Command) {
	var cmd = &cobra.Command{
		Use:   "disable",
		Short: "Disables 'development mode'",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.Disable()
		},
	}

	parent.AddCommand(cmd)
}

func initDevModeReroute(parent *cobra.Command) {
	var serviceName string
	var targetHost string
	var targetPort int

	var cmd = &cobra.Command{
		Use:   "reroute",
		Short: "Routes traffic to the given external system rather than the deployed service",
		Long:  "Useful for running particular services in your development/debugging environment",

		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.Reroute(serviceName, targetHost, targetPort)
		},
	}

	cmd.Flags().StringVar(&serviceName, "service", "", "Service to replace")
	cmd.Flags().StringVar(&targetHost, "targetHost", "localhost", "Host to proxy requests to")
	cmd.Flags().IntVar(&targetPort, "targetPort", 80, "Port to proxy requests to")

	ui.MarkFlagsRequired(cmd, "service")

	parent.AddCommand(cmd)
}

func initDevModeRemoveRoute(parent *cobra.Command) {
	var serviceName string

	var cmd = &cobra.Command{
		Use:   "remove-route",
		Short: "Removes a previously-configured reroute",

		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.RemoveRoute(serviceName)
		},
	}

	cmd.Flags().StringVar(&serviceName, "service", "", "Service to remove reroute")

	ui.MarkFlagsRequired(cmd, "service")

	parent.AddCommand(cmd)
}

func initDevModeShowRoutes(parent *cobra.Command) {

	var cmd = &cobra.Command{
		Use:   "show-routes",
		Short: "Shows dev reroute rules",

		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.ShowRoutes()
		},
	}

	parent.AddCommand(cmd)
}
