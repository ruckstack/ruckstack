package install_file

import (
	"bytes"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreatingInstallFile(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not create large files")
	}

	output := new(bytes.Buffer)
	ui.SetOutput(output)

	testTempRoot := environment.TempPath("install_file_test-*")
	environment.OutDir = testTempRoot + "/out"
	_ = os.MkdirAll(environment.OutDir, 0755)

	assert.FileExists(t, environment.RuckstackHome+"/builder/cli/install_root/resources/installer", "compiled installer does not exist. Should be created by BUILD.sh")

	installFile, err := StartCreation(environment.OutPath("test-project.installer"))
	assert.NoError(t, err)
	installFile.PackageConfig.Id = "test-project"
	installFile.PackageConfig.Version = "5.6.7"

	assert.Regexp(t, "out/test-project.installer$", installFile.file.Name())
	assert.FileExists(t, installFile.file.Name())

	//add some content
	assert.NoError(t, installFile.AddDirectory(environment.RuckstackHome+"/common", ""))
	assert.NoError(t, installFile.AddFile(environment.RuckstackHome+"/BUILD.sh", "was-build.sh"))
	assert.NoError(t, installFile.AddDownloadedFile("http://example.com/index.html", "example.html"))
	assert.NoError(t, installFile.AddDownloadedNestedFile("https://get.helm.sh/helm-v3.3.3-linux-amd64.tar.gz", "linux-amd64/README.md", "from-download/file.here"))
	assert.NoError(t, installFile.AddImage("alpine:3.12"))

	testManifestContent, err := ioutil.ReadFile(environment.RuckstackHome + "/builder/cli/internal/builder/install_file/install_file_test_manifest.yaml")
	assert.NoError(t, err)
	assert.NoError(t, installFile.AddImagesInManifest(testManifestContent))

	//now build
	err = installFile.CompleteCreation()
	assert.NoError(t, err)

	assert.Contains(t, output.String(), "Building test-project.installer...")

	//check contents
	unzipPath := testTempRoot + "/unzip"

	assert.NoError(t, global_util.UnzipFile(environment.OutPath("test-project.installer"), unzipPath))

	assert.FileExists(t, filepath.Join(unzipPath, ".package.config"))
	packageConfigContents, _ := ioutil.ReadFile(filepath.Join(unzipPath, ".package.config"))
	assert.Contains(t, string(packageConfigContents), "id: test-project")

	assert.FileExists(t, filepath.Join(unzipPath, "config/package_config.go"))
	assert.FileExists(t, filepath.Join(unzipPath, "ui/ui.go"))
	assert.FileExists(t, filepath.Join(unzipPath, "was-build.sh"))
	assert.FileExists(t, filepath.Join(unzipPath, "example.html"))
	assert.FileExists(t, filepath.Join(unzipPath, "from-download/file.here"))

	assert.FileExists(t, filepath.Join(unzipPath, "data/agent/images/images.untar/manifest.json"))
	assert.FileExists(t, filepath.Join(unzipPath, "data/agent/images/images.untar/repositories"))
	assert.FileExists(t, filepath.Join(unzipPath, "data/agent/images/images.untar/614088555b5b2f43a677aa33c55c7f9ccc5e2bd4d2d88bee4c127ec0f658c3df/json"))
	assert.FileExists(t, filepath.Join(unzipPath, "data/agent/images/images.untar/614088555b5b2f43a677aa33c55c7f9ccc5e2bd4d2d88bee4c127ec0f658c3df/layer.tar.gz"))
	assert.NoFileExists(t, filepath.Join(unzipPath, "data/agent/images/images.untar/614088555b5b2f43a677aa33c55c7f9ccc5e2bd4d2d88bee4c127ec0f658c3df/layer.tar"))

	manifestContent, err := ioutil.ReadFile(filepath.Join(unzipPath, "data/agent/images/images.untar/manifest.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(manifestContent), "\"RepoTags\":[\"traefik:2.3\"]")
	assert.Contains(t, string(manifestContent), "\"RepoTags\":[\"alpine:3.12\"]")
}
