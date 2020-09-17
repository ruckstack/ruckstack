package helm

import (
	"fmt"
	"github.com/mitchellh/go-wordwrap"
	"github.com/ruckstack/ruckstack/builder/internal/builder/global"
	"github.com/ruckstack/ruckstack/common/ui"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/helmpath/xdg"
	"helm.sh/helm/v3/pkg/repo"
	"os"
	"os/user"
	"path"
	"path/filepath"
)

var (
	helmHome        string
	chartDownloader *downloader.ChartDownloader

	reindexed bool = false
)

func Setup() error {
	helmHome, err := getHelmHome()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(helmHome, 0755); err != nil {
		return err
	}

	if err := os.Setenv(xdg.CacheHomeEnvVar, path.Join(helmHome, "cache")); err != nil {
		return err
	}
	if err := os.Setenv(xdg.ConfigHomeEnvVar, path.Join(helmHome, "config")); err != nil {
		return err
	}
	if err := os.Setenv(xdg.DataHomeEnvVar, path.Join(helmHome, "data")); err != nil {
		return err
	}

	_, err = os.Stat(helmpath.ConfigPath("repositories.yaml"))
	if os.IsNotExist(err) {
		ui.Println("Creating new helm metadata...")

		if err := os.MkdirAll(filepath.Dir(helmpath.ConfigPath("repositories.yaml")), 0755); err != nil {
			return err
		}

		entry := repo.Entry{
			Name: "stable",
			URL:  "https://kubernetes-charts.storage.googleapis.com",
		}

		repoFile := repo.NewFile()
		repoFile.Add(&entry)

		if err := repoFile.WriteFile(helmpath.ConfigPath("repositories.yaml"), 0664); err != nil {
			return err
		}

		if err := ReIndex(); err != nil {
			return err
		}

		ui.Println("Creating new helm metadata...DONE")
	}

	return nil
}

func getHelmHome() (string, error) {
	if helmHome == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}

		helmHome = filepath.Join(usr.HomeDir, ".ruckstack", "helm")

	}

	return helmHome, nil
}

func getDownloader() *downloader.ChartDownloader {
	if chartDownloader == nil {
		chartDownloader = &downloader.ChartDownloader{
			Out: os.Stdout,
			//Keyring:  f.keyring,
			Verify:           downloader.VerifyNever,
			RepositoryConfig: helmpath.ConfigPath("repositories.yaml"),
			RepositoryCache:  filepath.Join(helmHome, "cache", "helm", "repository"),
			Getters:          getter.All(cli.New()),
		}
	}

	return chartDownloader
}

func ReIndex() error {
	if reindexed {
		return nil
	}

	ui.Println("Reindexing helm repositories...")

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

	reindexed = true
	ui.Println("Reindexing helm repositories...DONE")

	return nil
}

func Search(chartRepo string, chartName string) error {
	repositories, err := repo.LoadFile(helmpath.ConfigPath("repositories.yaml"))
	if err != nil {
		return err
	}

	repository := repositories.Get(chartRepo)
	if repository == nil {
		return fmt.Errorf("unknown helm repository %s", chartRepo)
	}

	indexFile, err := repo.LoadIndexFile(filepath.Join(helmHome, "cache", "helm", "repository", helmpath.CacheIndexFile(repository.Name)))
	if err != nil {
		return err
	}

	versions := indexFile.Entries[chartName]
	if versions == nil {
		return fmt.Errorf("unknown chart %s/%s", chartRepo, chartName)
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

func DownloadChart(repo string, chart string, version string) (string, error) {
	cacheDir := global.BuildEnvironment.CacheDir + string(filepath.Separator) + "helm" + string(filepath.Separator) + repo
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	savePath := cacheDir + string(filepath.Separator) + chart + "-" + version + ".tgz"
	_, err := os.Stat(savePath)
	if os.IsNotExist(err) {
		ui.Printf("Downloading chart %s...", filepath.Base(savePath))

		_, _, err := getDownloader().DownloadTo(repo+"/"+chart, version, cacheDir)
		if err != nil {
			return "", err
		}
	} else {
		ui.Printf("%s already exists. Not re-downloading", savePath)
	}

	return savePath, nil
}
