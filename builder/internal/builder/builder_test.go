package builder

import (
	"bytes"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/builder/internal/util"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not build projects")
	}
	type args struct {
		projectFile string
		outDir      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can build",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRoot := environment.TempPath("builder_test-*")
			environment.OutDir = testRoot + "/out"
			environment.ProjectDir = testRoot + "/project"
			assert.NoError(t, os.MkdirAll(environment.OutDir, 0755))
			assert.NoError(t, os.MkdirAll(environment.ProjectDir, 0755))

			assert.NoError(t, util.CopyDir(os.DirFS(global_util.GetSourceRoot()+"/builder/internal/bundled/init/example"), environment.ProjectDir))

			assert.FileExists(t, global_util.GetSourceRoot()+"/builder/internal/bundled/system-control", "compiled system-control does not exist. Should be created by BUILD.sh")
			err := Build(0)
			if tt.wantErr == "" {
				if assert.NoError(t, err) {
					//check contents
					unzipPath := testRoot + "/unzip"

					assert.NoError(t, global_util.UnzipFile(environment.OutPath("example_1.0.5.installer"), unzipPath))

					assert.FileExists(t, filepath.Join(unzipPath, ".package.config"))
					assert.FileExists(t, filepath.Join(unzipPath, "config/system.config"))
					assert.FileExists(t, filepath.Join(unzipPath, "bin/example-manager"))
					assert.FileExists(t, filepath.Join(unzipPath, "lib/helm"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/agent/images/images.untar/repositories"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/agent/images/k3s.untar/repositories"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/manifests/frontend.yaml"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/manifests/backend.yaml"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/manifests/postgresql.yaml"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/manifests/traefik-config.yaml"))
					assert.NoFileExists(t, filepath.Join(unzipPath, "data/server/static/charts/frontend.tgz"))   //saved with hash
					assert.NoFileExists(t, filepath.Join(unzipPath, "data/server/static/charts/backend.tgz"))    //saved with hash
					assert.NoFileExists(t, filepath.Join(unzipPath, "data/server/static/charts/postgresql.tgz")) //saved with hash
					assert.FileExists(t, filepath.Join(unzipPath, "data/web/site-down.html"))

					packageContent, err := ioutil.ReadFile(filepath.Join(unzipPath, ".package.config"))
					assert.NoError(t, err)
					assert.True(t, strings.Contains(string(packageContent), "name: Example Project"))

					systemConfigContent, err := ioutil.ReadFile(filepath.Join(unzipPath, "config/system.config"))
					assert.NoError(t, err)
					assert.True(t, strings.Contains(string(systemConfigContent), "serviceName: postgresql"))

					//installer is executable
					cmd := exec.Command(environment.OutPath("example_1.0.5.installer"))
					var out bytes.Buffer
					cmd.Stdout = &out
					err = cmd.Run()
					assert.Error(t, err) //must be ran as root
					assert.Equal(t, "This installer must be ran as root", strings.TrimSpace(out.String()))
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}
