package service

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/builder/global"
	"github.com/ruckstack/ruckstack/builder/internal/builder/installer"
	"github.com/ruckstack/ruckstack/builder/internal/project"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type DockerfileService struct {
	ServiceConfig  *project.DockerfileServiceConfig
	ProjectConfig  *project.ProjectConfig
	serviceWorkDir string
}

func (service *DockerfileService) Build(app *installer.Installer) error {
	ui.Println("Service type: dockerfile")

	if service.ServiceConfig.ServiceVersion == "" {
		service.ServiceConfig.ServiceVersion = fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))[0:6]
	}

	service.serviceWorkDir = path.Join(global.BuildEnvironment.WorkDir, service.ServiceConfig.Id, "chart")
	if err := os.MkdirAll(path.Join(service.serviceWorkDir, "templates"), 0755); err != nil {
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

	if service.ServiceConfig.UrlPath != "" {
		if err := service.writeIngress(); err != nil {
			return err
		}
	}

	chart, err := service.buildChart()
	if err != nil {
		return err
	}
	if err := app.AddFile(chart, "data/server/static/charts/"+service.ServiceConfig.Id+".tgz"); err != nil {
		return err
	}

	manifest, err := service.writeManifest()
	if err != nil {
		return err
	}
	if err := app.AddFile(manifest, "data/server/manifests/"+service.ServiceConfig.Id+".yaml"); err != nil {
		return err
	}

	dockerTag := service.ProjectConfig.Id + "/" + service.ServiceConfig.Id + ":" + service.ServiceConfig.ServiceVersion
	if err := service.buildContainer(dockerTag); err != nil {
		return err
	}

	if err := app.IncludeDockerImage(dockerTag); err != nil {
		return err
	}

	return nil
}

func (service *DockerfileService) buildContainer(dockerTag string) error {
	dockerBuildCmd := exec.Command("docker", "build",
		"-t", dockerTag,
		"--label", "ruckstack.built=true",
		".")
	dockerBuildCmd.Dir = service.ServiceConfig.BaseDir
	dockerBuildCmd.Stdout = os.Stdout
	dockerBuildCmd.Stderr = os.Stderr
	return dockerBuildCmd.Run()
}

func (service *DockerfileService) writeChart() error {
	chart := map[string]interface{}{
		"apiVersion": "v1",
		"name":       service.ServiceConfig.Id,
		"version":    service.ServiceConfig.ServiceVersion,
		"appVersion": service.ProjectConfig.Version,
	}

	out, err := yaml.Marshal(chart)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path.Join(service.serviceWorkDir, "Chart.yaml"), out, 0644); err != nil {
		return err
	}
	return nil
}

func (service *DockerfileService) writeDaemonSet() error {
	daemonSet := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "DaemonSet",
		"metadata": map[string]interface{}{
			"name": service.ServiceConfig.Id,
			"labels": map[string]string{
				"app": service.ServiceConfig.Id,
			},
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"app": service.ServiceConfig.Id,
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": service.ServiceConfig.Id,
					},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  service.ServiceConfig.Id,
							"image": service.ProjectConfig.Id + "/" + service.ServiceConfig.Id + ":" + service.ServiceConfig.ServiceVersion,
							"ports": []map[string]int{
								{"containerPort": service.ServiceConfig.Port},
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

	if err := ioutil.WriteFile(service.serviceWorkDir+"/templates/daemonset.yaml", out, 0644); err != nil {
		return err
	}

	return nil
}

func (service *DockerfileService) writeService() error {
	serviceDef := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name": service.ServiceConfig.Id,
			"labels": map[string]string{
				"app": service.ServiceConfig.Id,
			},
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"app": service.ServiceConfig.Id,
			},
			"ports": []map[string]interface{}{
				{
					"protocol": "TCP",
					"port":     service.ServiceConfig.Port,
				},
			},
		},
	}

	out, err := yaml.Marshal(serviceDef)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(service.serviceWorkDir+"/templates/service.yaml", out, 0644); err != nil {
		return err
	}

	return nil
}

func (service *DockerfileService) writeIngress() error {
	annotations := map[string]string{}
	if service.ServiceConfig.PathPrefixStrip {
		annotations["traefik.frontend.rule.type"] = "PathPrefixStrip"
	}

	ingress := map[string]interface{}{
		"apiVersion": "extensions/v1beta1",
		"kind":       "Ingress",
		"metadata": map[string]interface{}{
			"name":        service.ServiceConfig.Id,
			"annotations": annotations,
			"labels": map[string]string{
				"app": service.ServiceConfig.Id,
			},
		},
		"spec": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"http": map[string]interface{}{
						"paths": []map[string]interface{}{
							{
								"path": service.ServiceConfig.UrlPath,
								"backend": map[string]interface{}{
									"serviceName": service.ServiceConfig.Id,
									"servicePort": service.ServiceConfig.Port,
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

	return ioutil.WriteFile(path.Join(service.serviceWorkDir, "templates", "ingress.yaml"), out, 0644)
}

func (service *DockerfileService) buildChart() (string, error) {
	chartFilePath := path.Join(global.BuildEnvironment.WorkDir, service.ServiceConfig.Id, service.ServiceConfig.Id+".tgz")
	chartFile, err := os.Create(chartFilePath)
	if err != nil {
		return "", err
	}

	ui.Printf("Creating %s...", chartFilePath)

	gzipWriter := gzip.NewWriter(chartFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	if err := filepath.Walk(service.serviceWorkDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(service.serviceWorkDir, path)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		defer file.Close()

		header := &tar.Header{
			Name:    service.ServiceConfig.Id + "/" + strings.ReplaceAll(relativePath, "\\", "/"),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    0644,
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return chartFilePath, nil
}

func (service *DockerfileService) writeManifest() (string, error) {
	manifest := map[string]interface{}{
		"apiVersion": "helm.cattle.io/v1",
		"kind":       "HelmChart",
		"metadata": map[string]interface{}{
			"name":      service.ServiceConfig.Id,
			"namespace": "kube-system",
		},
		"spec": map[string]interface{}{
			"chart":           fmt.Sprintf("https://%%{KUBERNETES_API}%%/static/charts/%s.tgz", service.ServiceConfig.Id),
			"version":         service.ServiceConfig.ServiceVersion,
			"targetNamespace": "default",
		},
	}

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return "", err
	}

	outputPath := path.Join(global.BuildEnvironment.WorkDir, service.ServiceConfig.Id+".yaml")
	err = ioutil.WriteFile(outputPath, out, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}
