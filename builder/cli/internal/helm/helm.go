package helm

import (
	"fmt"
	"github.com/mitchellh/go-wordwrap"
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
	helmHome string
)

func init() {
	ui.VPrintln("Initializing helm...")
	defer ui.VPrintln("Initializing helm...DONE")

	helmHome = environment.CachePath("helm")

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

	_, err := os.Stat(helmpath.ConfigPath("repositories.yaml"))
	if os.IsNotExist(err) {
		ui.VPrintln("Creating new helm metadata...")

		if err := os.MkdirAll(filepath.Dir(helmpath.ConfigPath("repositories.yaml")), 0755); err != nil {
			ui.VPrintln(err)
		}

		entry := repo.Entry{
			Name: "stable",
			URL:  "https://charts.helm.sh/stable",
		}

		repoFile := repo.NewFile()
		repoFile.Add(&entry)

		if err := repoFile.WriteFile(helmpath.ConfigPath("repositories.yaml"), 0664); err != nil {
			ui.VPrintln(err)
		}
	}
}

func ReIndex() error {
	ui.Println("Reindexing helm repositories...")
	defer ui.Println("Reindexing helm repositories...DONE")

	repoFile, err := repo.LoadFile(helmpath.ConfigPath("repositories.yaml"))
	if err != nil {
		return err
	}

	for _, repository := range repoFile.Repositories {
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

func Search(chartRepo string, chartName string) error {
	repositories, err := repo.LoadFile(helmpath.ConfigPath("repositories.yaml"))
	if err != nil {
		return err
	}

	repository := repositories.Get(chartRepo)
	if repository == nil {
		return fmt.Errorf("unknown helm repository: %s", chartRepo)
	}

	indexFile, err := repo.LoadIndexFile(filepath.Join(helmHome, "cache", "helm", "repository", helmpath.CacheIndexFile(repository.Name)))
	if err != nil {
		if os.IsNotExist(err) {
			if err = ReIndex(); err != nil {
				return err
			}
			return Search(chartRepo, chartName)
		}

		return err
	}

	versions := indexFile.Entries[chartName]
	if versions == nil {
		return fmt.Errorf("unknown helm chart: %s/%s", chartRepo, chartName)
	}

	latestVersion := versions[0]

	appVersion := latestVersion.AppVersion
	if appVersion == "" {
		appVersion = "n/a"
	}

	ui.Printf(`
Chart: %s/%s
Latest Version: %s (App Version %s)
%s

%s
`,
		chartRepo,
		latestVersion.Name,
		latestVersion.Version,
		appVersion,
		latestVersion.Home,
		wordwrap.WrapString(latestVersion.Description, 80))

	ui.Println("\nAll Available Versions:")

	for _, version := range versions {
		appVersion := version.AppVersion
		if appVersion == "" {
			appVersion = "n/a"
		}
		ui.Printf("  %s (App Version %s)\n", version.Version, appVersion)
	}

	return nil
}

/**
Download the given chart. Returns the path to the downloaded file. Will not re-download.
*/
func DownloadChart(repo string, chart string, version string) (string, error) {
	downloadDir := environment.CachePath("download/helm/" + repo)
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", err
	}

	chartDownloader := &downloader.ChartDownloader{
		Out: ui.GetOutput(),
		//Keyring:  f.keyring,
		Verify:           downloader.VerifyNever,
		RepositoryConfig: helmpath.ConfigPath("repositories.yaml"),
		RepositoryCache:  filepath.Join(helmHome, "cache", "helm", "repository"),
		Getters:          getter.All(cli.New()),
	}

	_, err := os.Stat(chartDownloader.RepositoryCache + "/" + repo + "-index.yaml")
	if os.IsNotExist(err) {
		ui.VPrintf("No index for repo %s. Forcing re-index", repo)
		if err := ReIndex(); err != nil {
			return "", err
		}
	}

	savePath := filepath.Join(downloadDir, chart+"-"+version+".tgz")
	_, err = os.Stat(savePath)
	if os.IsNotExist(err) {
		defer ui.StartProgressf("Downloading chart %s", filepath.Base(savePath)).Stop()

		savePath, _, err := chartDownloader.DownloadTo(repo+"/"+chart, version, downloadDir)
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
