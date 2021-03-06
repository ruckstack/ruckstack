package builder

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/internal/bundled"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/builder/internal/helm"
	"github.com/ruckstack/ruckstack/builder/internal/project"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/ui"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func Build(compressionLevel int) error {
	projectConfig, err := project.Parse(filepath.Join(environment.ProjectDir, "ruckstack.yaml"))
	if err != nil {
		return fmt.Errorf("error parsing project: %s", err)
	}

	installerPath := environment.OutPath(projectConfig.Id + "_" + projectConfig.Version + ".installer")
	err = os.Remove(installerPath)
	if os.IsNotExist(err) {
		ui.VPrintf("No existing %s to delete", installerPath)
	} else if err != nil {
		return err
	}

	installFile, err := install_file.StartCreation(installerPath, compressionLevel)
	if err != nil {
		return err
	}
	installFile.PackageConfig.Id = projectConfig.Id
	installFile.PackageConfig.Name = projectConfig.Name
	installFile.PackageConfig.Version = projectConfig.Version

	installFile.SystemConfig.ManagerFilename = projectConfig.ManagerFilename

	for _, proxyConfig := range projectConfig.Proxy {
		servicePort := proxyConfig.ServicePort
		if servicePort == 0 {
			servicePort = proxyConfig.Port
		}
		installFile.SystemConfig.Proxy = append(installFile.SystemConfig.Proxy, config.OpenPort{
			ServiceName: proxyConfig.ServiceName,
			ServicePort: servicePort,
			Port:        proxyConfig.Port,
		})
	}

	//add custom files
	customFiles := map[string]string{
		filepath.Join("ruckstack", "site-down.png"): "data/web/ops/img/public/site-down.png",
	}

	for customFile, targetPath := range customFiles {
		customFileObj, err := os.Open(customFile)
		if err == nil {
			ui.Printf("Adding custom %s", customFile)
			if err := installFile.AddFile(customFileObj, targetPath); err != nil {
				return fmt.Errorf("error adding custom file: %s", err)
			}
			_ = customFileObj.Close()
		}
	}

	//add install_dir
	installDir, err := bundled.OpenDir("install_dir")
	if err != nil {
		return err
	}
	if err = installFile.AddDirectory(installDir, ""); err != nil {
		return err
	}

	//add system-control
	systemControl, err := bundled.OpenFile("system-control")
	if err != nil {
		return err
	}
	if err = installFile.AddFile(systemControl, fmt.Sprintf("bin/%s", projectConfig.ManagerFilename)); err != nil {
		return err
	}

	//add 3rd party files
	if err := installFile.AddDownloadedNestedFile(fmt.Sprintf("https://get.helm.sh/helm-v%s-linux-amd64.tar.gz", url.PathEscape(projectConfig.HelmVersion)), "linux-amd64/helm", "lib/helm"); err != nil {
		return err
	}
	if err := installFile.AddDownloadedFile(fmt.Sprintf("https://github.com/k3s-io/k3s/releases/download/v%s/k3s", url.PathEscape(projectConfig.K3sVersion)), "lib/k3s"); err != nil {
		return err
	}
	if err := installFile.AddDownloadedFile(fmt.Sprintf("https://github.com/k3s-io/k3s/releases/download/v%s/k3s-airgap-images-amd64.tar", url.PathEscape(projectConfig.K3sVersion)), "data/agent/images/k3s.tar"); err != nil {
		return err
	}

	for _, helmConfig := range projectConfig.HelmRepos {
		if err := helm.AddRepository(helmConfig.Name, helmConfig.Url, helmConfig.Username, helmConfig.Password); err != nil {
			if strings.Contains(err.Error(), "is already configured") {
				ui.VPrintf(err.Error())
			} else {
				return err
			}
		}
	}

	for _, serviceConfig := range projectConfig.GetServices() {
		if err := serviceConfig.Build(installFile); err != nil {
			return err
		}
	}

	return installFile.CompleteCreation()
}
