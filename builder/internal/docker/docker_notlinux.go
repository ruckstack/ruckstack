// +build !linux

package docker

func ImageBuild(dockerfile string, tags []string, labels map[string]string) error {
	panic("Can only build on Linux")
}
