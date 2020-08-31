package dockerfile

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/global"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/installer"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func AddService(serviceConfig *project.DockerfileServiceConfig, app *installer.Installer, projectConfig *project.ProjectConfig) {
	log.Println("Service type: dockerfile")

	if serviceConfig.ServiceVersion == "" {
		serviceConfig.ServiceVersion = fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))[0:6]
	}

	serviceWorkDir := global.BuildEnvironment.WorkDir + "/" + serviceConfig.Id + "/chart"
	err := os.MkdirAll(serviceWorkDir+"/templates", 0755)
	util.Check(err)

	writeChart(serviceConfig, projectConfig, serviceWorkDir)
	writeDaemonSet(serviceConfig, projectConfig, serviceWorkDir)
	writeService(serviceConfig, projectConfig, serviceWorkDir)
	if serviceConfig.UrlPath != "" {
		writeIngress(serviceConfig, projectConfig, serviceWorkDir)
	}

	app.AddFile(buildChart(serviceConfig, projectConfig, serviceWorkDir), "data/server/static/charts/"+serviceConfig.Id+".tgz")
	app.AddFile(writeManifest(serviceConfig, projectConfig), "data/server/manifests/"+serviceConfig.Id+".yaml")

	dockerTag := projectConfig.Id + "/" + serviceConfig.Id + ":" + serviceConfig.ServiceVersion
	buildContainer(dockerTag, serviceConfig, err)

	app.IncludeDockerImage(dockerTag)

}

func buildContainer(dockerTag string, serviceConfig *project.DockerfileServiceConfig, err error) {
	dockerBuildCmd := exec.Command("docker", "build",
		"-t", dockerTag,
		"--label", "ruckstack.built=true",
		".")
	dockerBuildCmd.Dir = serviceConfig.BaseDir
	dockerBuildCmd.Stdout = os.Stdout
	dockerBuildCmd.Stderr = os.Stderr
	err = dockerBuildCmd.Run()
	util.Check(err)
}

func writeChart(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, serviceBuildDir string) {
	chart := map[string]interface{}{
		"apiVersion": "v1",
		"name":       serviceConfig.Id,
		"version":    serviceConfig.ServiceVersion,
		"appVersion": projectConfig.Version,
	}

	out, err := yaml.Marshal(chart)
	util.Check(err)

	err = ioutil.WriteFile(serviceBuildDir+"/Chart.yaml", out, 0644)
	util.Check(err)
}

func writeDaemonSet(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, serviceBuildDir string) {
	daemonSet := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "DaemonSet",
		"metadata": map[string]interface{}{
			"name": serviceConfig.Id,
			"labels": map[string]string{
				"app": serviceConfig.Id,
			},
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"app": serviceConfig.Id,
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": serviceConfig.Id,
					},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  serviceConfig.Id,
							"image": projectConfig.Id + "/" + serviceConfig.Id + ":" + serviceConfig.ServiceVersion,
							"ports": []map[string]int{
								{"containerPort": serviceConfig.Port},
							},
						},
					},
				},
			},
		},
	}

	out, err := yaml.Marshal(daemonSet)
	util.Check(err)

	err = ioutil.WriteFile(serviceBuildDir+"/templates/daemonset.yaml", out, 0644)
	util.Check(err)
}

func writeService(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, serviceBuildDir string) {
	service := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name": serviceConfig.Id,
			"labels": map[string]string{
				"app": serviceConfig.Id,
			},
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"app": serviceConfig.Id,
			},
			"ports": []map[string]interface{}{
				{
					"protocol": "TCP",
					"port":     serviceConfig.Port,
				},
			},
		},
	}

	out, err := yaml.Marshal(service)
	util.Check(err)

	err = ioutil.WriteFile(serviceBuildDir+"/templates/service.yaml", out, 0644)
	util.Check(err)
}

func writeIngress(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, serviceBuildDir string) {
	annotations := map[string]string{}
	if serviceConfig.PathPrefixStrip {
		annotations["traefik.frontend.rule.type"] = "PathPrefixStrip"
	}

	ingress := map[string]interface{}{
		"apiVersion": "extensions/v1beta1",
		"kind":       "Ingress",
		"metadata": map[string]interface{}{
			"name":        serviceConfig.Id,
			"annotations": annotations,
			"labels": map[string]string{
				"app": serviceConfig.Id,
			},
		},
		"spec": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"http": map[string]interface{}{
						"paths": []map[string]interface{}{
							{
								"path": serviceConfig.UrlPath,
								"backend": map[string]interface{}{
									"serviceName": serviceConfig.Id,
									"servicePort": serviceConfig.Port,
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
	util.Check(err)

	err = ioutil.WriteFile(serviceBuildDir+"/templates/ingress.yaml", out, 0644)
	util.Check(err)
}

func buildChart(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, serviceBuildDir string) string {
	chartFilePath := global.BuildEnvironment.WorkDir + "/" + serviceConfig.Id + "/" + serviceConfig.Id + ".tgz"
	chartFile, err := os.Create(chartFilePath)
	util.Check(err)

	log.Printf("Creating %s...", chartFilePath)

	gzipWriter := gzip.NewWriter(chartFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	filepath.Walk(serviceBuildDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(serviceBuildDir, path)
		util.Check(err)

		file, err := os.Open(path)
		util.Check(err)
		defer file.Close()

		header := &tar.Header{
			Name:    serviceConfig.Id + "/" + strings.ReplaceAll(relativePath, "\\", "/"),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    0644,
		}

		err = tarWriter.WriteHeader(header)
		util.Check(err)

		_, err = io.Copy(tarWriter, file)
		util.Check(err)

		return nil
	})

	return chartFilePath
}

func writeManifest(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig) string {
	manifest := map[string]interface{}{
		"apiVersion": "helm.cattle.io/v1",
		"kind":       "HelmChart",
		"metadata": map[string]interface{}{
			"name":      serviceConfig.Id,
			"namespace": "kube-system",
		},
		"spec": map[string]interface{}{
			"chart":           fmt.Sprintf("https://%%{KUBERNETES_API}%%/static/charts/%s.tgz", serviceConfig.Id),
			"version":         serviceConfig.ServiceVersion,
			"targetNamespace": "default",
		},
	}

	out, err := yaml.Marshal(manifest)
	util.Check(err)

	outputPath := global.BuildEnvironment.WorkDir + "/" + serviceConfig.Id + ".yaml"
	err = ioutil.WriteFile(outputPath, out, 0644)
	util.Check(err)

	return outputPath
}
