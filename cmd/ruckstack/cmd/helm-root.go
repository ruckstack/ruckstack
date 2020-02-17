package cmd

import (
	"github.com/spf13/cobra"
)

var helmRootCmd = &cobra.Command{
	Use:   "helm",
	Short: "Commands for interacting with the Helm repository",
	//Run: func(cmd *cobra.Command, args []string) {
	//	builder.Build(project, out)
	//fmt.Println("Helm!")
	//helm.GetCurrentVersion()
	//},
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

	rootCmd.AddCommand(helmRootCmd)

}
