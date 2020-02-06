package builder

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/artifact"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/service/dockerfile"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/service/helm"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/service/manifest"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/shared"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func Build(projectFile string, outDir string) {

	util.Check(os.MkdirAll(outDir, 0755))
	buildEnv := prepare(outDir)

	clean(buildEnv)

	projectConfig, err := project.Parse(projectFile)
	util.CheckWithMessage(err, "Error parsing project file %s", projectFile)

	//saveDockerImages(dockerImages, "app.tar", artifactWriter)

	appFilename := projectConfig.Id + "_" + projectConfig.Version + ".run"

	app := artifact.NewArtifact(appFilename, outDir)

	app.PackageConfig.Id = projectConfig.Id
	app.PackageConfig.Name = projectConfig.Name
	app.PackageConfig.Version = projectConfig.Version
	app.PackageConfig.BuildTime = time.Now().Unix()

	app.PackageConfig.FilePermissions = map[string]internal.InstalledFileConfig{
		".package.config": {
			AdminGroupReadable: true,
		},
		"config/*": {
			AdminGroupReadable: true,
		},
		"bin/*": {
			AdminGroupReadable: true,
			Executable:         true,
		},
		"lib/*": {
			AdminGroupReadable: true,
		},
		"lib/k3s": {
			AdminGroupReadable: true,
			Executable:         true,
		},
	}

	app.AddAssetDir("internal/ruckstack/builder/resources/install-dir", ".")
	app.AddFile(downloadFile(fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s", projectConfig.K3sVersion), buildEnv), "lib/k3s")
	app.AddFile(downloadFile(fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s-airgap-images-amd64.tar", projectConfig.K3sVersion), buildEnv), "data/agent/images/k3s.tar")

	serverBinaryName := projectConfig.ServerBinaryName
	if serverBinaryName == "" {
		serverBinaryName = projectConfig.Id
	}
	app.AddAsset("out/system-control", "bin/"+serverBinaryName)
	app.PackageConfig.SystemControlName = serverBinaryName

	for _, service := range projectConfig.DockerfileServices {
		dockerfile.AddService(service, app, projectConfig, buildEnv)
	}
	for _, service := range projectConfig.HelmServices {
		helm.AddService(service, app, projectConfig, buildEnv)
	}
	for _, service := range projectConfig.ManifestServices {
		manifest.AddService(service, app, projectConfig, buildEnv)
	}

	app.SaveDockerImages(buildEnv)

	app.Build(projectConfig, buildEnv)

}

func prepare(outDir string) *shared.BuildEnvironment {
	buildEnv := new(shared.BuildEnvironment)
	buildEnv.WorkDir = outDir + string(filepath.Separator) + "work"
	buildEnv.CacheDir = outDir + string(filepath.Separator) + "cache"
	return buildEnv
}

func clean(buildEnv *shared.BuildEnvironment) {
	log.Printf("Cleaning work directory %s...", buildEnv.WorkDir)
	//clean work dir
	err := os.RemoveAll(buildEnv.WorkDir)
	util.Check(err)
}

func downloadFile(url string, buildEnv *shared.BuildEnvironment) string {

	cacheKey := regexp.MustCompile(`https?://.+?/`).ReplaceAllString(url, "")

	savePath := buildEnv.CacheDir + string(filepath.Separator) + cacheKey

	saveDir, _ := filepath.Split(savePath)
	err := os.MkdirAll(saveDir, 0755)
	util.Check(err)

	_, err = os.Stat(savePath)
	if err == nil {
		log.Println(savePath + " already exists. Not re-downloading")
		return savePath
	}

	log.Println("Downloading " + url + "...")
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("Cannot download %s: %s", url, resp.Status))
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(savePath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return savePath
}
