package cmd

import (
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Commands for interacting with the cluster as a whole",
}

func init() {
	rootCmd.AddCommand(clusterCmd)

}
