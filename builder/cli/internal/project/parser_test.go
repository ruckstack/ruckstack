package project

import (
	"github.com/ruckstack/ruckstack/builder/cli/internal/project/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse_MinimalProject(t *testing.T) {
	project, err := Parse([]byte(`
[ruckstack-project]
id: test
name: Test Project
version: 1.0.5

[service-test_dockerfile]
type: dockerfile
base_dir: docker_basedir
port: 8080
base_url: /

[service-test_helm]
TYPE: helm
CHART: stable/helm-test
VERSION: 7.3.9
port: 1234

[service-test_manifest]
type: manifest
manifest: test-manifest.yaml
port: 8888


`))

	assert.NoError(t, err)

	assert.Equal(t, "test", project.Id)
	assert.Equal(t, "test", project.SystemControlName)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "1.0.5", project.Version)
	assert.NotEmpty(t, project.K3sVersion)
	assert.NotEmpty(t, project.HelmVersion)

	assert.Equal(t, 3, len(project.Services))
	firstService := project.Services[0].(*service.DockerfileService)
	secondService := project.Services[1].(*service.HelmService)
	thirdService := project.Services[2].(*service.ManifestService)

	assert.Equal(t, "test_dockerfile", firstService.Id)
	assert.Equal(t, "dockerfile", firstService.Type)
	assert.Equal(t, 8080, firstService.Port)
	assert.Equal(t, "Dockerfile", firstService.Dockerfile)
	assert.Equal(t, "", firstService.ServiceVersion)
	assert.Equal(t, "/", firstService.UrlPath)
	assert.Equal(t, false, firstService.PathPrefixStrip)

	assert.Equal(t, "test_helm", secondService.Id)
	assert.Equal(t, "helm", secondService.Type)
	assert.Equal(t, 1234, secondService.Port)
	assert.Equal(t, "stable/helm-test", secondService.Chart)
	assert.Equal(t, "7.3.9", secondService.Version)

	assert.Equal(t, "test_manifest", thirdService.Id)
	assert.Equal(t, "manifest", thirdService.Type)
	assert.Equal(t, 8888, thirdService.Port)
	assert.Equal(t, "test-manifest.yaml", thirdService.Manifest)
}
