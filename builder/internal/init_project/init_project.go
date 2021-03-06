package init_project

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/bundled"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/builder/internal/util"
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
	"path/filepath"
	"strings"
)

func InitProject(projectType string) error {
	commonDir, _ := bundled.OpenDir("init_common")
	sourceDir, _ := bundled.OpenDir("init/" + projectType)

	ui.VPrintf("Copying template project %s to %s", sourceDir, environment.OutDir)
	if err := util.CopyDir(commonDir, environment.OutDir); err != nil {
		return err
	}
	if err := util.CopyDir(sourceDir, environment.OutDir); err != nil {
		if os.IsNotExist(err) {
			projectTypes, _ := bundled.ReadDir("init")
			projectTypeNames := make([]string, len(projectTypes))

			for i := 0; i < len(projectTypes); i++ {
				projectTypeNames[i] = projectTypes[i].Name()
			}

			return fmt.Errorf("unknown template: '%s'. Available templates: %s", projectType, strings.Join(projectTypeNames, ", "))
		} else {
			return err
		}
	}

	outputDirAbs, _ := filepath.Abs(environment.OutDir)
	ui.Printf("Created %s project in %s\n", projectType, outputDirAbs)
	ui.Println("To initialize with a different template, use the --template flag.")
	ui.Println("")
	ui.Printf("Open %s/ruckstack.yaml in your favorite text editor to see the generated project file\n", outputDirAbs)
	ui.Printf("To build it, run `ruckstack build` from %s\n", outputDirAbs)
	ui.Println("")
	ui.Println("Happy Stacking!")
	ui.Println("")

	return nil
}
