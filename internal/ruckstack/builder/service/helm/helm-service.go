package helm

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/artifact"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/shared"
	"github.com/ruckstack/ruckstack/internal/ruckstack/helm"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func AddService(serviceConfig *project.HelmServiceConfig, app *artifact.Artifact, projectConfig *project.ProjectConfig, buildEnv *shared.BuildEnvironment) {
	log.Println("Service type: helm")

	serviceBuildDir := buildEnv.WorkDir + "/" + serviceConfig.Id
	err := os.MkdirAll(serviceBuildDir, 0755)
	util.Check(err)

	splitChart := strings.Split(serviceConfig.Chart, "/")
	repo := splitChart[0]
	chart := splitChart[1]

	manifest := map[string]interface{}{
		"apiVersion": "helm.cattle.io/v1",
		"kind":       "HelmChart",
		"metadata": map[string]interface{}{
			"name":      serviceConfig.Id,
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"chart":   serviceConfig.Chart,
			"version": serviceConfig.Version,
		},
	}

	out, err := yaml.Marshal(manifest)
	util.Check(err)

	manifestPath := serviceBuildDir + "/manifest.yaml"
	err = ioutil.WriteFile(manifestPath, out, 0644)
	util.Check(err)

	app.AddFile(manifestPath, "data/server/manifests/"+serviceConfig.Id+".yaml")

	chartFile := helm.DownloadChart(repo, chart, serviceConfig.Version, buildEnv)
	app.AddFile(chartFile, "data/server/static/charts/"+serviceConfig.Id+".tgz")
}
