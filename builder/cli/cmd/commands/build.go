package commands

import (
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
)

func init() {
	var project string
	var out string

	var cmd = &cobra.Command{
		Use:   "build",
		Short: "Builds project",
		Long:  `Builds your Ruckstack project into an installable archive`,

		RunE: func(cmd *cobra.Command, args []string) error {
			environment.OutDir = out
			environment.ProjectDir = project
			return builder.Build()
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project file to build")
	cmd.Flags().StringVar(&out, "out", "", "Directory to save built artifacts to")

	ui.MarkFlagsRequired(cmd, "project", "out")
	ui.MarkFlagsFilename(cmd, "project")
	ui.MarkFlagsDirname(cmd, "out")

	rootCmd.AddCommand(cmd)

}
