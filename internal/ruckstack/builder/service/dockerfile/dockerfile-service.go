package dockerfile

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/artifact"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/shared"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func AddService(serviceConfig *project.DockerfileServiceConfig, app *artifact.Artifact, projectConfig *project.ProjectConfig, buildConfig *shared.BuildEnvironment) {
	log.Println("Service type: dockerfile")

	if serviceConfig.ServiceVersion == "" {
		serviceConfig.ServiceVersion = fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))[0:6]
	}

	serviceWorkDir := buildConfig.WorkDir + "/" + serviceConfig.Id + "/chart"
	err := os.MkdirAll(serviceWorkDir+"/templates", 0755)
	util.Check(err)

	writeChart(serviceConfig, projectConfig, buildConfig, serviceWorkDir)
	writeDaemonSet(serviceConfig, projectConfig, buildConfig, serviceWorkDir)
	writeService(serviceConfig, projectConfig, buildConfig, serviceWorkDir)
	if serviceConfig.UrlPath != "" {
		writeIngress(serviceConfig, projectConfig, buildConfig, serviceWorkDir)
	}

	app.AddFile(buildChart(serviceConfig, projectConfig, buildConfig, serviceWorkDir), "data/server/static/charts/"+serviceConfig.Id+".tgz")
	app.AddFile(writeManifest(serviceConfig, projectConfig, buildConfig), "data/server/manifests/"+serviceConfig.Id+".yaml")

	dockerTag := projectConfig.Id + "/" + serviceConfig.Id + ":" + serviceConfig.ServiceVersion
	dockerBuildCmd := exec.Command("docker", "build", "-t", dockerTag, ".")
	dockerBuildCmd.Dir = serviceConfig.BaseDir
	dockerBuildCmd.Stdout = os.Stdout
	dockerBuildCmd.Stderr = os.Stderr
	err = dockerBuildCmd.Run()
	util.Check(err)

	app.IncludeDockerImage(dockerTag)

}

func writeChart(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, config *shared.BuildEnvironment, serviceBuildDir string) {
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

func writeDaemonSet(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, config *shared.BuildEnvironment, serviceBuildDir string) {
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

func writeService(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, config *shared.BuildEnvironment, serviceBuildDir string) {
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

func writeIngress(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, config *shared.BuildEnvironment, serviceBuildDir string) {
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

func buildChart(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, buildConfig *shared.BuildEnvironment, serviceBuildDir string) string {
	chartFilePath := buildConfig.WorkDir + "/" + serviceConfig.Id + "/" + serviceConfig.Id + ".tgz"
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
			Name:    serviceConfig.Id + "/" + relativePath,
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

func writeManifest(serviceConfig *project.DockerfileServiceConfig, projectConfig *project.ProjectConfig, buildConfig *shared.BuildEnvironment) string {
	manifest := map[string]interface{}{
		"apiVersion": "helm.cattle.io/v1",
		"kind":       "HelmChart",
		"metadata": map[string]interface{}{
			"name": serviceConfig.Id,
		},
		"spec": map[string]interface{}{
			"chart":   fmt.Sprintf("https://%%{KUBERNETES_API}%%/static/charts/%s.tgz", serviceConfig.Id),
			"version": serviceConfig.ServiceVersion,
		},
	}

	out, err := yaml.Marshal(manifest)
	util.Check(err)

	outputPath := buildConfig.WorkDir + "/" + serviceConfig.Id + ".yaml"
	err = ioutil.WriteFile(outputPath, out, 0644)
	util.Check(err)

	return outputPath
}
