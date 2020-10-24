### Development Notes

## Upgrading k3s

From https://github.com/rancher/k3s/releases find the git hash of the release.

For example, v1.19.3+k3s1 is 974ad30

1. run `go get github.com/rancher/k3s@HASH`

1. Go to the go.mod file at k3s' release tag, and copy the `requires` section to Ruckstack's go.mod

1. Sync go module