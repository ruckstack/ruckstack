package helm

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/global"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/installer"
	"github.com/ruckstack/ruckstack/internal/ruckstack/helm"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"io/ioutil"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/kubectl/pkg/scheme"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func AddService(serviceConfig *project.HelmServiceConfig, app *installer.Installer, projectConfig *project.ProjectConfig) {
	log.Println("Service type: helm")

	serviceBuildDir := filepath.Join(global.BuildEnvironment.WorkDir, serviceConfig.Id)
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
			"namespace": "kube-system",
		},
		"spec": map[string]interface{}{
			"chart":           "https://%{KUBERNETES_API}%/static/charts/" + serviceConfig.Id + ".tgz",
			"targetNamespace": "default",
		},
	}

	out, err := yaml.Marshal(manifest)
	util.Check(err)

	manifestPath := filepath.Join(serviceBuildDir, "manifest.yaml")
	err = ioutil.WriteFile(manifestPath, out, 0644)
	util.Check(err)

	app.AddFile(manifestPath, "data/server/manifests/"+serviceConfig.Id+".yaml")

	chartFile := helm.DownloadChart(repo, chart, serviceConfig.Version)
	app.AddFile(chartFile, "data/server/static/charts/"+serviceConfig.Id+".tgz")

	loadedChart, err := loader.Load(chartFile)
	util.Check(err)

	saveDockerImages(loadedChart, app)

}

func saveDockerImages(loadedChart *chart.Chart, app *installer.Installer) {
	options := chartutil.ReleaseOptions{
		Name:      "testRelease",
		Namespace: "default",
	}

	cvals, err := chartutil.CoalesceValues(loadedChart, map[string]interface{}{})
	valuesToRender, err := chartutil.ToRenderValues(loadedChart, cvals, options, nil)

	render, err := engine.Render(loadedChart, valuesToRender)
	util.Check(err)

	for filename, data := range render {
		data = strings.TrimSpace(data)
		if len(data) == 0 {
			continue
		}
		if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {

			decode := scheme.Codecs.UniversalDeserializer().Decode
			obj, groupVersionKind, err := decode([]byte(data), nil, nil)
			if err != nil {
				fmt.Printf("Cannot parse %s: %s\n", filename, err)
				continue
			}

			var podSpec *coreV1.PodSpec

			switch groupVersionKind.Kind {
			case "StatefulSet":
				podSpec = &obj.(*appV1.StatefulSet).Spec.Template.Spec
			case "Deployment":
				podSpec = &obj.(*appV1.Deployment).Spec.Template.Spec
			case "DaemonSet":
				podSpec = &obj.(*appV1.DaemonSet).Spec.Template.Spec
			case "ReplicaSet":
				podSpec = &obj.(*appV1.ReplicaSet).Spec.Template.Spec

			case "Pod":
				podSpec = &obj.(*coreV1.Pod).Spec
			}

			if podSpec != nil {
				for _, container := range podSpec.Containers {
					log.Printf("See ss image %s\n", container.Image)
					app.IncludeDockerImage(container.Image)
				}
			}

		}
	}
}
