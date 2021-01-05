package service

import (
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/cli/internal/helm"
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
)

type HelmService struct {
	//Common fields
	Id             string `validate:"required"`
	ProjectId      string
	ProjectVersion string

	//Unique Fields
	Chart   string `validate:"required"`
	Version string `validate:"required"`
}

var defaultRepoUrl = "https://charts.helm.sh/stable"

func (serviceConfig *HelmService) GetId() string {
	return serviceConfig.Id
}
func (serviceConfig *HelmService) SetId(id string) {
	serviceConfig.Id = id
}

func (serviceConfig *HelmService) GetType() string {
	return "helm"
}

func (serviceConfig *HelmService) SetProjectId(projectId string) {
	serviceConfig.ProjectId = projectId
}

func (serviceConfig *HelmService) SetProjectVersion(projectVersion string) {
	serviceConfig.ProjectVersion = projectVersion
}

func (service *HelmService) Validate(structValidator *validator.Validate) error {
	if err := structValidator.Struct(service); err != nil {
		return err
	}
	return nil
}

func (service *HelmService) Build(installFile *install_file.InstallFile) error {
	ui.Printf("Building Helm Service %s", service.Id)

	serviceWorkDir := environment.TempPath(service.Id + "-*")
	if err := os.MkdirAll(serviceWorkDir, 0755); err != nil {
		return err
	}

	chartFile, err := helm.DownloadChart(service.Chart, service.Version)
	if err != nil {
		return err
	}

	if err := installFile.AddHelmChart(chartFile, service.Id); err != nil {
		return err
	}

	return nil
}
