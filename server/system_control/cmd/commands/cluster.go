package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/cluster"
	"github.com/spf13/cobra"
)

func init() {
	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "Commands for interacting with the cluster as a whole",
	}
	initAddNode(clusterCmd)

	rootCmd.AddCommand(clusterCmd)

}

func initAddNode(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use: "add-node",
		Annotations: map[string]string{
			REQUIRES_ROOT: "true",
		},
		Short: "Adds a node to the cluster",
		Long:  `Used during installation of additional server nodes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cluster.AddNode()
		},
	})

}
