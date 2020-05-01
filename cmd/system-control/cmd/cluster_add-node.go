package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/cluster"
	"github.com/spf13/cobra"
)

var clusterAddNodeCmd = &cobra.Command{
	Use: "add-node",
	Annotations: map[string]string{
		ANNOTATION_REQUIRES_ROOT: "true",
	},
	Short: "Adds a node to the cluster",
	Long:  `Used during installation of additional server nodes`,
	Run: func(cmd *cobra.Command, args []string) {
		cluster.AddNode()
	},
}

func init() {
	clusterCmd.AddCommand(clusterAddNodeCmd)
}
