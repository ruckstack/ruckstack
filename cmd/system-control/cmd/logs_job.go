package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/logs"
	"github.com/spf13/cobra"
)

var logsJobCmd = &cobra.Command{
	Use:   "job [job]",
	Short: "Display logs for a job",
	Args:  cobra.ExactValidArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		logs.ShowJobLogs(systemJobs, args[0], watchLogs)
	},
}

var systemJobs bool

func init() {

	logsJobCmd.Flags().BoolVar(&systemJobs, "system-job", false, "Set this flag if the job is a system job")
	logsJobCmd.Flags().BoolVar(&watchLogs, "watch", false, "Continue to output log messages")

	logsCmd.AddCommand(logsJobCmd)
}
