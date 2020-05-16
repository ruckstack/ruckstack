package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/logs"
	"github.com/spf13/cobra"
)

var logsServiceCmd = &cobra.Command{
	Use:   "service [service]",
	Short: "Display logs for all containers in a service",
	Args:  cobra.ExactValidArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		logs.ShowServiceLogs(systemService, args[0], watchLogs, logsSince, logsNode)
	},
}

var logsNode string

func init() {

	logsServiceCmd.Flags().BoolVar(&systemService, "system-service", false, "Set this flag if the service is a system service")
	logsServiceCmd.Flags().BoolVar(&watchLogs, "watch", false, "Continue to output log messages")
	logsServiceCmd.Flags().StringVar(&logsSince, "since", "24h", "Oldest logs to show. Specify as a number and unit, such as 15m or 3h. Defaults to 24h. To list all logs, specify 'all'")
	logsServiceCmd.Flags().StringVar(&logsNode, "node", "all", "Show only containers on the given node. To list logs across all nodes, specify 'all'")

	logsCmd.AddCommand(logsServiceCmd)
}
