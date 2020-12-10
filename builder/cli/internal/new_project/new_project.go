package new_project

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/cli/internal/util"
	"github.com/ruckstack/ruckstack/builder/internal/argwrapper"
	"github.com/ruckstack/ruckstack/common/ui"
	"io/ioutil"
	"os"
	"strings"
)

func NewProject(projectType string) error {
	commonDir, err := environment.ResourcePath("new_project/common")
	sourceDir, err := environment.ResourcePath("new_project/" + projectType)
	if err != nil {
		if os.IsNotExist(err) {
			newProjectDir, _ := environment.ResourcePath("new_project")
			projectTypes, _ := ioutil.ReadDir(newProjectDir)
			projectTypeNames := make([]string, len(projectTypes))

			for i := 0; i < len(projectTypes); i++ {
				projectTypeNames[i] = projectTypes[i].Name()
			}

			return fmt.Errorf("unknown template: '%s'. Available templates: %s", projectType, strings.Join(projectTypeNames, ", "))
		} else {
			return err
		}
	}

	ui.VPrintf("Copying template project %s to %s", sourceDir, environment.OutDir)
	if err := util.CopyDir(commonDir, environment.OutDir); err != nil {
		return err
	}
	if err := util.CopyDir(sourceDir, environment.OutDir); err != nil {
		return err
	}

	outputDirToShow := argwrapper.GetOriginalValue("out", environment.OutDir)

	ui.Printf("Created %s project in %s\n", projectType, outputDirToShow)
	ui.Println("To create a project with a different template, use the --type flag.")
	ui.Println("")
	ui.Printf("Open %s/ruckstack.yaml in your favorite text editor to see the generated project file\n", outputDirToShow)
	ui.Printf("To build it, run `ruckstack build` from %s\n", outputDirToShow)
	ui.Println("")
	ui.Println("Happy Stacking!")
	ui.Println("")

	return nil
}
