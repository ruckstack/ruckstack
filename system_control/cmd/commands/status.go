package commands

import (
	"github.com/ruckstack/ruckstack/system_control/internal/status"
	"github.com/spf13/cobra"
)

func init() {
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Checks system status",
	}

	initStatusJobs(statusCmd)
	initStatusNodes(statusCmd)
	initStatusServices(statusCmd)

	rootCmd.AddCommand(statusCmd)
}

func initStatusJobs(parent *cobra.Command) {

	var watchJobs bool
	var includeSystemJobs bool

	var statusJobsCmd = &cobra.Command{
		Use:   "jobs",
		Short: "Display status of jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return status.ShowJobStatus(includeSystemJobs, watchJobs)
		},
	}

	statusJobsCmd.Flags().BoolVar(&includeSystemJobs, "include-system", false, "Include system-level jobs in output")
	statusJobsCmd.Flags().BoolVar(&watchJobs, "watch", false, "Continue watching for changes to the jobs")

	parent.AddCommand(statusJobsCmd)

}

func initStatusNodes(parent *cobra.Command) {
	var watch bool

	var statusNodesCmd = &cobra.Command{
		Use:   "nodes",
		Short: "Display status of nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return status.ShowNodeStatus(watch)
		},
	}

	statusNodesCmd.Flags().BoolVar(&watch, "watch", false, "Continue watching for changes to the nodes")

	parent.AddCommand(statusNodesCmd)

}

func initStatusServices(parent *cobra.Command) {
	var watchServices bool
	var includeSystemServices bool

	var statusServicesCmd = &cobra.Command{
		Use:   "services",
		Short: "Display status of services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return status.ShowServiceStatus(includeSystemServices, watchServices)
		},
	}

	statusServicesCmd.Flags().BoolVar(&includeSystemServices, "include-system", false, "Include system-level services in output")
	statusServicesCmd.Flags().BoolVar(&watchServices, "watch", false, "Continue watching for changes to the services")

	parent.AddCommand(statusServicesCmd)

}
