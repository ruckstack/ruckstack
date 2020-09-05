package newproject

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/resources"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func NewProject(out string, projectType string) error {
	baseAssetPath := "internal/ruckstack/builder/resources/new-project/" + projectType

	foundAsset := false
	for _, assetPath := range resources.AssetNames() {
		if strings.HasPrefix(assetPath, baseAssetPath) {
			assetTargetPath := filepath.Join(out, assetPath[len(baseAssetPath):])

			asset, err := resources.Asset(assetPath)

			if err := os.MkdirAll(filepath.Dir(assetTargetPath), 0755); err != nil {
				return err
			}

			_, err = os.Stat(assetTargetPath)
			if os.IsNotExist(err) {
				err = ioutil.WriteFile(assetTargetPath, asset, 0644)
			} else {
				return fmt.Errorf("%s already exists", assetTargetPath)
			}
			if err != nil {
				return err
			}
			foundAsset = true
		}
	}
	if !foundAsset {
		return fmt.Errorf("unknown type: " + projectType)
	}

	absOut, err := filepath.Abs(out)
	if err != nil {
		return err
	}

	fmt.Printf("Created %s project in %s\n", projectType, absOut)
	fmt.Println("")
	fmt.Printf("Open %s in your favorite text editor to see the generated project file\n", absOut+"/ruckstack.conf")
	fmt.Printf("To build it, run `ruckstack build --project %s --out ruckstack-out`\n", absOut+"/ruckstack.conf")
	fmt.Println("")
	fmt.Println("Happy Stacking!")
	fmt.Println("")

	return nil
}
