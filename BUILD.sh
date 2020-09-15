#!/bin/bash

set -e

##Ideally this comes from $(out/linux/bin/ruckstack --version)
VERSION=0.8.3

build_all() {
  compile
  test
  build_docker

  echo "Done"
}

compile() {
  echo "Building ruckstack ${VERSION}..."

  echo "Compiling system-control..."
  (export GOOS=linux && go build -o out/image/system/system_control cmd/system_control/main.go)

  echo "Compiling installer..."
  (export GOOS=linux && go build -o out/image/system/installer cmd/installer/main.go)

  echo "Compiling ruckstack..."
  (export GOOS=linux && export CGO_ENABLED=0 && go build -o out/image/bin/ruckstack cmd/ruckstack/main.go)

  echo "Compiling ruckstack launcher..."
  (export GOOS=linux && go build -o out/artifacts/linux/ruckstack cmd/ruckstack_launcher/main.go)
  (export GOOS=windows && go build -o out/artifacts/win/ruckstack.exe cmd/ruckstack_launcher/main.go)
  (export GOOS=darwin && go build -o out/artifacts/mac/ruckstack cmd/ruckstack_launcher/main.go)
  chmod 755 out/artifacts/linux/ruckstack
  chmod 755 out/artifacts/mac/ruckstack

  echo "Creating ruckstack distribution..."
  cp ./LICENSE out/image
  cp -r dist/* out/image
  chmod 755 out/image/bin/ruckstack
}

test() {
  echo "Running tests..."
  go test ./...
}

build_docker() {
  echo "Building docker image..."
  mkdir -p out/artifacts/docker
  docker build -t ghcr.io/ruckstack/ruckstack:local out/image
}

clean() {
  echo "Cleaning..."
  rm -rf out
  echo "Done"
}

if [ $# -eq 0 ]
then
    build_all
else
    "$@"
fi
