package cmd

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/newproject"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"github.com/spf13/cobra"
)

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

func init() {
	newProjectCmd.Flags().StringVar(&newProjectType, "type", "", "Type of project to create. Possible value: `starter` (default) or `example`")
	newProjectCmd.Flags().StringVar(&newProjectOut, "out", "", "Directory to create project in (required)")

	util.Check(newProjectCmd.MarkFlagFilename("out"))
	util.Check(newProjectCmd.MarkFlagRequired("out"))

	rootCmd.AddCommand(newProjectCmd)

}
