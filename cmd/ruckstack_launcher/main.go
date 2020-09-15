package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/ruckstack/ruckstack/internal/ruckstack/ui"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {

	containerConfig := &container.Config{
		Tty: false,
	}

	hostConfig := &container.HostConfig{}

	parsedArgs, updatedArgs, env, mountPoints := processArguments(os.Args[1:])
	containerConfig.Cmd = updatedArgs
	containerConfig.Env = env
	hostConfig.Mounts = mountPoints

	_, verbose := parsedArgs["--verbose"]
	if verbose {
		ui.SetVerbose(verbose)
	}

	containerConfig.Env = append(containerConfig.Env, "RUCKSTACK_DOCKERIZED=true")

	launchUser := parsedArgs["--launch-user"]
	if launchUser != "" {
		containerConfig.User = launchUser
	}

	useVersion := parsedArgs["--launch-version"]
	if useVersion == "" {
		useVersion = "latest"
	}

	if regexp.MustCompile("^\\d.*").MatchString(useVersion) {
		useVersion = "v" + useVersion
	}

	imageName := parsedArgs["--launch-image"]
	if imageName == "" {
		imageName = "ghcr.io/ruckstack/ruckstack"
	}

	_, forcePull := parsedArgs["--launch-force-pull"]

	containerConfig.Image = fmt.Sprintf("%s:%s", imageName, useVersion)

	ui.VPrintf("LAUNCHER: Force Pull: %t", forcePull)
	ui.VPrintf("LAUNCHER: Image: %s", containerConfig.Image)
	ui.VPrintf("LAUNCHER: User: %s", containerConfig.User)
	ui.VPrintf("LAUNCHER: Command: %s", containerConfig.Cmd)
	ui.VPrintf("LAUNCHER: Environment:\n    %s", strings.Join(env, "\n    "))

	for _, mountPt := range hostConfig.Mounts {
		ui.VPrintf("LAUNCHER: Mount Point %s -> %s\n", mountPt.Source, mountPt.Target)
	}

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		exitWithError(fmt.Errorf("cannot create docker client: %s", err))
	}

	if forcePull {
		pullImage(cli, ctx, containerConfig)
	} else {
		listFilters := filters.NewArgs()
		listFilters.Add("reference", containerConfig.Image)
		imageList, err := cli.ImageList(ctx, types.ImageListOptions{
			Filters: listFilters,
		})
		if err != nil {
			exitWithError(err)
		}

		if len(imageList) == 0 {
			ui.VPrintf("No local images found for %s", containerConfig.Image)

			pullImage(cli, ctx, containerConfig)
		} else {
			ui.VPrintf("Not pulling image %s: already in local image cache", containerConfig.Image)
		}
	}

	resp, err := cli.ContainerCreate(ctx,
		containerConfig,
		hostConfig,
		nil, "ruckstack-run")
	if err != nil {
		exitWithError(fmt.Errorf("cannot create CLI container: %s", err))
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		exitWithError(fmt.Errorf("cannot start CLI container: %s", err))
	}

	defer cleanup(cli, resp)

	waitOk, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			exitWithError(fmt.Errorf("error running CLI container: %s", err))
		}
	case <-waitOk:
		//ran correctly
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		exitWithError(fmt.Errorf("error getting container logs: %s", err))
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func pullImage(cli *client.Client, ctx context.Context, containerConfig *container.Config) {
	ui.VPrintf("Pulling %s...", containerConfig.Image)

	reader, err := cli.ImagePull(ctx, containerConfig.Image, types.ImagePullOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "manifest unknown") {
			exitWithError(fmt.Errorf("Cannot pull image %s. May be an invalid version?", containerConfig.Image))
		} else {
			exitWithError(err)
		}
	}
	defer reader.Close()

	termFd, isTerm := term.GetFdInfo(os.Stdout)
	_ = jsonmessage.DisplayJSONMessagesStream(reader, os.Stdout, termFd, isTerm, nil)
}

func exitWithError(err error) {
	if strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
		ui.VPrintln("LAUNCHER:", err.Error())
		ui.Println("Error launching Ruckstack: Ruckstack requires Docker to run. Please install and/or start the Docker daemon process and re-run Ruckstack")
	} else {
		ui.Println("Error launching Ruckstack:", err)
	}
	os.Exit(-1)
}

/**
Takes the original process args and replaces any that need to have different value for Docker
and stores the original value in an array for use in environment variables
*/
func processArguments(originalArgs []string) (map[string]string, []string, []string, []mount.Mount) {
	parsedArgs := make(map[string]string)
	envVariables := make([]string, 0)
	newArgs := make([]string, len(originalArgs))
	mountPoints := make([]mount.Mount, 0)

	for i := 0; i < len(originalArgs); i++ {
		newArgs[i] = originalArgs[i]

		thisArg := originalArgs[i]
		if strings.HasPrefix(thisArg, "-") {
			value := ""
			if i+1 < len(originalArgs) {
				value = originalArgs[i+1]
				if strings.HasPrefix(value, "-") {
					//next flag is an argument
					value = ""
				}
			}

			parsedArgs[thisArg] = value
		}

		if i == 0 {
			//can't check for the flag name on the first arg
			continue
		}
		possibleArg := originalArgs[i-1]
		switch possibleArg {
		case "--out":
			newArgs[i] = "/data/out"
			abs, err := filepath.Abs(originalArgs[i])
			if err != nil {
				exitWithError(err)
			}
			envVariables = append(envVariables, fmt.Sprintf("WRAPPED_OUT=%s", originalArgs[i]))
			envVariables = append(envVariables, fmt.Sprintf("WRAPPED_OUT_ABS=%s", abs))

			sourcePath, _ := filepath.Abs(originalArgs[i])

			err = os.MkdirAll(sourcePath, 0755)
			if err != nil {
				exitWithError(fmt.Errorf("cannot create directory %s: %s", sourcePath, err))
			}
			mountPoints = append(mountPoints, mount.Mount{
				Type:     mount.TypeBind,
				Source:   sourcePath,
				Target:   "/data/out",
				ReadOnly: false,
			})
		case "--project":
			newArgs[i] = "/data/project"

			abs, err := filepath.Abs(originalArgs[i])
			if err != nil {
				exitWithError(err)
			}
			envVariables = append(envVariables, fmt.Sprintf("WRAPPED_PROJECT=%s", originalArgs[i]))
			envVariables = append(envVariables, fmt.Sprintf("WRAPPED_PROJECT_ABS=%s", abs))

			sourcePath, _ := filepath.Abs(originalArgs[i])

			err = os.MkdirAll(sourcePath, 0755)
			if err != nil {
				exitWithError(fmt.Errorf("cannot create directory %s: %s", sourcePath, err))
			}
			mountPoints = append(mountPoints, mount.Mount{
				Type:     mount.TypeBind,
				Source:   sourcePath,
				Target:   "/data/project",
				ReadOnly: false,
			})

		}
	}

	return parsedArgs, newArgs, envVariables, mountPoints
}

func cleanup(cli *client.Client, resp container.ContainerCreateCreatedBody) {
	if err := cli.ContainerRemove(
		context.Background(),
		resp.ID,
		types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		},
	); err != nil {
		exitWithError(fmt.Errorf("error cleaning up docker container: %s", err))
	}
}
