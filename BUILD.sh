#!/bin/bash

set -e

##Ideally this comes from $(out/linux/bin/ruckstack --version)
VERSION=0.9.0

build_all() {
  compile
  test
  build_docker

  echo "Done"
}

compile() {
  echo "Building ruckstack ${VERSION}..."

  echo "Compiling system-control..."
  (export GOOS=linux && go build -o builder/cli/install_root/resources/install_dir/bin/system-control server/system_control/cmd/main.go)

  echo "Compiling installer..."
  (export GOOS=linux && go build -o builder/cli/install_root/resources/installer server/installer/cmd/main.go)

  echo "Compiling builder..."
  (export GOOS=linux && export CGO_ENABLED=0 && go build -o out/builder_image/bin/ruckstack builder/cli/cmd/main.go)

  echo "Compiling builder launcher..."
  (export GOOS=linux && go build -o out/artifacts/linux/ruckstack builder/launcher/cmd/main.go)
  (export GOOS=windows && go build -o out/artifacts/win/ruckstack.exe builder/launcher/cmd/main.go)
  (export GOOS=darwin && go build -o out/artifacts/mac/ruckstack builder/launcher/cmd/main.go)
  chmod 755 out/artifacts/linux/ruckstack
  chmod 755 out/artifacts/mac/ruckstack

  echo "Creating ruckstack distribution..."
  cp ./LICENSE out/builder_image
  cp -r builder/cli/install_root/* out/builder_image
  chmod 755 out/builder_image/bin/ruckstack
}

test() {
  echo "Running tests..."
  go test ./...
}

build_docker() {
  echo "Building docker image..."
  mkdir -p out/artifacts/docker
  docker build -t ghcr.io/ruckstack/ruckstack:local out/builder_image
}

clean() {
  echo "Cleaning..."
  rm -rf out
  rm -f builder/cli/install_root/resources/install_dir/bin/system-control
  rm -f builder/cli/install_root/resources/installer
  echo "Done"
}

if [ $# -eq 0 ]
then
    build_all
else
    "$@"
fi
