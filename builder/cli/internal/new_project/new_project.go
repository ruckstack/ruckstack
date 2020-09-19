package new_project

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/cli/internal/util"
	"github.com/ruckstack/ruckstack/common/ui"
	"io/ioutil"
	"os"
	"strings"
)

func NewProject(outputDirectory string, projectType string) error {
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

	if err := util.CopyDir(sourceDir, outputDirectory); err != nil {
		return err
	}

	outputDirToShow := util.WrappedValue("out_abs", outputDirectory)

	ui.Printf("Created %s project in %s\n", projectType, outputDirToShow)
	ui.Println("")
	ui.Printf("Open %s/ruckstack.conf in your favorite text editor to see the generated project file\n", outputDirToShow)
	ui.Printf("To build it, run `ruckstack build --project %s/ruckstack.conf --out ruckstack-out`\n", outputDirToShow)
	ui.Println("")
	ui.Println("Happy Stacking!")
	ui.Println("")

	return nil
}
