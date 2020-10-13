#!/bin/bash

set -e

##Ideally this comes from $(out/linux/bin/ruckstack --version)
VERSION=0.9.0

build_all() {
  compile
  build_docker
  test

  echo "Done"
}

fast() {
  echo "Building without tests.... Good luck!"
  compile
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
  if [ ! -f tmp/test-installer/example_1.0.5.installer ]; then
    echo "Building test installer package..."
    echo "-- Generating example project in tmp/test-installer/project..."
    mkdir -p tmp/test-installer
    out/artifacts/linux/ruckstack --launch-version local new-project --type example --out tmp/test-installer/project

    echo "-- Building example project in tmp/test-installer/out..."
    out/artifacts/linux/ruckstack --launch-version local build --project tmp/test-installer/project --out tmp/test-installer/out

    echo "-- Extracting to tmp/test-installer/extracted..."
    ADMIN_GROUP=`id -gn`
    tmp/test-installer/out/example_1.0.5.installer --extract-only --install-path tmp/test-installer/extracted --admin-group ${ADMIN_GROUP}
  fi
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
