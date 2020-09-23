package project

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/project/service"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Parse(source interface{}) (*Project, error) {
	projectPath := "in-memory"
	if stringPath, ok := source.(string); ok {
		projectPath = stringPath
	}

	projectConfigFile, err := ini.InsensitiveLoad(source)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("project file %s not found", projectPath)
	}
	if err != nil {
		return nil, err
	}

	projectConfigFile.NameMapper = ini.TitleUnderscore

	projectConfig := &Project{
		K3sVersion:  "1.18.6+k3s1",
		HelmVersion: "3.2.4",
	}

	projectConfigFile.Section("ruckstack-project").MapTo(projectConfig)

	matched, _ := regexp.MatchString(`^[a-z0-9-_]+$`, projectConfig.Id)
	if !matched {
		return nil, fmt.Errorf("project id must be lower case, alphanumeric, with no whitespace")
	}

	for _, service := range projectConfigFile.Section("services").Keys() {
		err := parseServiceFile(service.Value(), projectConfig)
		if err != nil {
			return nil, err
		}
	}

	for _, section := range projectConfigFile.Sections() {
		sectionName := section.Name()
		if strings.HasPrefix(sectionName, "service-") {
			err = parseServiceSection(strings.TrimPrefix(sectionName, "service-"), section, projectConfig, projectPath)
			if err != nil {
				return nil, err
			}

		}
	}

	if len(projectConfig.Services) == 0 {
		return nil, fmt.Errorf("No services are defined in %s", projectPath)
	}

	if projectConfig.SystemControlName == "" {
		projectConfig.SystemControlName = projectConfig.Id
	}

	if err := projectConfig.Validate(); err != nil {
		return nil, err
	}

	return projectConfig, nil

}

func parseServiceFile(serviceConfigPath string, projectConfig *Project) error {
	serviceConfigFile, err := ini.InsensitiveLoad(serviceConfigPath)
	if err != nil {
		return nil
	}
	serviceConfigFile.NameMapper = ini.TitleUnderscore

	serviceSection := serviceConfigFile.Section("service")
	if serviceSection == nil {
		return fmt.Errorf("no 'service' section in %s", serviceConfigPath)
	}

	err = parseServiceSection(filepath.Base(filepath.Dir(serviceConfigPath)), serviceSection, projectConfig, serviceConfigPath)

	return err
}

func parseServiceSection(defaultId string, serviceSection *ini.Section, projectConfig *Project, filePath string) error {
	if !serviceSection.HasKey("type") {
		if serviceSection.Name() == "service" {
			return fmt.Errorf("no service type in %s", filePath)
		} else {
			return fmt.Errorf("no service type in %s section [%s] ", filePath, serviceSection.Name())
		}
	}

	var serviceConfig Service

	switch serviceType := serviceSection.Key("type").Value(); serviceType {
	case "dockerfile":
		dockerServiceConfig := &service.DockerfileService{
			Dockerfile:      "Dockerfile",
			PathPrefixStrip: false,
		}

		serviceConfig = dockerServiceConfig
	case "helm":
		serviceConfig = &service.HelmService{}

	case "manifest":
		serviceConfig = &service.ManifestService{}
	default:
		return fmt.Errorf("unknown service type: %s", serviceType)
	}

	serviceConfig.SetId(defaultId)
	if err := serviceSection.MapTo(serviceConfig); err != nil {
		return err
	}

	projectConfig.Services = append(projectConfig.Services, serviceConfig)

	return nil

}
