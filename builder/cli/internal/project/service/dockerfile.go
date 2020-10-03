package service

import (
	"crypto/md5"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/internal/docker"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type DockerfileService struct {
	//Common fields
	Id             string `validate:"required"`
	Type           string `validate:"required,oneof=dockerfile helm manifest"`
	Port           int    `validate:"required"`
	ProjectId      string
	ProjectVersion string

	//Unique Fields
	Dockerfile      string `validate:"required"`
	ServiceVersion  string `ini:"service_version"`
	UrlPath         string `ini:"base_url"`
	PathPrefixStrip bool   `ini:"path_prefix_strip"`

	serviceWorkDir string
}

func (serviceConfig *DockerfileService) GetId() string {
	return serviceConfig.Id
}

func (serviceConfig *DockerfileService) SetId(id string) {
	serviceConfig.Id = id
}

func (serviceConfig *DockerfileService) GetType() string {
	return serviceConfig.Type
}

func (serviceConfig *DockerfileService) GetPort() int {
	return serviceConfig.Port
}

func (serviceConfig *DockerfileService) SetProjectId(projectId string) {
	serviceConfig.ProjectId = projectId
}

func (serviceConfig *DockerfileService) SetProjectVersion(projectVersion string) {
	serviceConfig.ProjectVersion = projectVersion
}

func (service *DockerfileService) Validate(structValidator *validator.Validate) error {
	if err := structValidator.Struct(service); err != nil {
		return err
	}

	return nil
}

func (service *DockerfileService) Build(app *install_file.InstallFile) error {
	ui.Printf("Building Dockerfile Service %s", service.Id)

	if service.ServiceVersion == "" {
		service.ServiceVersion = fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))[0:6]
	}

	service.serviceWorkDir = environment.TempPath(service.Id + "-*")
	if err := os.MkdirAll(service.serviceWorkDir+"/chart/templates", 0755); err != nil {
		return err
	}

	dockerTag := "build.local/" + service.ProjectId + "/" + service.Id + ":" + service.ServiceVersion
	if err := service.buildContainer(dockerTag); err != nil {
		return err
	}

	if err := service.writeChart(); err != nil {
		return err
	}

	if err := service.writeDaemonSet(); err != nil {
		return err
	}

	if err := service.writeService(); err != nil {
		return err
	}

	if service.UrlPath != "" {
		if err := service.writeIngress(); err != nil {
			return err
		}
	}

	chart, err := service.buildChart()
	if err != nil {
		return err
	}

	if err := app.AddHelmChart(chart, service.Id); err != nil {
		return err
	}

	return nil
}

func (service *DockerfileService) buildContainer(dockerTag string) error {
	dockerfile := service.Dockerfile
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	err := docker.ImageBuild(environment.ProjectDir+"/"+dockerfile,
		[]string{dockerTag},
		map[string]string{
			"ruckstack.built": "true",
		})

	if err != nil {
		return err
	}

	return nil
}

func (service *DockerfileService) writeChart() error {
	chart := map[string]interface{}{
		"apiVersion": "v1",
		"name":       service.Id,
		"version":    service.ServiceVersion,
		"appVersion": service.ProjectVersion,
	}

	out, err := yaml.Marshal(chart)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path.Join(service.serviceWorkDir, "chart/Chart.yaml"), out, 0644); err != nil {
		return err
	}
	return nil
}

func (service *DockerfileService) writeDaemonSet() error {
	daemonSet := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "DaemonSet",
		"metadata": map[string]interface{}{
			"name": service.Id,
			"labels": map[string]string{
				"app": service.Id,
			},
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"app": service.Id,
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": service.Id,
					},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  service.Id,
							"image": "build.local/" + service.ProjectId + "/" + service.Id + ":" + service.ServiceVersion,
							"ports": []map[string]int{
								{"containerPort": service.Port},
							},
						},
					},
				},
			},
		},
	}

	out, err := yaml.Marshal(daemonSet)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(service.serviceWorkDir+"/chart/templates/daemonset.yaml", out, 0644); err != nil {
		return err
	}

	return nil
}

func (service *DockerfileService) writeService() error {
	serviceDef := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name": service.Id,
			"labels": map[string]string{
				"app": service.Id,
			},
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"app": service.Id,
			},
			"ports": []map[string]interface{}{
				{
					"protocol": "TCP",
					"port":     service.Port,
				},
			},
		},
	}

	out, err := yaml.Marshal(serviceDef)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(service.serviceWorkDir+"/chart/templates/service.yaml", out, 0644); err != nil {
		return err
	}

	return nil
}

func (service *DockerfileService) writeIngress() error {
	annotations := map[string]string{}
	if service.PathPrefixStrip {
		annotations["traefik.frontend.rule.type"] = "PathPrefixStrip"
	}

	ingress := map[string]interface{}{
		"apiVersion": "extensions/v1beta1",
		"kind":       "Ingress",
		"metadata": map[string]interface{}{
			"name":        service.Id,
			"annotations": annotations,
			"labels": map[string]string{
				"app": service.Id,
			},
		},
		"spec": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"http": map[string]interface{}{
						"paths": []map[string]interface{}{
							{
								"path": service.UrlPath,
								"backend": map[string]interface{}{
									"serviceName": service.Id,
									"servicePort": service.Port,
								},
							},
						},
					},
				},
			},
			//"ports": []map[string]interface{}{
			//	{
			//		"protocol":      "TCP",
			//		"containerPort": serviceConfig.Port,
			//	},
			//},
		},
	}

	out, err := yaml.Marshal(ingress)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(service.serviceWorkDir+"/chart/templates/ingress.yaml", out, 0644)
}

func (service *DockerfileService) buildChart() (string, error) {
	chartFilePath := service.serviceWorkDir + "/" + service.Id + ".tgz"

	ui.Printf("Creating %s...", chartFilePath)

	if err := global_util.TarDirectory(service.serviceWorkDir+"/chart", chartFilePath, true); err != nil {
		return "", err
	}

	return chartFilePath, nil
}
