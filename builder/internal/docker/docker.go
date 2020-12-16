package docker

import (
	"archive/zip"
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
	"io/ioutil"
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

	containerConfig.Tty = false
	containerConfig.AttachStdin = false
	containerConfig.AttachStdout = true
	containerConfig.AttachStderr = true
	containerConfig.OpenStdin = false

	createdContainer, err := dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		containerName)
	if err != nil {
		return fmt.Errorf("cannot create CLI container: %s", cleanErrorMessage(err))
	}

	if removeWhenDone {
		defer ContainerRemove(createdContainer.ID)
	}

	if err := dockerClient.ContainerStart(ctx, createdContainer.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("cannot start CLI container: %s", cleanErrorMessage(err))
	}

	containerResponse, err := dockerClient.ContainerAttach(ctx, createdContainer.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stdin:  false,
		Stderr: true,
	})

	if err != nil {
		return fmt.Errorf("cannot attach to container: %s", cleanErrorMessage(err))
	}

	_, err = stdcopy.StdCopy(ui.GetOutput(), os.Stderr, containerResponse.Reader)
	if err != nil {
		return fmt.Errorf("error reading docker output: %s", cleanErrorMessage(err))
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

	return nil
}

func SaveImages(outputPath string, imageRefs ...string) error {
	defer ui.StartProgressf("Saving images out of Docker").Stop()

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

func ImageLoad(tarFile *zip.File) error {
	tarStream, err := tarFile.Open()
	if err != nil {
		return err
	}
	reader, err := dockerClient.ImageLoad(context.Background(), tarStream, false)

	if err != nil {
		return cleanErrorMessage(err)
	}

	defer reader.Body.Close()

	sendOutputToUi(reader.Body)

	return nil
}

func GetImageId(ref string) (string, error) {
	inspect, _, err := dockerClient.ImageInspectWithRaw(context.Background(), ref)
	if err != nil {
		if client.IsErrNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return inspect.ID, nil
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

func discardOutput(output io.ReadCloser) {
	termFd, isTerm := term.GetFdInfo(ioutil.Discard)
	jsonmessage.DisplayJSONMessagesStream(output, ioutil.Discard, termFd, isTerm, nil)
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
