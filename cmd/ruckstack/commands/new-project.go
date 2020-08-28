package commands

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/newproject"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"github.com/spf13/cobra"
)

func init() {
	var newProjectType string
	var newProjectOut string

	var newProjectCmd = &cobra.Command{
		Use:   "new-project",
		Short: "Sets up a new project config in the current directory",
		Long:  `Generates a starting setup for your Ruckstack project`,
		Run: func(cmd *cobra.Command, args []string) {
			if newProjectType == "" {
				newProjectType = "example"
			}

			newproject.NewProject(newProjectOut, newProjectType)
		},
	}

	newProjectCmd.Flags().StringVar(&newProjectType, "type", "", "Type of project to create. Possible value: `starter` (default) or `example`")
	newProjectCmd.Flags().StringVar(&newProjectOut, "out", "", "Directory to create project in (required)")

	util.ExpectNoError(newProjectCmd.MarkFlagFilename("out"))
	util.ExpectNoError(newProjectCmd.MarkFlagRequired("out"))

	rootCmd.AddCommand(newProjectCmd)

}
