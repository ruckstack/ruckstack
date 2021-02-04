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

dockerfileServices:
  - id: test_dockerfile
    dockerfile: Dockerfile
    http:
        port: 8080
        pathPrefix: /
`), "in-memory")

	assert.NoError(t, err)

	assert.Equal(t, "test", project.Id)
	assert.Equal(t, "test", project.ManagerFilename)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "1.0.5", project.Version)
	assert.NotEmpty(t, project.K3sVersion)
	assert.NotEmpty(t, project.HelmVersion)

	assert.Equal(t, 1, len(project.GetServices()))

	assert.Equal(t, project.Id, project.DockerfileServices[0].ProjectId)
	assert.Equal(t, project.Version, project.DockerfileServices[0].ProjectVersion)

	assert.Equal(t, "test_dockerfile", project.DockerfileServices[0].Id)
	assert.Equal(t, "dockerfile", project.DockerfileServices[0].GetType())
	assert.Equal(t, 8080, project.DockerfileServices[0].Http.Port)
	assert.Equal(t, "Dockerfile", project.DockerfileServices[0].Dockerfile)
	assert.Equal(t, "", project.DockerfileServices[0].ServiceVersion)
	assert.Equal(t, "/", project.DockerfileServices[0].Http.PathPrefix)
	assert.Equal(t, false, project.DockerfileServices[0].Http.PathPrefixStrip)
}

func TestParse_FullProject(t *testing.T) {
	project, err := ParseData(strings.NewReader(`
id: test
name: Test Full Project
version: 1.0.5

proxy:
  - serviceName: firstService
    servicePort: 1234
    port: 5678
  - serviceName: other
    port: 8888

helmRepos:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami

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
    env:
      - name: postgres_password
        secretName: postgresql
        secretKey: postgresql-postgres-password
      - name: map_location
        configMapName: config_name
        configMapKey: config_key
    mount:
      - name: postgres-dir
        secretName: postgresql
        path: /path/to/pg
      - name: config-files
        configMapName: myConfig
        path: /path/to/config

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
	assert.Equal(t, "Test Full Project", project.Name)
	assert.Equal(t, "1.0.5", project.Version)
	assert.NotEmpty(t, project.K3sVersion)
	assert.NotEmpty(t, project.HelmVersion)

	assert.Equal(t, "bitnami", project.HelmRepos[0].Name)
	assert.Equal(t, "https://charts.bitnami.com/bitnami", project.HelmRepos[0].Url)

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

	assert.Equal(t, 2, len(project.DockerfileServices[1].Env))
	assert.Equal(t, "postgres_password", project.DockerfileServices[1].Env[0].Name)
	assert.Equal(t, "postgresql", project.DockerfileServices[1].Env[0].SecretName)
	assert.Equal(t, "postgresql-postgres-password", project.DockerfileServices[1].Env[0].SecretKey)

	assert.Equal(t, "map_location", project.DockerfileServices[1].Env[1].Name)
	assert.Equal(t, "config_name", project.DockerfileServices[1].Env[1].ConfigMapName)
	assert.Equal(t, "config_key", project.DockerfileServices[1].Env[1].ConfigMapKey)

	assert.Equal(t, 2, len(project.DockerfileServices[1].Mount))
	assert.Equal(t, "postgres-dir", project.DockerfileServices[1].Mount[0].Name)
	assert.Equal(t, "postgresql", project.DockerfileServices[1].Mount[0].SecretName)
	assert.Equal(t, "/path/to/pg", project.DockerfileServices[1].Mount[0].Path)

	assert.Equal(t, "config-files", project.DockerfileServices[1].Mount[1].Name)
	assert.Equal(t, "myConfig", project.DockerfileServices[1].Mount[1].ConfigMapName)
	assert.Equal(t, "/path/to/config", project.DockerfileServices[1].Mount[1].Path)

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
