package manifest

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/artifact"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/shared"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
	"log"
	"path/filepath"
)

func AddService(serviceConfig *project.ManifestServiceConfig, app *artifact.Artifact, projectConfig *project.ProjectConfig, buildEnv *shared.BuildEnvironment) {
	log.Println("Service type: manifest")

	fullManifestPath := filepath.Join(filepath.Dir(serviceConfig.BaseDir), serviceConfig.Manifest)

	app.AddFile(fullManifestPath, "data/server/manifests/"+serviceConfig.Id+".yaml")

}
