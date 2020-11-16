package project

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/project/service"
)

type Project struct {
	Id      string `validate:"required"`
	Name    string `validate:"required"`
	Version string `validate:"required"`

	HelmVersion       string
	K3sVersion        string
	SystemControlName string

	ManifestServices   []service.ManifestService   `yaml:"manifestServices"`
	HelmServices       []service.HelmService       `yaml:"helmServices"`
	DockerfileServices []service.DockerfileService `yaml:"dockerfileServices"`
}

func (project *Project) GetServices() []Service {
	returnList := []Service{}

	for _, item := range project.ManifestServices {
		returnList = append(returnList, &item)
	}
	for _, item := range project.HelmServices {
		returnList = append(returnList, &item)
	}

	for _, item := range project.DockerfileServices {
		returnList = append(returnList, &item)
	}

	return returnList
}

func (project *Project) Validate() error {
	structValidator := validator.New()

	if err := structValidator.Struct(project); err != nil {
		return fmt.Errorf("error parsing project file: %s", err)
	}

	if len(project.GetServices()) == 0 {
		return fmt.Errorf("error parsing project file: at least one service block is required")
	}
	for _, serviceConfig := range project.GetServices() {
		if err := serviceConfig.Validate(structValidator); err != nil {
			return fmt.Errorf("error parsing service %s: %s", serviceConfig.GetId(), err)
		}
	}

	return nil
}

type Service interface {
	GetId() string
	SetId(id string)

	GetType() string
	GetPort() int

	SetProjectId(string)
	SetProjectVersion(string)

	/**
	Validate that the service is configured correctly
	*/
	Validate(*validator.Validate) error

	/**
	Build the service, adding anything needed to the InstallFile
	*/
	Build(*install_file.InstallFile) error
}
