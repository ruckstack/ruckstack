package service

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/global"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/installer"
	"github.com/ruckstack/ruckstack/internal/ruckstack/helm"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
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

type HelmService struct {
	ServiceConfig  *project.HelmServiceConfig
	ProjectConfig  *project.ProjectConfig
	serviceWorkDir string
}

func (service *HelmService) Build(app *installer.Installer) error {
	log.Println("Service type: helm")

	service.serviceWorkDir = filepath.Join(global.BuildEnvironment.WorkDir, service.ServiceConfig.Id)
	if err := os.MkdirAll(service.serviceWorkDir, 0755); err != nil {
		return err
	}

	splitChart := strings.Split(service.ServiceConfig.Chart, "/")
	repo := splitChart[0]
	chartName := splitChart[1]

	manifest := map[string]interface{}{
		"apiVersion": "helm.cattle.io/v1",
		"kind":       "HelmChart",
		"metadata": map[string]interface{}{
			"name":      service.ServiceConfig.Id,
			"namespace": "kube-system",
		},
		"spec": map[string]interface{}{
			"chart":           "https://%{KUBERNETES_API}%/static/charts/" + service.ServiceConfig.Id + ".tgz",
			"targetNamespace": "default",
		},
	}

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(service.serviceWorkDir, "manifest.yaml")
	if err := ioutil.WriteFile(manifestPath, out, 0644); err != nil {
		return err
	}

	if err := app.AddFile(manifestPath, "data/server/manifests/"+service.ServiceConfig.Id+".yaml"); err != nil {
		return err
	}

	chartFile, err := helm.DownloadChart(repo, chartName, service.ServiceConfig.Version)
	if err != nil {
		return err
	}

	if err := app.AddFile(chartFile, "data/server/static/charts/"+service.ServiceConfig.Id+".tgz"); err != nil {
		return err
	}

	loadedChart, err := loader.Load(chartFile)
	if err != nil {
		return err
	}

	return service.saveDockerImages(loadedChart, app)
}

func (service *HelmService) saveDockerImages(loadedChart *chart.Chart, app *installer.Installer) error {
	options := chartutil.ReleaseOptions{
		Name:      "testRelease",
		Namespace: "default",
	}

	cvals, err := chartutil.CoalesceValues(loadedChart, map[string]interface{}{})
	if err != nil {
		return err
	}
	valuesToRender, err := chartutil.ToRenderValues(loadedChart, cvals, options, nil)
	if err != nil {
		return err
	}

	render, err := engine.Render(loadedChart, valuesToRender)
	if err != nil {
		return err
	}

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
					if err := app.IncludeDockerImage(container.Image); err != nil {
						return err
					}
				}
			}

		}
	}

	return nil
}
