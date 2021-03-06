package commands

import (
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/builder/internal/init_project"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
)

func init() {
	var newProjectTemplate string
	var newProjectOut string

	var cmd = &cobra.Command{
		Use:   "init",
		Short: "Creates a Ruckstack project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if newProjectTemplate == "" {
				newProjectTemplate = "empty"
			}

			environment.OutDir = newProjectOut

			return init_project.InitProject(newProjectTemplate)
		},
	}

	cmd.Flags().StringVar(&newProjectTemplate, "template", "empty", "Type of project to create. Possible values: empty or example")
	cmd.Flags().StringVar(&newProjectOut, "out", ".", "Directory to create project in. Defaults to current directory")

	ui.MarkFlagsDirname(cmd, "out")

	RootCmd.AddCommand(cmd)

}
