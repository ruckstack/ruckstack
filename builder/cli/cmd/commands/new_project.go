package commands

import (
	"github.com/ruckstack/ruckstack/builder/cli/internal/new_project"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
)

func init() {
	var newProjectType string
	var newProjectOut string

	var cmd = &cobra.Command{
		Use:   "new-project",
		Short: "Sets up a new project config in the current directory",
		Long:  "Generates a starting setup for your Ruckstack project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if newProjectType == "" {
				newProjectType = "example"
			}

			return new_project.NewProject(newProjectOut, newProjectType)
		},
	}

	cmd.Flags().StringVar(&newProjectType, "type", "empty", "Type of project to create. Possible values: empty or example")
	cmd.Flags().StringVar(&newProjectOut, "out", "", "Directory to create project in")

	ui.MarkFlagsRequired(cmd, "out")
	ui.MarkFlagsDirname(cmd, "out")

	rootCmd.AddCommand(cmd)

}