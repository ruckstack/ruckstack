package commands

import (
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder"
	"github.com/ruckstack/ruckstack/builder/cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	var project string
	var out string

	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Builds project",
		Long:  `Builds your Ruckstack project into an installable archive`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return builder.Build(project, out)
		},
	}

	buildCmd.Flags().StringVar(&project, "project", "", "Project file to build (required)")
	buildCmd.Flags().StringVar(&out, "out", "", "Directory to save built artifacts to (required)")

	util.ExpectNoError(buildCmd.MarkFlagFilename("project"))
	util.ExpectNoError(buildCmd.MarkFlagRequired("project"))

	util.ExpectNoError(buildCmd.MarkFlagDirname("out"))
	util.ExpectNoError(buildCmd.MarkFlagRequired("out"))

	rootCmd.AddCommand(buildCmd)

}
