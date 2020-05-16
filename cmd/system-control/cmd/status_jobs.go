package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/status"
	"github.com/spf13/cobra"
)

var watchJobs bool
var includeSystemJobs bool

var statusJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Display status of jobs",
	Run: func(cmd *cobra.Command, args []string) {
		status.ShowJobStatus(includeSystemJobs, watchJobs)
	},
}

func init() {

	statusJobsCmd.Flags().BoolVar(&includeSystemJobs, "include-system", false, "Include system-level jobs in output")
	statusJobsCmd.Flags().BoolVar(&watchJobs, "watch", false, "Continue watching for changes to the jobs")

	statusCmd.AddCommand(statusJobsCmd)
}
