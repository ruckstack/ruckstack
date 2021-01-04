package helm

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func AddRepository(repoName string, repoUrl string, username string, password string) error {
	ui.StartProgressf("Adding new Helm repository %s", repoName)
	ui.VPrintf("Indexing new helm repository %s as %s", repoUrl, repoName)
	newEntry := &repo.Entry{
		Name:     repoName,
		URL:      repoUrl,
		Username: username,
		Password: password,
	}

	repoConfig, err := openRepoConfig()
	if err != nil {
		return err
	}

	if repoConfig.Has(repoName) {
		ui.Fatalf("Repository %s is already configured", repoName)
	}

	repoConfig.Add(newEntry)

	chartRepository, err := repo.NewChartRepository(newEntry, getter.All(cli.New()))
	if err != nil {
		return err
	}

	_, err = chartRepository.DownloadIndexFile()
	if err != nil {
		return err
	}

	if err := repoConfig.WriteFile(repoConfigYamlPath, 0644); err != nil {
		return fmt.Errorf("error writing %s: %s", repoConfigYamlPath, err)
	}

	return nil
}

func RemoveRepository(repoName string) error {
	repoConfig, err := openRepoConfig()
	if err != nil {
		return err
	}

	removed := repoConfig.Remove(repoName)

	if removed {
		ui.Printf("Helm repository %s removed", repoName)
	} else {
		ui.Printf("No repository named %s was found", repoName)
	}

	if err := repoConfig.WriteFile(repoConfigYamlPath, 0644); err != nil {
		return fmt.Errorf("cannot write %s: %s", repoConfigYamlPath, err)
	}

	return nil
}
