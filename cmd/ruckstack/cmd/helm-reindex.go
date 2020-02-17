package cmd

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/helm"
	"github.com/spf13/cobra"
)

var helmReindexCmd = &cobra.Command{
	Use:   "re-index",
	Short: "Refreshes list of available helm charts",
	Run: func(cmd *cobra.Command, args []string) {
		helm.Setup()
		helm.ReIndex()
	},
}

func init() {
	//buildCmd.Flags().StringVar(&project, "project", "", "Project file to build (required)")
	//buildCmd.Flags().StringVar(&out, "out", "", "Directory to save built artifacts to (required)")
	//
	//util.Check(buildCmd.MarkFlagFilename("project"))
	//util.Check(buildCmd.MarkFlagRequired("project"))
	//
	//util.Check(buildCmd.MarkFlagDirname("out"))
	//util.Check(buildCmd.MarkFlagRequired("out"))

	helmRootCmd.AddCommand(helmReindexCmd)

}
