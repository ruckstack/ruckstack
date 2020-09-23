package project

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
)

type Project struct {
	Id      string `validate:"required"`
	Name    string `validate:"required"`
	Version string `validate:"required"`

	HelmVersion       string `ini:"helm_version"`
	K3sVersion        string `ini:"k3s_version"`
	SystemControlName string `ini:"system_control_name"`

	Services []Service
}

func (project *Project) Validate() error {
	structValidator := validator.New()

	if err := structValidator.Struct(project); err != nil {
		return fmt.Errorf("error parsing project file: %s", err)
	}

	if len(project.Services) == 0 {
		return fmt.Errorf("error parsing project file: at least one service block is required")
	}
	for _, service := range project.Services {
		if err := service.Validate(structValidator); err != nil {
			return fmt.Errorf("error parsing service %s: %s", service.GetId(), err)
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
