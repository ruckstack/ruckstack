package service

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"io/ioutil"
	"path/filepath"
)

type ManifestService struct {
	//Common fields
	Id             string `validate:"required"`
	ProjectId      string
	ProjectVersion string

	//Unique Fields
	Manifest string `validate:"required"`
}

func (serviceConfig *ManifestService) GetId() string {
	return serviceConfig.Id
}

func (serviceConfig *ManifestService) SetId(id string) {
	serviceConfig.Id = id
}

func (serviceConfig *ManifestService) GetType() string {
	return "manifest"
}

func (serviceConfig *ManifestService) SetProjectId(projectId string) {
	serviceConfig.ProjectId = projectId
}

func (serviceConfig *ManifestService) SetProjectVersion(projectVersion string) {
	serviceConfig.ProjectVersion = projectVersion
}

func (service *ManifestService) Validate(structValidator *validator.Validate) error {
	if err := structValidator.Struct(service); err != nil {
		return err
	}

	if filepath.IsAbs(service.Manifest) {
		return fmt.Errorf("manifest paths must be relative to the project root")
	}

	return nil
}

func (service *ManifestService) Build(installFile *install_file.InstallFile) error {
	ui.Printf("Building Manifest Service %s", service.Id)

	fullManifestPath := filepath.Join(environment.ProjectDir, service.Manifest)
	fullManifestContent, err := ioutil.ReadFile(fullManifestPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %s", fullManifestPath, err)
	}

	if len(fullManifestContent) == 0 {
		return fmt.Errorf("empty manifest file %s", service.Manifest)
	}

	if err := installFile.AddFile(fullManifestPath, "data/server/manifests/"+service.Id+".yaml"); err != nil {
		return fmt.Errorf("error adding %s to installer: %s", fullManifestPath, err)
	}

	if err := installFile.AddImagesInManifest(fullManifestContent); err != nil {
		return fmt.Errorf("error parsing manifest %s: %s", service.Manifest, err)
	}

	return nil
}
