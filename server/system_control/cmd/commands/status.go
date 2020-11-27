package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/status"
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

	var followJobs bool
	var includeSystemJobs bool

	var statusJobsCmd = &cobra.Command{
		Use:   "jobs",
		Short: "Display status of jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return status.ShowJobStatus(includeSystemJobs, followJobs)
		},
	}

	statusJobsCmd.Flags().BoolVar(&includeSystemJobs, "include-system", false, "Include system-level jobs in output")
	statusJobsCmd.Flags().BoolVarP(&followJobs, "follow", "f", false, "Continue watching for changes to the jobs")

	parent.AddCommand(statusJobsCmd)

}

func initStatusNodes(parent *cobra.Command) {
	var follow bool

	var statusNodesCmd = &cobra.Command{
		Use:   "nodes",
		Short: "Display status of nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return status.ShowNodeStatus(follow)
		},
	}

	statusNodesCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Continue watching for changes to the nodes")

	parent.AddCommand(statusNodesCmd)

}

func initStatusServices(parent *cobra.Command) {
	var followServices bool
	var includeSystemServices bool

	var statusServicesCmd = &cobra.Command{
		Use:   "services",
		Short: "Display status of services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return status.ShowServiceStatus(includeSystemServices, followServices)
		},
	}

	statusServicesCmd.Flags().BoolVar(&includeSystemServices, "include-system", false, "Include system-level services in output")
	statusServicesCmd.Flags().BoolVarP(&followServices, "follow", "f", false, "Continue watching for changes to the services")

	parent.AddCommand(statusServicesCmd)

}
