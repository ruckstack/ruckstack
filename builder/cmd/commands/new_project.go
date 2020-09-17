package commands

import (
	"github.com/ruckstack/ruckstack/builder/internal/new_project"
	"github.com/ruckstack/ruckstack/builder/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	var newProjectType string
	var newProjectOut string

	var newProjectCmd = &cobra.Command{
		Use:   "new-project",
		Short: "Sets up a new project config in the current directory",
		Long:  `Generates a starting setup for your Ruckstack project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if newProjectType == "" {
				newProjectType = "example"
			}

			return new_project.NewProject(newProjectOut, newProjectType)
		},
	}

	newProjectCmd.Flags().StringVar(&newProjectType, "type", "empty", "Type of project to create. Possible values: empty or example")
	newProjectCmd.Flags().StringVar(&newProjectOut, "out", "", "Directory to create project in (required)")

	util.ExpectNoError(newProjectCmd.MarkFlagFilename("out"))
	util.ExpectNoError(newProjectCmd.MarkFlagRequired("out"))

	rootCmd.AddCommand(newProjectCmd)

}
