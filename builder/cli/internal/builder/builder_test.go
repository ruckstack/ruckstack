package builder

import (
	"bytes"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/cli/internal/util"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/stretchr/testify/assert"
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

			assert.NoError(t, util.CopyDir(environment.RuckstackHome+"/builder/cli/install_root/resources/init/example", environment.ProjectDir))

			assert.FileExists(t, environment.RuckstackHome+"/builder/cli/install_root/resources/system-control", "compiled system-control does not exist. Should be created by BUILD.sh")
			err := Build(0)
			if tt.wantErr == "" {
				if assert.NoError(t, err) {
					//check contents
					unzipPath := testRoot + "/unzip"

					assert.NoError(t, global_util.UnzipFile(environment.OutPath("example_1.0.5.installer"), unzipPath))

					assert.FileExists(t, filepath.Join(unzipPath, "config/system.config"))
					assert.FileExists(t, filepath.Join(unzipPath, "bin/example-manager"))
					assert.FileExists(t, filepath.Join(unzipPath, "lib/helm"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/agent/images/images.untar/repositories"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/agent/images/k3s.untar/repositories"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/manifests/frontend.yaml"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/manifests/backend.yaml"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/manifests/postgresql.yaml"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/manifests/traefik.yaml"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/static/charts/frontend.tgz"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/static/charts/backend.tgz"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/server/static/charts/postgresql.tgz"))
					assert.FileExists(t, filepath.Join(unzipPath, "data/web/site-down.html"))

					//installer is executable
					cmd := exec.Command(environment.OutPath("example_1.0.5.installer"))
					var out bytes.Buffer
					cmd.Stdout = &out
					err := cmd.Run()
					assert.Error(t, err) //must be ran as root
					assert.Equal(t, "This installer must be ran as root", strings.TrimSpace(out.String()))
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}
