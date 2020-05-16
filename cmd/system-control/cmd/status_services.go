package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/status"
	"github.com/spf13/cobra"
)

var watchServices bool
var includeSystemServices bool

var statusServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Display status of services",
	Run: func(cmd *cobra.Command, args []string) {
		status.ShowServiceStatus(includeSystemServices, watchServices)
	},
}

func init() {

	statusServicesCmd.Flags().BoolVar(&includeSystemServices, "include-system", false, "Include system-level services in output")
	statusServicesCmd.Flags().BoolVar(&watchServices, "watch", false, "Continue watching for changes to the services")

	statusCmd.AddCommand(statusServicesCmd)
}
