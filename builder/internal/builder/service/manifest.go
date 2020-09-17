package service

import (
	"github.com/ruckstack/ruckstack/builder/internal/builder/installer"
	"github.com/ruckstack/ruckstack/builder/internal/project"
	"github.com/ruckstack/ruckstack/common/ui"
	"path/filepath"
)

type ManifestService struct {
	ServiceConfig  *project.ManifestServiceConfig
	ProjectConfig  *project.ProjectConfig
	serviceWorkDir string
}

func (service *ManifestService) Build(app *installer.Installer) error {
	ui.Println("Service type: manifest")

	fullManifestPath := filepath.Join(filepath.Dir(service.ServiceConfig.BaseDir), service.ServiceConfig.Manifest)

	return app.AddFile(fullManifestPath, "data/server/manifests/"+service.ServiceConfig.Id+".yaml")

}
