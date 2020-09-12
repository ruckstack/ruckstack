package resources

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"os"
	"path/filepath"
)

func ResourcePath(path ...string) (string, error) {
	ruckstackHome := util.GetRuckstackHome()

	resourcesDir := filepath.Join(ruckstackHome, "resources")
	if _, err := os.Stat(resourcesDir); os.IsNotExist(err) {
		resourcesDir = filepath.Join(ruckstackHome, "dist", "resources")
	}

	resourcePath := filepath.Join(append([]string{resourcesDir}, path...)...)

	if _, err := os.Stat(resourcePath); err != nil {
		return "", err
	}

	return resourcePath, nil
}
