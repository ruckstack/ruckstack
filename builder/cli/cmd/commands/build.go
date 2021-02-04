package commands

import (
	"compress/flate"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/internal/argwrapper"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
)

func init() {
	var project string
	var out string
	var compressionLevel int

	var cmd = &cobra.Command{
		Use:   "build",
		Short: "Builds project",
		Long:  `Builds your Ruckstack project into an installable archive`,

		RunE: func(cmd *cobra.Command, args []string) error {

			if out == "." && environment.IsRunningLauncher() {
				out = "/data/out"
				argwrapper.SaveOriginalValue("out", ".", []string{})
			}

			if out == "." && environment.IsRunningLauncher() {
				out = "/data/project"
				argwrapper.SaveOriginalValue("project", ".", []string{})
			}

			environment.OutDir = out
			environment.ProjectDir = project
			return builder.Build(compressionLevel)
		},
	}

	cmd.Flags().StringVar(&project, "project", ".", "Project directory")
	cmd.Flags().StringVar(&out, "out", ".", "Directory to save installer to")
	cmd.Flags().IntVar(&compressionLevel, "compression-level", flate.BestCompression, "Compression level to use. Range from 0 (no compression) to 9 (best compression)")

	ui.MarkFlagsDirname(cmd, "project")
	ui.MarkFlagsDirname(cmd, "out")

	RootCmd.AddCommand(cmd)

}
