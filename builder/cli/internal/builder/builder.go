package builder

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/service"
	"github.com/ruckstack/ruckstack/builder/cli/internal/project"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"net/url"
	"os"
)

func Build(projectFile string, outDir string) error {

	environment.OutDir = outDir

	ui.Printf("Cleaning out directory...")
	if err := os.RemoveAll(outDir); err != nil {
		return err
	}

	projectConfig, err := project.Parse(projectFile)
	if err != nil {
		return fmt.Errorf("error parsing project file %s: %s", projectFile, err)
	}

	installFile, err := install_file.StartInstallFile(projectConfig)
	if err != nil {
		return nil
	}

	//add install_dir
	installDir, err := environment.ResourcePath("install_dir")
	if err != nil {
		return err
	}
	if err = installFile.AddDirectory(installDir, ""); err != nil {
		return err
	}

	//add 3rd party files
	if err := installFile.AddDownloadedNestedFile(fmt.Sprintf("https://get.helm.sh/helm-v%s-linux-amd64.tar.gz", url.PathEscape(projectConfig.HelmVersion)), "linux-amd64/helm", "lib/helm"); err != nil {
		return err
	}
	if err := installFile.AddDownloadedFile(fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s", url.PathEscape(projectConfig.K3sVersion)), "lib/k3s"); err != nil {
		return err
	}
	if err := installFile.AddDownloadedFile(fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s-airgap-images-amd64.tar", url.PathEscape(projectConfig.K3sVersion)), "data/agent/images/k3s.tar"); err != nil {
		return err
	}

	installFile.PackageConfig.SystemControlName = projectConfig.ServerBinaryName

	for _, serviceConfig := range projectConfig.DockerfileServices {
		builder := &service.DockerfileService{
			ProjectConfig: projectConfig,
			ServiceConfig: serviceConfig,
		}
		if err := builder.Build(installFile); err != nil {
			return err
		}
	}

	for _, serviceConfig := range projectConfig.HelmServices {
		builder := &service.HelmService{
			ProjectConfig: projectConfig,
			ServiceConfig: serviceConfig,
		}
		if err := builder.Build(installFile); err != nil {
			return err
		}
	}
	for _, serviceConfig := range projectConfig.ManifestServices {
		builder := &service.ManifestService{
			ProjectConfig: projectConfig,
			ServiceConfig: serviceConfig,
		}
		if err := builder.Build(installFile); err != nil {
			return err
		}
	}

	return installFile.Build()
}
