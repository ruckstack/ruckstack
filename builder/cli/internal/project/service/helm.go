package service

import (
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/helm"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
	"strings"
)

type HelmService struct {
	//Common fields
	Id             string `validate:"required"`
	Type           string `validate:"required,oneof=dockerfile helm manifest"`
	Port           int    `validate:"required"`
	ProjectId      string
	ProjectVersion string

	//Unique Fields
	Chart   string `validate:"required"`
	Version string `validate:"required"`
}

func (serviceConfig *HelmService) GetId() string {
	return serviceConfig.Id
}
func (serviceConfig *HelmService) SetId(id string) {
	serviceConfig.Id = id
}

func (serviceConfig *HelmService) GetType() string {
	return serviceConfig.Type
}

func (serviceConfig *HelmService) GetPort() int {
	return serviceConfig.Port
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

	splitChart := strings.Split(service.Chart, "/")
	repo := splitChart[0]
	chartName := splitChart[1]

	chartFile, err := helm.DownloadChart(repo, chartName, service.Version)
	if err != nil {
		return err
	}

	if err := installFile.AddHelmChart(chartFile, service.Id); err != nil {
		return err
	}

	return nil
}
