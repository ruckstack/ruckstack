package manifest

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/installer"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
	"log"
	"path/filepath"
)

func AddService(serviceConfig *project.ManifestServiceConfig, app *installer.Installer, projectConfig *project.ProjectConfig) {
	log.Println("Service type: manifest")

	fullManifestPath := filepath.Join(filepath.Dir(serviceConfig.BaseDir), serviceConfig.Manifest)

	app.AddFile(fullManifestPath, "data/server/manifests/"+serviceConfig.Id+".yaml")

}
