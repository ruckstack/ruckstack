package helm

import (
	"fmt"
	"github.com/mitchellh/go-wordwrap"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/global"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/helmpath/xdg"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var (
	helmHome        string
	chartDownloader *downloader.ChartDownloader

	reindexed bool = false
)

func Setup() {
	helmHome = getHelmHome()
	err := os.MkdirAll(helmHome, 0755)
	util.Check(err)

	os.Setenv(xdg.CacheHomeEnvVar, filepath.Join(helmHome, "/cache"))
	os.Setenv(xdg.ConfigHomeEnvVar, filepath.Join(helmHome, "/config"))
	os.Setenv(xdg.DataHomeEnvVar, filepath.Join(helmHome, "/data"))

	_, err = os.Stat(helmpath.ConfigPath("repositories.yaml"))
	if os.IsNotExist(err) {
		fmt.Println("Creating new helm metadata...")

		err := os.MkdirAll(filepath.Dir(helmpath.ConfigPath("repositories.yaml")), 0755)
		util.Check(err)

		entry := repo.Entry{
			Name: "stable",
			URL:  "https://kubernetes-charts.storage.googleapis.com",
		}
		util.Check(err)

		repoFile := repo.NewFile()
		repoFile.Add(&entry)

		err = repoFile.WriteFile(helmpath.ConfigPath("repositories.yaml"), 0664)

		ReIndex()

		fmt.Println("Creating new helm metadata...DONE")
	}
}

func getHelmHome() string {
	if helmHome == "" {
		usr, err := user.Current()
		util.Check(err)

		helmHome = filepath.Join(usr.HomeDir, ".ruckstack", "helm")

	}

	return helmHome
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

func ReIndex() {
	if reindexed {
		return
	}

	fmt.Println("Reindexing helm repositories...")

	repoFile, err := repo.LoadFile(helmpath.ConfigPath("repositories.yaml"))
	util.Check(err)

	for _, repository := range repoFile.Repositories {
		chartRepository, err := repo.NewChartRepository(repository, getter.All(cli.New()))
		util.Check(err)

		//_, err = os.Stat(helmHome.Cache())
		//if os.IsNotExist(err) {
		//	os.MkdirAll(helmHome.Cache(), 0755)
		//}

		_, err = chartRepository.DownloadIndexFile()
		//panic(fmt.Errorf("Looks like %q is not a valid chart repository or cannot be reached: %s", entry.URL, err.Error()))
		util.Check(err)
	}

	reindexed = true
	fmt.Println("Reindexing helm repositories...DONE")
}

func Search(chartRepo string, chartName string) {
	repositories, err := repo.LoadFile(helmpath.ConfigPath("repositories.yaml"))
	util.Check(err)

	repository := repositories.Get(chartRepo)
	if repository == nil {
		panic(fmt.Sprintf("Unknown helm repository %s", chartRepo))
	}

	indexFile, err := repo.LoadIndexFile(filepath.Join(helmHome, "cache", "helm", "repository", helmpath.CacheIndexFile(repository.Name)))
	util.Check(err)

	versions := indexFile.Entries[chartName]
	if versions == nil {
		panic(fmt.Sprintf("Unknown chart %s/%s", chartRepo, chartName))
	}

	latestVersion := versions[0]

	appVersion := latestVersion.AppVersion
	if appVersion == "" {
		appVersion = "n/a"
	}

	fmt.Printf(`
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

	fmt.Println("\nAll Available Versions:")

	for _, version := range versions {
		appVersion := version.AppVersion
		if appVersion == "" {
			appVersion = "n/a"
		}
		fmt.Printf("  %s (App Version %s)\n", version.Version, appVersion)
	}

}

func DownloadChart(repo string, chart string, version string) string {
	cacheDir := global.BuildEnvironment.CacheDir + string(filepath.Separator) + "helm" + string(filepath.Separator) + repo
	err := os.MkdirAll(cacheDir, 0755)
	util.Check(err)

	savePath := cacheDir + string(filepath.Separator) + chart + "-" + version + ".tgz"
	_, err = os.Stat(savePath)
	if os.IsNotExist(err) {
		log.Printf("Downloading chart %s...", filepath.Base(savePath))

		_, _, err := getDownloader().DownloadTo(repo+"/"+chart, version, cacheDir)
		util.Check(err)
	} else {
		log.Printf("%s already exists. Not re-downloading", savePath)
	}

	return savePath
}

//func GetImages(chartPath string) {
//	loader.Load(chartPath)
//}
