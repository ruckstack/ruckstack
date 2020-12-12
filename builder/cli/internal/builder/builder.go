package builder

import (
	"compress/flate"
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/cli/internal/project"
	"github.com/ruckstack/ruckstack/common/ui"
	"net/url"
	"os"
	"path/filepath"
)

func Build() error {
	projectConfig, err := project.Parse(environment.ProjectDir + "/ruckstack.yaml")
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

	installFile, err := install_file.StartCreation(installerPath, flate.BestCompression)
	if err != nil {
		return err
	}
	installFile.PackageConfig.Id = projectConfig.Id
	installFile.PackageConfig.Name = projectConfig.Name
	installFile.PackageConfig.Version = projectConfig.Version
	installFile.PackageConfig.ManagerFilename = projectConfig.ManagerFilename

	//add custom files
	customFiles := map[string]string{
		filepath.Join("ruckstack", "site-down.png"): "data/web/ops/img/public/site-down.png",
	}

	for customFile, targetPath := range customFiles {
		_, err := os.Stat(customFile)
		if err == nil {
			ui.Printf("Adding custom %s", customFile)
			if err := installFile.AddFile(customFile, targetPath); err != nil {
				return fmt.Errorf("error adding custom file: %s", err)
			}
		}
	}

	//add install_dir
	installDir, err := environment.ResourcePath("install_dir")
	if err != nil {
		return err
	}
	if err = installFile.AddDirectory(installDir, ""); err != nil {
		return err
	}

	//add system-control
	systemControl, err := environment.ResourcePath("system-control")
	if err != nil {
		return err
	}
	if err = installFile.AddFile(systemControl, fmt.Sprintf("bin/%s", projectConfig.ManagerFilename)); err != nil {
		return err
	}

	//add 3rd party files
	//for user performance reasons, these should be pre-downloaded in builder/cli/cmd/commands/internal_build.go
	if err := installFile.AddDownloadedNestedFile(fmt.Sprintf("https://get.helm.sh/helm-v%s-linux-amd64.tar.gz", url.PathEscape(projectConfig.HelmVersion)), "linux-amd64/helm", "lib/helm"); err != nil {
		return err
	}
	if err := installFile.AddDownloadedFile(fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s", url.PathEscape(projectConfig.K3sVersion)), "lib/k3s"); err != nil {
		return err
	}
	if err := installFile.AddDownloadedFile(fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s-airgap-images-amd64.tar", url.PathEscape(projectConfig.K3sVersion)), "data/agent/images/k3s.tar"); err != nil {
		return err
	}

	for _, serviceConfig := range projectConfig.GetServices() {
		if err := serviceConfig.Build(installFile); err != nil {
			return err
		}
	}

	return installFile.CompleteCreation()
}
