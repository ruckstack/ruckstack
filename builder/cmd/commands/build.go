package commands

import (
	"compress/flate"
	"github.com/ruckstack/ruckstack/builder/internal/builder"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
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
