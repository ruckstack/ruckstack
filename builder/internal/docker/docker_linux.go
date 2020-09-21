// +build linux

package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"path/filepath"
)

func ImageBuild(dockerfile string, tags []string, labels map[string]string) error {
	buildContext, _ := archive.TarWithOptions(filepath.Dir(dockerfile), &archive.TarOptions{})

	resp, err := cli.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		Tags:   tags,
		Labels: labels,
	})

	if err != nil {
		return fmt.Errorf("cannot build image: %s", err)
	}
	defer resp.Body.Close()

	sendOutputToUi(resp.Body)

	return nil
}
