package service

import (
	"bytes"
	"compress/flate"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerfileService_Build(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not build installers")
	}

	output := new(bytes.Buffer)
	ui.SetOutput(output)

	environment.ProjectDir = "."

	type args struct {
		dockerfile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can build",
			args: args{
				dockerfile: "dockerfile_test_complete.txt",
			},
		},
		{
			name:    "Invalid dockerfile",
			wantErr: "cannot build image: Dockerfile parse error line 1: unknown instruction: THIS",
			args: args{
				dockerfile: "dockerfile_test_invalid.txt",
			},
		},
		{
			name:    "Wrong dockerfile",
			wantErr: "cannot find file",
			args: args{
				dockerfile: "dockerfile_test_wrong.txt",
			},
		},
		{
			name:    "Empty dockerfile",
			wantErr: "cannot build image: the Dockerfile (dockerfile_test_empty.txt) cannot be empty",
			args: args{
				dockerfile: "dockerfile_test_empty.txt",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := environment.TempPath("dockerfile-test-*")
			outFile := filepath.Join(testDir, "out.installer")
			assert.NoError(t, os.MkdirAll(testDir, 0755))

			service := &DockerfileService{
				Id:             "test-service",
				Port:           8000,
				ProjectId:      "test-project",
				ProjectVersion: "1.2.3",

				Dockerfile:     tt.args.dockerfile,
				ServiceVersion: "0.5.2",
				BaseUrl:        "/my-url",
			}

			installFile, err := install_file.StartCreation(outFile, flate.BestSpeed)
			assert.NoError(t, err)

			err = service.Build(installFile)

			assert.NoError(t, installFile.CompleteCreation())

			if tt.wantErr != "" {
				if assert.Error(t, err) {
					assert.True(t, strings.HasPrefix(err.Error(), tt.wantErr))
				}
			} else {
				assert.NoError(t, err)

				unzipPath := testDir + "/unzip"
				assert.NoError(t, global_util.UnzipFile(outFile, unzipPath))

				assert.FileExists(t, unzipPath+"/data/server/static/charts/test-service.tgz")

				assert.FileExists(t, unzipPath+"/data/server/manifests/test-service.yaml")

				dockerfileContent, err := ioutil.ReadFile(filepath.Join(unzipPath, "data/agent/images/images.untar/manifest.json"))
				assert.NoError(t, err)
				assert.Contains(t, string(dockerfileContent), "\"RepoTags\":[\"build.local/test-project/test-service:")

			}
		})
	}
}
