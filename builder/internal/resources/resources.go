package resources

import (
	"github.com/ruckstack/ruckstack/common"
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
	"path/filepath"
)

var resourceRoot string

/**
Returns the full path to the given subpath of "resources" in ruckstackHome.
Returns an error if the file does not exist
*/
func ResourcePath(path string) (string, error) {
	resourcePath := filepath.Join(GetResourceRoot(), path)

	if _, err := os.Stat(resourcePath); err != nil {
		return "", err
	}

	return resourcePath, nil
}

func GetResourceRoot() string {
	if resourceRoot != "" {
		return resourceRoot
	}

	ruckstackHome := common.GetRuckstackHome()

	resourceRoot = ruckstackHome + "/resources"
	if _, err := os.Stat(resourceRoot); os.IsNotExist(err) {
		resourceRoot = ruckstackHome + "/builder/home/resources"
	}

	ui.VPrintf("Ruckstack resource root: %s", resourceRoot)

	return resourceRoot
}
