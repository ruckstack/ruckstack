package builder

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/builder/global"
	"github.com/ruckstack/ruckstack/builder/internal/builder/installer"
	"github.com/ruckstack/ruckstack/builder/internal/builder/service"
	"github.com/ruckstack/ruckstack/builder/internal/project"
	"github.com/ruckstack/ruckstack/common/ui"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
)

func Build(projectFile string, outDir string) error {

	if err := prepareBuildEnvironment(outDir); err != nil {
		return err
	}

	projectConfig, err := project.Parse(projectFile)
	if err != nil {
		return fmt.Errorf("error parsing project file %s: %s", projectFile, err)
	}

	installFile, err := installer.NewInstaller(projectConfig)
	if err != nil {
		return nil
	}

	//add standard files to the installer
	if err := installFile.AddResourceDir("install_dir", "."); err != nil {
		return err
	}
	if err := installFile.AddDownloadedNestedFile(fmt.Sprintf("https://get.helm.sh/helm-v%s-linux-amd64.tar.gz", url.PathEscape(projectConfig.HelmVersion)), "linux-amd64/helm", "lib/helm"); err != nil {
		return err
	}
	if err := installFile.AddDownloadedFile(fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s", url.PathEscape(projectConfig.K3sVersion)), "lib/k3s"); err != nil {
		return err
	}
	if err := installFile.AddDownloadedFile(fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s-airgap-images-amd64.tar", url.PathEscape(projectConfig.K3sVersion)), "data/agent/images/k3s.tar"); err != nil {
		return err
	}
	if err := installFile.AddAsset("out/system-control", "bin/"+projectConfig.ServerBinaryName); err != nil {
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

	if err := installFile.SaveDockerImages(); err != nil {
		return err
	}
	if err := installFile.ClearDockerImages(); err != nil {
		return err
	}

	return installFile.Build()
}

/**
Configures the global BuildEnvironment data and creates/cleans directories as needed.
*/
func prepareBuildEnvironment(outDir string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	global.BuildEnvironment.OutDir = outDir
	global.BuildEnvironment.WorkDir = filepath.Join(outDir, "work")
	global.BuildEnvironment.CacheDir = filepath.Join(usr.HomeDir, ".ruckstack", "cache")

	ui.Printf("Cleaning out directory %s...", global.BuildEnvironment.OutDir)
	if err := os.RemoveAll(global.BuildEnvironment.WorkDir); err != nil {
		return err
	}

	for _, dir := range []string{global.BuildEnvironment.WorkDir, global.BuildEnvironment.CacheDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
