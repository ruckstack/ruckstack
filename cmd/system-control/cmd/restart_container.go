package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/restart"
	"github.com/spf13/cobra"
)

var systemContainer bool

var restartContainerCmd = &cobra.Command{
	Use:   "container [container-id]",
	Short: "Restart a container",
	Args:  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		restart.Container(systemContainer, args[0])
	},
}

func init() {
	restartContainerCmd.Flags().BoolVar(&systemContainer, "system", false, "Set this flag if the container is a system container")

	restartCmd.AddCommand(restartContainerCmd)
}
