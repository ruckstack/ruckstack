package main

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/ruckstack/ruckstack/builder/internal/argwrapper"
	"github.com/ruckstack/ruckstack/builder/internal/docker"
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	//controls if project/out/etc. directories are auto-crated. Purely for testing purposes.
	autoMkDirs = true

	/**
	Copy of what is in builder's RootCommand so we know default mounted directories to pass along.
	Ideally there would be a test to make sure this is kept in sync.
	We can't directly depend on RootCommand or it compiles in all of the builder into the launcher
	*/
	commandDefaults = map[string]map[string]string{
		"new-project": {
			"out": ".",
		},
		"build": {
			"out":     ".",
			"project": ".",
		},
	}
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
		containerConfig.Env = append(containerConfig.Env, "RUCKSTACK_VERBOSE=true")
	}

	containerConfig.Env = append(containerConfig.Env, "RUCKSTACK_DOCKERIZED=true")

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
	ui.VPrintf("LAUNCHER: Environment:\n    %s", strings.Join(containerConfig.Env, "\n    "))

	for _, mountPt := range hostConfig.Mounts {
		ui.VPrintf("LAUNCHER: Mount Point %s -> %s\n", mountPt.Source, mountPt.Target)
	}

	if forcePull {
		ui.VPrintf("Forced pull No local images found for %s", containerConfig.Image)
		if err := docker.ImagePull(containerConfig.Image); err != nil {
			exitWithError(err)
		}

	} else {
		listFilters := filters.NewArgs()
		listFilters.Add("reference", containerConfig.Image)
		imageList, err := docker.ImageList(types.ImageListOptions{
			Filters: listFilters,
		})
		if err != nil {
			exitWithError(err)
		}

		if len(imageList) == 0 {
			ui.VPrintf("No local images found for %s", containerConfig.Image)
			docker.ImagePull(containerConfig.Image)
		} else {
			ui.VPrintf("Not pulling image %s: already in local image cache", containerConfig.Image)
		}
	}

	if err := docker.ContainerRun(containerConfig, hostConfig, nil, "ruckstack-run", true); err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	errorMessage := fmt.Sprintf("Error launching Ruckstack: %s", err)

	if strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
		ui.VPrintln("LAUNCHER:", err.Error())
		errorMessage = "Error launching Ruckstack: Ruckstack requires Docker to run. Please install and/or start the Docker daemon process and re-run Ruckstack"
	} else {
		errorMessage = "Error launching Ruckstack:" + err.Error()
	}
	ui.Fatal(errorMessage)
}

/**
Takes the original process args and replaces any that need to have different value for Docker
and stores the original value in an array for use in environment variables
*/
func processArguments(originalArgs []string) (map[string]string, []string, []string, []mount.Mount) {

	commandDefaults := getCommandDefaults(originalArgs)
	for commandArg, defaultValue := range commandDefaults {
		foundArg := false
		for _, passedArg := range originalArgs {
			if passedArg == "--"+commandArg {
				foundArg = true
			}
		}
		if !foundArg {
			originalArgs = append(originalArgs, "--"+commandArg, defaultValue)
		}
	}

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
			envVariables = argwrapper.SaveOriginalValue("out", originalArgs[i], envVariables)

			sourcePath, _ := filepath.Abs(originalArgs[i])

			if autoMkDirs {
				if err := os.MkdirAll(sourcePath, 0755); err != nil {
					exitWithError(fmt.Errorf("cannot create directory %s: %s", sourcePath, err))
				}
			}
			mountPoints = append(mountPoints, mount.Mount{
				Type:     mount.TypeBind,
				Source:   sourcePath,
				Target:   "/data/out",
				ReadOnly: false,
			})
		case "--project":
			newArgs[i] = "/data/project"

			envVariables = argwrapper.SaveOriginalValue("project", originalArgs[i], envVariables)

			sourcePath, _ := filepath.Abs(originalArgs[i])

			if autoMkDirs {
				if err := os.MkdirAll(sourcePath, 0755); err != nil {
					exitWithError(fmt.Errorf("cannot create directory %s: %s", sourcePath, err))
				}
			}
			mountPoints = append(mountPoints, mount.Mount{
				Type:     mount.TypeBind,
				Source:   sourcePath,
				Target:   "/data/project",
				ReadOnly: false,
			})

		}
	}

	currentUser, err := user.Current()
	if err != nil {
		ui.Fatalf("Error getting current user: %s", err)
	}
	localCacheDir := filepath.Join(currentUser.HomeDir, ".ruckstack")
	_ = os.MkdirAll(localCacheDir, 0755)
	mountPoints = append(mountPoints, mount.Mount{
		Type:     mount.TypeBind,
		Source:   localCacheDir,
		Target:   "/data/cache",
		ReadOnly: false,
	})

	return parsedArgs, newArgs, envVariables, mountPoints
}

func getCommandDefaults(args []string) map[string]string {
	for _, arg := range args {
		commandDefaults, found := commandDefaults[arg]
		if found {
			return commandDefaults
		}
	}
	return nil
}
