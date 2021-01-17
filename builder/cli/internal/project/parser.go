package project

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"regexp"
)

func Parse(projectPath string) (*Project, error) {
	content, err := os.Open(projectPath)

	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %s", projectPath, err)
	}
	defer content.Close()

	return ParseData(content, projectPath)

}

func ParseData(data io.Reader, projectPath string) (*Project, error) {
	decoder := yaml.NewDecoder(data)
	decoder.SetStrict(true)

	projectConfig := Project{
		K3sVersion:  environment.PackagedK3sVersion,
		HelmVersion: environment.PackagedHelmVersion,
	}
	if err := decoder.Decode(&projectConfig); err != nil {
		return nil, fmt.Errorf("error parsing %s: %s", projectPath, err)
	}

	matched, _ := regexp.MatchString(`^[a-z0-9-_]+$`, projectConfig.Id)
	if !matched {
		return nil, fmt.Errorf("project id must be lower case, alphanumeric, with no whitespace")
	}

	if len(projectConfig.GetServices()) == 0 {
		return nil, fmt.Errorf("No services are defined in %s", projectPath)
	}

	if projectConfig.ManagerFilename == "" {
		projectConfig.ManagerFilename = projectConfig.Id
	}

	for i, _ := range projectConfig.Proxy {
		if projectConfig.Proxy[i].ServicePort == 0 {
			projectConfig.Proxy[i].ServicePort = projectConfig.Proxy[i].Port
		}
	}

	for i, _ := range projectConfig.ManifestServices {
		projectConfig.ManifestServices[i].ProjectVersion = projectConfig.Version
		projectConfig.ManifestServices[i].ProjectId = projectConfig.Id
	}

	for i, _ := range projectConfig.HelmServices {
		projectConfig.HelmServices[i].ProjectVersion = projectConfig.Version
		projectConfig.HelmServices[i].ProjectId = projectConfig.Id
	}

	for i, _ := range projectConfig.DockerfileServices {
		projectConfig.DockerfileServices[i].ProjectVersion = projectConfig.Version
		projectConfig.DockerfileServices[i].ProjectId = projectConfig.Id
	}

	if err := projectConfig.Validate(); err != nil {
		return nil, err
	}

	return &projectConfig, nil

}
