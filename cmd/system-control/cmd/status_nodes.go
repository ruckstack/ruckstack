package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/status"
	"github.com/spf13/cobra"
)

var watch bool

var statusNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Display status of nodes",
	Run: func(cmd *cobra.Command, args []string) {
		status.ShowNodeStatus(watch)
	},
}

func init() {

	statusNodesCmd.Flags().BoolVar(&watch, "watch", false, "Continue watching for changes to the nodes")

	statusCmd.AddCommand(statusNodesCmd)
}
