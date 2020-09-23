// +build linux

package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"os"
	"path/filepath"
)

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

	resp, err := cli.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		Dockerfile: filepath.Base(dockerfile),
		Tags:       tags,
		Labels:     labels,
	})

	if err != nil {
		return fmt.Errorf("cannot build image: %s", cleanErrorMessage(err))
	}
	defer resp.Body.Close()

	sendOutputToUi(resp.Body)

	return nil
}
