package docker

import (
	"bytes"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/test_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	alpineImage = "alpine:3.12"
)

func TestImagePull(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not pull docker images")
	}
	output := new(bytes.Buffer)
	ui.SetOutput(output)

	type args struct {
		imageRef string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Can pull image",
			args: args{
				imageRef: alpineImage,
			},
		},
		{
			name:    "Error on invalid image",
			wantErr: true,
			args: args{
				imageRef: "invalid:3.12",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ImagePull(tt.args.imageRef)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output.String(), "Pulling from library/alpine")
			}

		})
	}
}

func TestImageList(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not pull docker images")
	}

	//make sure we've pulled at least one image
	err := ImagePull(alpineImage)
	assert.NoError(t, err)

	images, err := ImageList(types.ImageListOptions{})
	assert.NoError(t, err)
	assert.Greater(t, len(images), 0)
}

func TestRunContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not run docker images")
	}

	containerName := "test-container"
	containerId, err := GetContainerId(containerName)
	assert.NoError(t, err)
	if containerId != "" {
		assert.NoError(t, ContainerRemove(containerId))
	}

	output := new(bytes.Buffer)
	ui.SetOutput(output)

	type args struct {
		removeWhenDone bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Can run command and remove container",
			args: args{
				removeWhenDone: true,
			},
		},
		{
			name: "Can run command and keep container",
			args: args{
				removeWhenDone: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ContainerRun(
				&container.Config{
					Image: alpineImage,
					Cmd:   []string{"echo", "Test Passed"},
					Tty:   false,
				},
				nil,
				nil,
				containerName,
				tt.args.removeWhenDone,
			)
			assert.NoError(t, err)
			assert.Contains(t, output.String(), "Test Passed")

			containerId, err = GetContainerId(containerName)
			assert.NoError(t, err)

			if tt.args.removeWhenDone {
				assert.Empty(t, containerId)
			} else {
				assert.NotEmpty(t, containerId)
				assert.NoError(t, ContainerRemove(containerId))
			}
		})
	}
}

func TestSaveImages(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not run docker images")
	}

	output := new(bytes.Buffer)
	ui.SetOutput(output)

	outputPath := test_util.TempPath("test_docker_save-*.tar")
	type args struct {
		outputPath string
		imageRefs  []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can save image",
			args: args{
				outputPath: outputPath,
				imageRefs:  []string{"alpine:3.12"},
			},
		},
		{
			name:    "Error on invalid image",
			wantErr: "Error saving images alpine:3.12, invalid:1.3: reference does not exist",
			args: args{
				outputPath: outputPath,
				imageRefs:  []string{"alpine:3.12", "invalid:1.3"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveImages(tt.args.outputPath, tt.args.imageRefs...)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.FileExists(t, outputPath)
			}
		})
	}
}

func TestImageRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not run docker images")
	}
	output := new(bytes.Buffer)
	ui.SetOutput(output)

	//use different image so we can safely delete it
	alpineImageToDelete := "alpine:3.11"
	assert.NoError(t, ImagePull(alpineImageToDelete))

	imageId, err := GetImageId(alpineImageToDelete)
	assert.NoError(t, err)

	type args struct {
		imageId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Can delete image",
			args: args{
				imageId: imageId,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ImageRemove(tt.args.imageId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				imageId, err = GetImageId(alpineImageToDelete)
				assert.NoError(t, err)
				assert.Empty(t, imageId)

			}
		})
	}
}

func TestImageBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not run docker images")
	}
	output := new(bytes.Buffer)
	ui.SetOutput(output)

	dockerfile := global_util.GetSourceRoot() + "/builder/cli/install_root/Dockerfile"

	type args struct {
		dockerfile string
		tags       []string
		labels     map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Can build",
			args: args{
				dockerfile: dockerfile,
				tags:       []string{"ruckstack:test-built"},
				labels:     map[string]string{"ruckstack.test": "true"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ImageBuild(tt.args.dockerfile, tt.args.tags, tt.args.labels)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output.String(), "Successfully tagged ruckstack:test-built")
			}

		})
	}
}
