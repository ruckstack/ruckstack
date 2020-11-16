package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/ruckstack/ruckstack/common/ui"
	"io"
	"os"
	"strings"
)

var dockerClient *client.Client

func init() {
	var err error

	dockerClient, err = client.NewClientWithOpts(client.FromEnv)
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
func ImageList(options types.ImageListOptions) ([]types.ImageSummary, error) {
	return dockerClient.ImageList(context.Background(), options)
}

func ContainerRun(containerConfig *container.Config, hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig, containerName string, removeWhenDone bool) error {
	ctx := context.Background()

	createdContainer, err := dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		containerName)
	if err != nil {
		return fmt.Errorf("cannot create CLI container: %s", cleanErrorMessage(err))
	}

	if err := dockerClient.ContainerStart(ctx, createdContainer.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("cannot start CLI container: %s", cleanErrorMessage(err))
	}

	if removeWhenDone {
		defer ContainerRemove(createdContainer.ID)
	}

	containerResponse, err := dockerClient.ContainerAttach(ctx, createdContainer.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stdin:  false,
		Stderr: true,
	})

	if err != nil {
		return fmt.Errorf("cannot attach to container: %s", err)
	}

	_, err = stdcopy.StdCopy(ui.GetOutput(), os.Stderr, containerResponse.Reader)
	if err != nil {
		return fmt.Errorf("error reading docker output: %s", err)
	}

	waitOk, errCh := dockerClient.ContainerWait(ctx, createdContainer.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error running CLI container: %s", cleanErrorMessage(err))
		}
	case <-waitOk:
		//ran correctly
	}

	ui.Println("Docker done")

	return nil
}

func SaveImages(outputPath string, imageRefs ...string) error {
	progressMonitor := ui.StartProgressMonitor("Saving images")
	defer progressMonitor.Stop()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}

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

func ImageRemove(imageId string) error {
	_, err := dockerClient.ImageRemove(context.Background(), imageId, types.ImageRemoveOptions{
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("error removing image %s: %s", imageId, cleanErrorMessage(err))
	}

	return err
}

func sendOutputToUi(output io.ReadCloser) {
	termFd, isTerm := term.GetFdInfo(ui.GetOutput())
	jsonmessage.DisplayJSONMessagesStream(output, ui.GetOutput(), termFd, isTerm, nil)
}

func GetContainerId(containerName string) (string, error) {

	listFilters := filters.NewArgs()
	listFilters.Add("name", containerName)

	containerList, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{
		All:     true,
		Filters: listFilters,
	})
	if err != nil {
		return "", fmt.Errorf("error finding container %s: %s", containerName, cleanErrorMessage(err))
	}

	if len(containerList) == 0 {
		return "", nil
	} else if len(containerList) > 1 {
		return "", fmt.Errorf("found %d containers matching %s", len(containerList), containerName)
	}

	return containerList[0].ID, nil
}

func GetImageId(imageRef string) (string, error) {

	listFilters := filters.NewArgs()
	listFilters.Add("reference", imageRef)

	imageList, err := dockerClient.ImageList(context.Background(), types.ImageListOptions{
		All:     true,
		Filters: listFilters,
	})
	if err != nil {
		return "", fmt.Errorf("error finding image %s: %s", imageRef, cleanErrorMessage(err))
	}

	if len(imageList) == 0 {
		return "", nil
	} else if len(imageList) > 1 {
		return "", fmt.Errorf("found %d images matching %s", len(imageList), imageRef)
	}

	return imageList[0].ID, nil
}

func ContainerRemove(containerId string) error {
	err := dockerClient.ContainerRemove(
		context.Background(),
		containerId,
		types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		},
	)

	if err != nil {
		return fmt.Errorf("error cleaning up docker container: %s", cleanErrorMessage(err))
	}

	return nil
}
