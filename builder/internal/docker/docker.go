package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/ruckstack/ruckstack/common/ui"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var dockerClient *client.Client

func init() {
	var err error

	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		ui.Fatalf("cannot create docker client: %s", cleanErrorMessage(err))
	}
}

func ImagePull(imageRef string) error {
	if strings.HasPrefix(imageRef, "build.local/") {
		ui.VPrintf("Don't pull local image %s", imageRef)
		return nil
	}

	ui.VPrintf("Pulling %s...", imageRef)
	reader, err := dockerClient.ImagePull(context.Background(), imageRef, types.ImagePullOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "manifest unknown") {
			return fmt.Errorf("Cannot pull image %s. May be an invalid version?", imageRef)
		} else {
			return cleanErrorMessage(err)
		}
	}
	defer reader.Close()

	sendOutputToUi(reader)

	return nil
}

func SaveImages(outputPath string, imageRefs ...string) error {
	defer ui.StartProgressf("Exporting images").Stop()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	ui.VPrintf("exporting %s to %s", strings.Join(imageRefs, ", "), outputPath)
	tarStream, err := dockerClient.ImageSave(context.Background(), imageRefs)
	if err != nil {
		return fmt.Errorf("Error saving images %s: %s", strings.Join(imageRefs, ", "), cleanErrorMessage(err))
	}

	_, err = io.Copy(outputFile, tarStream)
	if err != nil {
		return err
	}

	return nil
}

func cleanErrorMessage(err error) error {
	errMsg := err.Error()
	errMsg = strings.Replace(errMsg, "Error response from daemon: ", "", 1)
	return fmt.Errorf(errMsg)
}

func sendOutputToUi(output io.ReadCloser) {
	termFd, isTerm := term.GetFdInfo(ui.GetOutput())
	jsonmessage.DisplayJSONMessagesStream(output, ui.GetOutput(), termFd, isTerm, nil)
}

func ImageBuild(dockerfile string, tags []string, labels map[string]string) error {
	dockerfile, err := filepath.Abs(dockerfile)
	if err != nil {
		return err
	}

	stat, err := os.Stat(dockerfile)
	if os.IsNotExist(err) {
		return fmt.Errorf("cannot find file %s", dockerfile)
	}
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("dockerfile %s is a directory", dockerfile)
	}

	buildContext, _ := archive.TarWithOptions(filepath.Dir(dockerfile), &archive.TarOptions{})

	resp, err := dockerClient.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		//Version: types.BuilderBuildKit,
		Dockerfile:  filepath.Base(dockerfile),
		Tags:        tags,
		Labels:      labels,
		Remove:      true,
		ForceRemove: true,
	})

	if err != nil {
		return fmt.Errorf("cannot build image: %s", cleanErrorMessage(err))
	}
	defer resp.Body.Close()

	sendOutputToUi(resp.Body)

	return nil
}
