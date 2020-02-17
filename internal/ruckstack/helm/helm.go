package helm

import (
	"fmt"
	"github.com/mitchellh/go-wordwrap"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
	"os"
	"os/user"
	"path/filepath"
)

var (
	helmHome        helmpath.Home
	chartDownloader *downloader.ChartDownloader

	reindexed bool = false
)

func Setup() {
	helmHome := getHelmHome()
	stat, err := os.Stat(helmHome.String())
	if os.IsNotExist(err) {
		err := os.MkdirAll(helmHome.String(), 0755)
		util.Check(err)
	} else {
		if !stat.IsDir() {
			panic(fmt.Sprintf("%s is not a directory", helmHome.String()))
		}
	}

	stat, err = os.Stat(helmHome.RepositoryFile())
	if os.IsNotExist(err) {
		fmt.Println("Creating new helm metadata...")

		err := os.MkdirAll(filepath.Dir(helmHome.RepositoryFile()), 0755)
		util.Check(err)

		entry := repo.Entry{
			Name:  "stable",
			URL:   "https://kubernetes-charts.storage.googleapis.com",
			Cache: helmHome.CacheIndex("stable"),
		}
		util.Check(err)

		repoFile := repo.NewRepoFile()
		repoFile.Add(&entry)

		err = repoFile.WriteFile(helmHome.RepositoryFile(), 0664)

		ReIndex()

		fmt.Println("Creating new helm metadata...DONE")
	}
}

func getHelmHome() helmpath.Home {
	if helmHome == "" {
		usr, err := user.Current()
		util.Check(err)

		helmHome = helmpath.Home(usr.HomeDir + string(filepath.Separator) + ".ruckstack" + string(filepath.Separator) + "helm")

	}

	return helmHome
}

func getDownloader() *downloader.ChartDownloader {
	if chartDownloader == nil {
		chartDownloader = &downloader.ChartDownloader{
			HelmHome: getHelmHome(),
			Out:      os.Stdout,
			//Keyring:  f.keyring,
			Verify:  downloader.VerifyNever,
			Getters: getter.All(helm_env.EnvSettings{}),
		}
	}

	return chartDownloader
}

func ReIndex() {
	if reindexed {
		return
	}

	fmt.Println("Reindexing helm repositories...")

	repoFile, err := repo.LoadRepositoriesFile(getHelmHome().RepositoryFile())
	util.Check(err)

	for _, repository := range repoFile.Repositories {
		chartRepository, err := repo.NewChartRepository(repository, getter.All(helm_env.EnvSettings{}))
		util.Check(err)

		_, err = os.Stat(helmHome.Cache())
		if os.IsNotExist(err) {
			os.MkdirAll(helmHome.Cache(), 0755)
		}

		// In this case, the cacheFile is always absolute. So passing empty string is safe
		err = chartRepository.DownloadIndexFile("")
		//panic(fmt.Errorf("Looks like %q is not a valid chart repository or cannot be reached: %s", entry.URL, err.Error()))
		util.Check(err)
	}

	reindexed = true
	fmt.Println("Reindexing helm repositories...DONE")
}

func Search(chartRepo string, chartName string) {
	repositories, err := repo.LoadRepositoriesFile(getHelmHome().RepositoryFile())
	util.Check(err)

	repository, found := repositories.Get(chartRepo)
	if !found {
		panic(fmt.Sprintf("Unknown helm repository %s", chartRepo))
	}

	indexFile, err := repo.LoadIndexFile(repository.Cache)
	util.Check(err)

	versions := indexFile.Entries[chartName]
	if versions == nil {
		panic(fmt.Sprintf("Unknown chart %s/%s", chartRepo, chartName))
	}

	lastestVersion := versions[0]
	fmt.Printf(`
Chart: %s/%s
Latest Version: %s (App Version %s)
%s

%s
`,
		chartRepo,
		lastestVersion.GetName(),
		lastestVersion.GetVersion(),
		lastestVersion.GetAppVersion(),
		lastestVersion.GetHome(),
		wordwrap.WrapString(lastestVersion.GetDescription(), 80))

	fmt.Println("\nAll Available Versions:")

	for _, version := range versions {
		fmt.Printf("  %s (App Version %s)\n", version.GetVersion(), version.GetAppVersion())
	}

}

func Download() {

	//to, verification, err := downloader.DownloadTo("stable/mariadb", "7.3.8", "/tmp/mariadb-helm")
	//util.Check(err)
	//log.Println(to)
	//log.Println(verification.FileName)

}
