package project

import (
	"github.com/stretchr/testify/assert"
	"strings"
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
base_dir: manifest_basedir
manifest: test-manifest.yaml
port: 8888


`))

	assert.NoError(t, err)

	assert.Equal(t, "test", project.Id)
	assert.Equal(t, "test", project.ServerBinaryName)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "1.0.5", project.Version)
	assert.NotEmpty(t, project.K3sVersion)
	assert.NotEmpty(t, project.HelmVersion)

	assert.Equal(t, 1, len(project.DockerfileServices))
	assert.Equal(t, "test_dockerfile", project.DockerfileServices[0].Id)
	assert.Equal(t, "dockerfile", project.DockerfileServices[0].Type)
	assert.Equal(t, 8080, project.DockerfileServices[0].Port)
	assert.Equal(t, "Dockerfile", project.DockerfileServices[0].Dockerfile)
	assert.True(t, strings.HasSuffix(project.DockerfileServices[0].BaseDir, "docker_basedir"))
	assert.Equal(t, "", project.DockerfileServices[0].ServiceVersion)
	assert.Equal(t, "/", project.DockerfileServices[0].UrlPath)
	assert.Equal(t, false, project.DockerfileServices[0].PathPrefixStrip)

	assert.Equal(t, 1, len(project.HelmServices))
	assert.Equal(t, "test_helm", project.HelmServices[0].Id)
	assert.Equal(t, "helm", project.HelmServices[0].Type)
	assert.Equal(t, 1234, project.HelmServices[0].Port)
	assert.Equal(t, "stable/helm-test", project.HelmServices[0].Chart)
	assert.Equal(t, "7.3.9", project.HelmServices[0].Version)

	assert.Equal(t, 1, len(project.ManifestServices))
	assert.Equal(t, "test_manifest", project.ManifestServices[0].Id)
	assert.Equal(t, "manifest", project.ManifestServices[0].Type)
	assert.Equal(t, 8888, project.ManifestServices[0].Port)
	assert.Equal(t, "test-manifest.yaml", project.ManifestServices[0].Manifest)
	assert.Equal(t, "manifest_basedir", project.ManifestServices[0].BaseDir)
}
