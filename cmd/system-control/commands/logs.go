package commands

import (
	"github.com/ruckstack/ruckstack/internal/system-control/logs"
	"github.com/spf13/cobra"
)

func init() {
	var logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Display logs",
	}

	initContainerLogs(logsCmd)
	initJobLogs(logsCmd)
	initServiceLogs(logsCmd)
	rootCmd.AddCommand(logsCmd)

}

func initContainerLogs(parent *cobra.Command) {
	var systemService bool
	var previousLogs bool
	var watchLogs bool
	var logsSince string

	var cmd = &cobra.Command{
		Use:   "container [container-id]",
		Short: "Display logs for a container",
		Args:  cobra.ExactValidArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			logs.ShowContainerLogs(systemService, args[0], watchLogs, previousLogs, logsSince)

			return nil
		},
	}

	cmd.Flags().BoolVar(&systemService, "system", false, "Set this flag if the container is for a system service")
	cmd.Flags().BoolVar(&previousLogs, "previous", false, "Output logs from the previously ran instance")
	cmd.Flags().BoolVar(&watchLogs, "watch", false, "Continue to output log messages")
	cmd.Flags().StringVar(&logsSince, "since", "24h", "Oldest logs to show. Specify as a number and unit, such as 15m or 3h. Defaults to 24h. To list all logs, specify 'all'")

	parent.AddCommand(cmd)
}

func initJobLogs(parent *cobra.Command) {
	var systemJobs bool
	var watchLogs bool

	var cmd = &cobra.Command{
		Use:   "job [job]",
		Short: "Display logs for a job",
		Args:  cobra.ExactValidArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			logs.ShowJobLogs(systemJobs, args[0], watchLogs)
			return nil
		},
	}

	cmd.Flags().BoolVar(&systemJobs, "system", false, "Set this flag if the job is a system job")
	cmd.Flags().BoolVar(&watchLogs, "watch", false, "Continue to output log messages")

	parent.AddCommand(cmd)
}

func initServiceLogs(parent *cobra.Command) {
	var systemService bool
	var logsSince string
	var watchLogs bool
	var logsNode string

	var cmd = &cobra.Command{
		Use:   "service [service]",
		Short: "Display logs for all containers in a service",
		Args:  cobra.ExactValidArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			logs.ShowServiceLogs(systemService, args[0], watchLogs, logsSince, logsNode)
			return nil
		},
	}

	cmd.Flags().BoolVar(&systemService, "system", false, "Set this flag if the service is a system service")
	cmd.Flags().BoolVar(&watchLogs, "watch", false, "Continue to output log messages")
	cmd.Flags().StringVar(&logsSince, "since", "24h", "Oldest logs to show. Specify as a number and unit, such as 15m or 3h. Defaults to 24h. To list all logs, specify 'all'")
	cmd.Flags().StringVar(&logsNode, "node", "all", "Show only containers on the given node. To list logs across all nodes, specify 'all'")

	parent.AddCommand(cmd)
}
