package project

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParse_MinimalProject(t *testing.T) {
	project, err := ParseData(strings.NewReader(`
id: test
name: Test Project
version: 1.0.5

proxy:
  - serviceName: firstService
    servicePort: 1234
    port: 5678
  - serviceName: other
    port: 8888


dockerfileServices:
  - id: test_dockerfile
    dockerfile: Dockerfile
    http:
        port: 8080
        pathPrefix: /
  - id: test_dockerfile2
    dockerfile: Dockerfile2
    http:
        port: 8082
        pathPrefix: /2

helmServices:
  - id: test_helm
    chart: stable/helm-test
    version: 7.3.9

manifestServices:
  - id: test_manifest
    manifest: test-manifest.yaml

`), "in-memory")

	assert.NoError(t, err)

	assert.Equal(t, "test", project.Id)
	assert.Equal(t, "test", project.ManagerFilename)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "1.0.5", project.Version)
	assert.NotEmpty(t, project.K3sVersion)
	assert.NotEmpty(t, project.HelmVersion)

	assert.Equal(t, 4, len(project.GetServices()))

	assert.Equal(t, project.Id, project.DockerfileServices[0].ProjectId)
	assert.Equal(t, project.Version, project.DockerfileServices[0].ProjectVersion)
	assert.Equal(t, project.Id, project.DockerfileServices[1].ProjectId)
	assert.Equal(t, project.Version, project.DockerfileServices[1].ProjectVersion)

	assert.Equal(t, "test_dockerfile", project.DockerfileServices[0].Id)
	assert.Equal(t, "dockerfile", project.DockerfileServices[0].GetType())
	assert.Equal(t, 8080, project.DockerfileServices[0].Http.Port)
	assert.Equal(t, "Dockerfile", project.DockerfileServices[0].Dockerfile)
	assert.Equal(t, "", project.DockerfileServices[0].ServiceVersion)
	assert.Equal(t, "/", project.DockerfileServices[0].Http.PathPrefix)
	assert.Equal(t, false, project.DockerfileServices[0].Http.PathPrefixStrip)

	assert.Equal(t, "test_dockerfile2", project.DockerfileServices[1].Id)
	assert.Equal(t, 8082, project.DockerfileServices[1].Http.Port)
	assert.Equal(t, "Dockerfile2", project.DockerfileServices[1].Dockerfile)
	assert.Equal(t, "/2", project.DockerfileServices[1].Http.PathPrefix)

	assert.Equal(t, "test_helm", project.HelmServices[0].Id)
	assert.Equal(t, "helm", project.HelmServices[0].GetType())
	assert.Equal(t, "stable/helm-test", project.HelmServices[0].Chart)
	assert.Equal(t, "7.3.9", project.HelmServices[0].Version)

	assert.Equal(t, "test_manifest", project.ManifestServices[0].Id)
	assert.Equal(t, "manifest", project.ManifestServices[0].GetType())
	assert.Equal(t, "test-manifest.yaml", project.ManifestServices[0].Manifest)

	assert.Equal(t, 2, len(project.Proxy))
	assert.Equal(t, "firstService", project.Proxy[0].ServiceName)
	assert.Equal(t, 1234, project.Proxy[0].ServicePort)
	assert.Equal(t, 5678, project.Proxy[0].Port)

	assert.Equal(t, "other", project.Proxy[1].ServiceName)
	assert.Equal(t, 8888, project.Proxy[1].ServicePort)
	assert.Equal(t, 8888, project.Proxy[1].Port)
}
