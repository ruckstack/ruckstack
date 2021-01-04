package helm

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/helmpath/xdg"
	"helm.sh/helm/v3/pkg/repo"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	helmHome           string
	repoConfigYamlPath string
)

func init() {
	ui.VPrintln("Initializing helm...")
	defer ui.VPrintln("Initializing helm...DONE")

	helmHome = environment.CachePath("helm")
	if err := os.Setenv(helmpath.ConfigHomeEnvVar, helmHome); err != nil {
		ui.VPrintf("cannot set helm env: %s", err)
	}

	repoConfigYamlPath = helmpath.ConfigPath("repositories.yaml")

	ui.VPrintf("Helm config: %s", repoConfigYamlPath)

	if err := os.MkdirAll(helmHome, 0755); err != nil {
		ui.Fatalf("Cannot create Helm directory: %s", err)
	}

	if err := os.Setenv(xdg.CacheHomeEnvVar, path.Join(helmHome, "cache")); err != nil {
		ui.VPrintln(err)
	}
	if err := os.Setenv(xdg.ConfigHomeEnvVar, path.Join(helmHome, "config")); err != nil {
		ui.VPrintln(err)
	}
	if err := os.Setenv(xdg.DataHomeEnvVar, path.Join(helmHome, "data")); err != nil {
		ui.VPrintln(err)
	}

	_, err := os.Stat(repoConfigYamlPath)
	if os.IsNotExist(err) {
		ui.VPrintln("Creating new helm metadata...")

		if err := os.MkdirAll(filepath.Dir(repoConfigYamlPath), 0755); err != nil {
			ui.VPrintln(err)
		}

		entry := repo.Entry{
			Name: "stable",
			URL:  "https://charts.helm.sh/stable",
		}

		repoFile := repo.NewFile()
		repoFile.Add(&entry)

		if err := repoFile.WriteFile(repoConfigYamlPath, 0664); err != nil {
			ui.VPrintln(err)
		}
	}
}

func ReIndex() error {
	ui.Println("Reindexing helm repositories...")
	defer ui.Println("Reindexing helm repositories...DONE")

	repoFile, err := openRepoConfig()
	if err != nil {
		return err
	}

	for _, repository := range repoFile.Repositories {
		ui.VPrintf("Reindexing %s from %s", repository.Name, repository.URL)
		chartRepository, err := repo.NewChartRepository(repository, getter.All(cli.New()))
		if err != nil {
			return err
		}

		_, err = chartRepository.DownloadIndexFile()
		if err != nil {
			return err
		}
	}

	return nil
}

func openRepoConfig() (*repo.File, error) {
	ui.VPrintf("Loading helm config from %s", repoConfigYamlPath)
	repoFile, err := repo.LoadFile(repoConfigYamlPath)
	if err != nil {
		return nil, err
	}
	return repoFile, nil
}

/**
Download the given chart. Returns the path to the downloaded file. Will not re-download.
*/
func DownloadChart(chart string, version string) (string, error) {
	splitChart := strings.Split(chart, "/")
	repoName := splitChart[0]
	chartName := splitChart[1]

	downloadDir := environment.CachePath("download/helm/" + repoName)
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", err
	}

	chartDownloader := &downloader.ChartDownloader{
		Out: ui.GetOutput(),
		//Keyring:  f.keyring,
		Verify:           downloader.VerifyNever,
		RepositoryConfig: repoConfigYamlPath,
		RepositoryCache:  filepath.Join(helmHome, "cache", "helm", "repository"),
		Getters:          getter.All(cli.New()),
	}

	_, err := os.Stat(chartDownloader.RepositoryCache + "/" + repoName + "-index.yaml")
	if os.IsNotExist(err) {
		ui.VPrintf("No index for repo %s. Forcing re-index", repoName)
		if err := ReIndex(); err != nil {
			return "", err
		}
	}
	_, err = os.Stat(chartDownloader.RepositoryCache + "/" + repoName + "-index.yaml")
	if os.IsNotExist(err) {
		return "", fmt.Errorf("no Helm repository named %s is configured. Add it with `ruckstack helm repo add`", repoName)
	}

	savePath := filepath.Join(downloadDir, chartName+"-"+version+".tgz")
	_, err = os.Stat(savePath)
	if os.IsNotExist(err) {
		defer ui.StartProgressf("Downloading chart %s", filepath.Base(savePath)).Stop()

		savePath, _, err := chartDownloader.DownloadTo(repoName+"/"+chartName, version, downloadDir)
		if err != nil {
			errMessage := err.Error()
			errMessage = strings.Replace(errMessage, ". (try 'helm repo update')", "", 1)
			return "", fmt.Errorf(errMessage)
		}
		ui.VPrintf("Saved to %s", savePath)
	} else {
		ui.VPrintf("Already downloaded chart %s to %s", filepath.Base(savePath), savePath)
	}

	return savePath, nil
}
