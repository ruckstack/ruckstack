package project

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Parse(projectPath string) (*ProjectConfig, error) {
	projectConfigFile, err := ini.InsensitiveLoad(projectPath)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("Project file %s not found", projectPath))
	}
	if err != nil {
		return nil, err
	}

	projectConfigFile.NameMapper = ini.TitleUnderscore

	projectConfig := &ProjectConfig{
		K3sVersion:  "1.18.6+k3s1",
		HelmVersion: "3.2.4",
	}

	projectConfigFile.Section("ruckstack-project").MapTo(projectConfig)

	err = util.Validate(projectConfig)
	if err != nil {
		return nil, err
	}

	matched, _ := regexp.MatchString(`^[a-z0-9-_]+$`, projectConfig.Id)
	if !matched {
		panic("Project id must be lower case, alphanumeric, with no whitespace")
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

	if len(projectConfig.DockerfileServices) == 0 && len(projectConfig.HelmServices) == 0 && len(projectConfig.ManifestServices) == 0 {
		return nil, fmt.Errorf("No services are defined in %s", projectPath)
	}

	return projectConfig, nil

}

func parseServiceFile(serviceConfigPath string, projectConfig *ProjectConfig) error {
	serviceConfigFile, err := ini.InsensitiveLoad(serviceConfigPath)
	util.Check(err)
	serviceConfigFile.NameMapper = ini.TitleUnderscore

	serviceSection := serviceConfigFile.Section("service")
	if serviceSection == nil {
		return fmt.Errorf("no 'service' section in %s", serviceConfigPath)
	}

	err = parseServiceSection(filepath.Base(filepath.Dir(serviceConfigPath)), serviceSection, projectConfig, serviceConfigPath)

	return err
}

func parseServiceSection(defaultId string, serviceSection *ini.Section, projectConfig *ProjectConfig, filePath string) error {
	if !serviceSection.HasKey("type") {
		if serviceSection.Name() == "service" {
			return fmt.Errorf("no service type in %s", filePath)
		} else {
			return fmt.Errorf("no service type in %s section [%s] ", filePath, serviceSection.Name())
		}
	}

	switch serviceType := serviceSection.Key("type").Value(); serviceType {
	case "dockerfile":
		serviceConfig := &DockerfileServiceConfig{
			Id:              defaultId,
			Dockerfile:      "Dockerfile",
			PathPrefixStrip: false,
		}
		serviceSection.MapTo(serviceConfig)

		err := util.Validate(serviceConfig)
		if err != nil {
			return err
		}

		if !filepath.IsAbs(serviceConfig.BaseDir) {
			serviceConfig.BaseDir, err = filepath.Abs(filepath.Join(filepath.Dir(filePath), serviceConfig.BaseDir))
			if err != nil {
				return nil
			}
		}

		projectConfig.DockerfileServices = append(projectConfig.DockerfileServices, serviceConfig)
	case "helm":
		serviceConfig := &HelmServiceConfig{
			Id: defaultId,
		}
		serviceSection.MapTo(serviceConfig)

		err := util.Validate(serviceConfig)
		if err != nil {
			return err
		}

		projectConfig.HelmServices = append(projectConfig.HelmServices, serviceConfig)
	case "manifest":
		serviceConfig := &ManifestServiceConfig{
			Id: defaultId,
		}
		serviceSection.MapTo(serviceConfig)

		err := util.Validate(serviceConfig)
		if err != nil {
			return err
		}

		projectConfig.ManifestServices = append(projectConfig.ManifestServices, serviceConfig)

	default:
		panic(fmt.Sprintf("Unknown service type: %s", serviceType))
	}

	return nil

}
