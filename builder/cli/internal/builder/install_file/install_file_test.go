package install_file

import (
	"archive/zip"
	"bytes"
	"github.com/ruckstack/ruckstack/builder/cli/internal/project"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreatingInstallFile(t *testing.T) {
	output := new(bytes.Buffer)
	ui.SetOutput(output)
	defer ui.SetOutput(os.Stdout)

	environment.OutDir = environment.TempPath("out")
	_ = os.MkdirAll(environment.OutDir, 0755)

	assert.FileExists(t, environment.RuckstackHome+"/builder/cli/install_root/resources/installer", "compiled installer does not exist. Should be created by BUILD.sh")
	assert.FileExists(t, environment.RuckstackHome+"/builder/cli/install_root/resources/system-control", "compiled system-control does not exist. Should be created by BUILD.sh")

	projectConfig := project.ProjectConfig{
		Id:      "test-project",
		Name:    "Test Project",
		Version: "1.2.3",
	}

	installFile, err := StartInstallFile(&projectConfig)
	assert.NoError(t, err)

	assert.Regexp(t, "tmp/out/test-project_1.2.3.installer$", installFile.file.Name())
	assert.FileExists(t, installFile.file.Name())

	assert.Equal(t, "test-project", installFile.PackageConfig.Id)
	assert.Equal(t, "Test Project", installFile.PackageConfig.Name)
	assert.Equal(t, "1.2.3", installFile.PackageConfig.Version)

	//add some content
	assert.Nil(t, installFile.AddDirectory(environment.RuckstackHome+"/common", ""))
	assert.Nil(t, installFile.AddFile(environment.RuckstackHome+"/BUILD.sh", "was-build.sh"))
	assert.Nil(t, installFile.AddDownloadedFile("http://example.com/index.html", "example.html"))
	assert.Nil(t, installFile.AddDownloadedNestedFile("https://get.helm.sh/helm-v3.3.3-linux-amd64.tar.gz", "linux-amd64/README.md", "from-download/file.here"))
	assert.Nil(t, installFile.AddDockerImage("alpine:3.12"))

	//now build
	err = installFile.Build()
	assert.NoError(t, err)

	assert.Contains(t, output.String(), "Building test-project_1.2.3.installer...")

	//check contents
	unzipPath := environment.TempPath("unzipped")
	zipReader, err := zip.OpenReader(environment.OutPath("test-project_1.2.3.installer"))
	assert.NoError(t, err)
	defer zipReader.Close()

	err = global_util.Unzip(unzipPath, zipReader)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(unzipPath, ".package.config"))
	packageConfigContents, _ := ioutil.ReadFile(filepath.Join(unzipPath, ".package.config"))
	assert.Contains(t, string(packageConfigContents), "id: test-project")

	assert.FileExists(t, filepath.Join(unzipPath, "config/config.go"))
	assert.FileExists(t, filepath.Join(unzipPath, "ui/ui.go"))
	assert.FileExists(t, filepath.Join(unzipPath, "was-build.sh"))
	assert.FileExists(t, filepath.Join(unzipPath, "example.html"))
	assert.FileExists(t, filepath.Join(unzipPath, "from-download/file.here"))

}
