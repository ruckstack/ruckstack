package service

import (
	"bytes"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestHelmService_Build(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not build installers")
	}

	output := new(bytes.Buffer)
	ui.SetOutput(output)
	defer ui.SetOutput(os.Stdout)

	type args struct {
		chart   string
		version string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can build",
			args: args{
				chart:   "stable/tomcat",
				version: "0.4.1",
			},
		},
		{
			name:    "Invalid chart",
			wantErr: "chart \"invalid\" matching 0.4.1 not found in stable index: no chart name found",
			args: args{
				chart:   "stable/invalid",
				version: "0.4.1",
			},
		},
		{
			name:    "Invalid version",
			wantErr: "chart \"tomcat\" matching 99.99.99 not found in stable index: no chart version found for tomcat-99.99.99",
			args: args{
				chart:   "stable/tomcat",
				version: "99.99.99",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := environment.TempPath("helm-test-*")
			outFile := filepath.Join(testDir, "out.installer")
			assert.NoError(t, os.MkdirAll(testDir, 0755))

			service := &HelmService{
				Id:             "test-service",
				Type:           "helm",
				Port:           8000,
				ProjectId:      "test-project",
				ProjectVersion: "1.2.3",
				Chart:          tt.args.chart,
				Version:        tt.args.version,
			}

			installFile, err := install_file.StartCreation(outFile)
			assert.NoError(t, err)

			err = service.Build(installFile)

			assert.NoError(t, installFile.CompleteCreation())

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			} else {
				assert.NoError(t, err)

				unzipPath := testDir + "/unzip"
				assert.NoError(t, global_util.UnzipFile(outFile, unzipPath))

				assert.FileExists(t, unzipPath+"/data/server/static/charts/test-service.tgz")

				assert.FileExists(t, unzipPath+"/data/server/manifests/test-service.yaml")

				savedContents, err := ioutil.ReadFile(unzipPath + "/data/server/manifests/test-service.yaml")
				assert.NoError(t, err)
				assert.Contains(t, string(savedContents), "chart: https://%{KUBERNETES_API}%/static/charts/test-service.tgz")

				helmContent, err := ioutil.ReadFile(filepath.Join(unzipPath, "data/agent/images/images.untar/manifest.json"))
				assert.NoError(t, err)
				assert.Contains(t, string(helmContent), "\"RepoTags\":[\"tomcat:7.0\"]")

			}
		})
	}
}
