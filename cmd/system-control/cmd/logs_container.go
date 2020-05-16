package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/logs"
	"github.com/spf13/cobra"
)

var (
	previousLogs bool
	watchLogs    bool
	logsSince    string
)

var logsContainerCmd = &cobra.Command{
	Use:   "container [container-id]",
	Short: "Display logs for a container",
	Args:  cobra.ExactValidArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		logs.ShowContainerLogs(systemService, args[0], watchLogs, previousLogs, logsSince)
	},
}

func init() {
	logsContainerCmd.Flags().BoolVar(&systemService, "system-container", false, "Set this flag if the container is for a system service")
	logsContainerCmd.Flags().BoolVar(&previousLogs, "previous", false, "Output logs from the previously ran instance")
	logsContainerCmd.Flags().BoolVar(&watchLogs, "watch", false, "Continue to output log messages")
	logsContainerCmd.Flags().StringVar(&logsSince, "since", "24h", "Oldest logs to show. Specify as a number and unit, such as 15m or 3h. Defaults to 24h. To list all logs, specify 'all'")

	logsCmd.AddCommand(logsContainerCmd)
}
