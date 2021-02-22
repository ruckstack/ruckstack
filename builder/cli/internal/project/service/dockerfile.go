package service

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/internal/docker"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type DockerfileService struct {
	//Common fields
	Id             string `validate:"required"`
	ProjectId      string
	ProjectVersion string

	//Unique Fields
	Dockerfile     string `validate:"required"`
	ServiceVersion string `yaml:"serviceVersion"`
	Http           DockerfileServiceHttp
	Env            []DockerfileServiceEnv
	Mount          []DockerfileServiceMount

	serviceWorkDir string
}

type DockerfileServiceHttp struct {
	ContainerPort   int    `yaml:"containerPort"`
	PathPrefix      string `yaml:"pathPrefix"`
	PathPrefixStrip bool   `yaml:"pathPrefixStrip"`
}

type DockerfileServiceEnv struct {
	Name          string `validate:"required"`
	SecretName    string `yaml:"secretName"`
	SecretKey     string `yaml:"secretKey"`
	ConfigMapName string `yaml:"configMapName"`
	ConfigMapKey  string `yaml:"configMapKey"`
}

type DockerfileServiceMount struct {
	Name          string `validate:"required"`
	SecretName    string `yaml:"secretName"`
	ConfigMapName string `yaml:"configMapName"`
	Path          string `validate:"required"`
}

func (serviceConfig *DockerfileService) GetId() string {
	return serviceConfig.Id
}

func (serviceConfig *DockerfileService) SetId(id string) {
	serviceConfig.Id = id
}

func (serviceConfig *DockerfileService) GetType() string {
	return "dockerfile"
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

	for _, env := range service.Env {
		if (env.SecretKey != "" || env.SecretName != "") && (env.ConfigMapKey != "" || env.ConfigMapName != "") {
			return fmt.Errorf("environment variable %s cannot specify both secret and configMap configurations", env.Name)
		}

		if env.SecretKey == "" && env.ConfigMapKey == "" {
			return fmt.Errorf("environment variable %s must specify either secret and configMap configurations", env.Name)
		}

		if env.SecretKey != "" && env.SecretName == "" {
			return fmt.Errorf("environment variable %s must specify both secretKey and secretName", env.Name)
		}

		if env.ConfigMapKey != "" && env.ConfigMapName == "" {
			return fmt.Errorf("environment variable %s must specify both configMapKey and configMapName", env.Name)
		}
	}

	for _, mount := range service.Mount {
		if !regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$").MatchString(mount.Name) { //regexp from k8s
			return fmt.Errorf("mount name '%s' must consist of lower case alphanumeric characters or '-'", mount.Name)
		}
		if mount.SecretName != "" && mount.ConfigMapName != "" {
			return fmt.Errorf("mount point %s cannot specify both secret and configMap configurations", mount.Name)
		}
	}

	return nil
}

func (service *DockerfileService) Build(app *install_file.InstallFile) error {
	ui.Printf("Building Dockerfile Service %s", service.Id)

	if service.ServiceVersion == "" {
		service.ServiceVersion = fmt.Sprintf("0.0.%d", time.Now().Unix())
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

	if service.Http.PathPrefix != "" {
		if err := service.writeIngress(); err != nil {
			return err
		}
	}

	chart, err := service.buildChart()
	if err != nil {
		return err
	}

	if err := app.AddHelmChart(chart, service.Id, nil); err != nil {
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

	envDef := []map[string]interface{}{}
	for _, envConfig := range service.Env {
		if envConfig.SecretKey != "" {
			envDef = append(envDef, map[string]interface{}{
				"name": strings.ToUpper(envConfig.Name),
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": envConfig.SecretName,
						"key":  envConfig.SecretKey,
					},
				},
			})
		} else if envConfig.ConfigMapKey != "" {
			envDef = append(envDef, map[string]interface{}{
				"name": strings.ToUpper(envConfig.Name),
				"valueFrom": map[string]interface{}{
					"configMapKeyRef": map[string]interface{}{
						"name": envConfig.ConfigMapName,
						"key":  envConfig.ConfigMapKey,
					},
				},
			})
		}
	}

	volumeMounts := []map[string]interface{}{}
	volumes := []map[string]interface{}{}
	for _, mountConfig := range service.Mount {
		volumeMounts = append(volumeMounts, map[string]interface{}{
			"name":      mountConfig.Name,
			"mountPath": mountConfig.Path,
			"readOnly":  true,
		})

		if mountConfig.SecretName != "" {
			volumes = append(volumes, map[string]interface{}{
				"name": mountConfig.Name,
				"secret": map[string]interface{}{
					"secretName": mountConfig.SecretName,
				},
			})
		} else if mountConfig.ConfigMapName != "" {
			volumes = append(volumes, map[string]interface{}{
				"name": mountConfig.Name,
				"configMap": map[string]interface{}{
					"name": mountConfig.ConfigMapName,
				},
			})
		}
	}

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
								{"containerPort": service.Http.ContainerPort},
							},
							"env":          envDef,
							"volumeMounts": volumeMounts,
						},
					},
					"volumes": volumes,
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
					"port":     service.Http.ContainerPort,
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
	if service.Http.PathPrefixStrip {
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
								"path": service.Http.PathPrefix,
								"backend": map[string]interface{}{
									"serviceName": service.Id,
									"servicePort": service.Http.ContainerPort,
								},
							},
						},
					},
				},
			},
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
