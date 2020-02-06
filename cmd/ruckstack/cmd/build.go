package cmd

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"github.com/spf13/cobra"
)

var (
	project string
	out     string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds project",
	Long:  `Builds your Ruckstack project into an installable archive`,
	Run: func(cmd *cobra.Command, args []string) {
		builder.Build(project, out)
	},
}

func init() {
	buildCmd.Flags().StringVar(&project, "project", "", "Project file to build (required)")
	buildCmd.Flags().StringVar(&out, "out", "", "Directory to save built artifacts to (required)")

	util.Check(buildCmd.MarkFlagFilename("project"))
	util.Check(buildCmd.MarkFlagRequired("project"))

	util.Check(buildCmd.MarkFlagDirname("out"))
	util.Check(buildCmd.MarkFlagRequired("out"))

	rootCmd.AddCommand(buildCmd)

}
