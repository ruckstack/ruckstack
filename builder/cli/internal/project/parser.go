package project

import (
	"fmt"
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
		K3sVersion:  "1.19.2+k3s1",
		HelmVersion: "3.2.4",
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

	if projectConfig.SystemControlName == "" {
		projectConfig.SystemControlName = projectConfig.Id
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
