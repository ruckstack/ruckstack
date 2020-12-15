#!/bin/bash

set -e

##Ideally this comes from $(out/linux/bin/ruckstack --version)
VERSION=0.10.0
export RUCKSTACK_TEMP_DIR="$(pwd)/tmp"
export RUCKSTACK_CACHE_DIR="$(pwd)/cache"

build_all() {
  fast
  test
}

fast() {
  compile
  build_docker
  finish_artifacts
}

compile() {
  echo "Building ruckstack ${VERSION}..."

  echo "Compiling system-control..."
  (export GOOS=linux && go build -o builder/cli/install_root/resources/system-control server/system_control/cmd/main.go)

  echo "Compiling installer..."
  (export GOOS=linux && go build -o builder/cli/install_root/resources/installer installer/cmd/main.go)

  echo "Compiling builder..."
  (export GOOS=linux && export CGO_ENABLED=0 && go build -o out/builder_image/bin/ruckstack builder/cli/cmd/main.go)

  echo "Compiling builder launcher..."
  rm -rf out/artifacts/linux
  rm -rf out/artifacts/win
  rm -rf out/artifacts/mac

  (export GOOS=linux && go build -o out/artifacts/linux/ruckstack.launcher builder/launcher/cmd/main.go)
  (export GOOS=windows && go build -o out/artifacts/win/ruckstack.launcher.exe builder/launcher/cmd/main.go)
  (export GOOS=darwin && go build -o out/artifacts/mac/ruckstack.launcher builder/launcher/cmd/main.go)
  chmod 755 out/artifacts/linux/ruckstack.launcher
  chmod 755 out/artifacts/mac/ruckstack.launcher

  echo "Creating ruckstack distribution..."
  cp ./LICENSE out/builder_image
  cp -r builder/cli/install_root/* out/builder_image
  chmod 755 out/builder_image/bin/ruckstack
  rm -rf out/builder_image/tmp
  rm -rf out/builder_image/resources/cache/helm

  echo "Compiling file_join..."
  (export GOOS=linux && go build -o tmp/build_utils/file_join build_utils/file_join/file_join.go)
}

test() {
  if [ ! -f tmp/test-installer/out/example_1.0.5.installer ]; then
    echo "Building test installer package..."
    echo "-- Generating example project in tmp/test-installer/project..."
    mkdir -p tmp/test-installer
    out/artifacts/linux/ruckstack init --template example --out tmp/test-installer/project

    echo "-- Building example project in tmp/test-installer/out..."
    out/artifacts/linux/ruckstack build --project tmp/test-installer/project --out tmp/test-installer/out

    echo "-- Extracting to tmp/test-installer/extracted..."
    ADMIN_GROUP=$(id -gn)
    tmp/test-installer/out/example_1.0.5.installer --extract-only --install-path tmp/test-installer/extracted --admin-group ${ADMIN_GROUP}
  fi
  echo "Running tests..."
  go test ./...
}

build_docker() {
  echo "Building docker image..."
  mkdir -p out/artifacts/docker
  docker build -t ruckstack/ruckstack:v${VERSION} out/builder_image
  docker save ruckstack/ruckstack:v${VERSION} --output out/artifacts/docker/ruckstack.image.tar
}

finish_artifacts() {
  echo "Appending packaged containers to launcher..."
  cp out/artifacts/linux/ruckstack.launcher out/artifacts/linux/ruckstack
  cp out/artifacts/win/ruckstack.launcher.exe out/artifacts/win/ruckstack.exe
  cp out/artifacts/mac/ruckstack.launcher out/artifacts/mac/ruckstack

  tmp/build_utils/file_join out/artifacts/linux/ruckstack out/artifacts/docker/ruckstack.image.tar $(docker image inspect --format "{{.Id}}"  ruckstack/ruckstack:v${VERSION})
  tmp/build_utils/file_join out/artifacts/win/ruckstack.exe out/artifacts/docker/ruckstack.image.tar $(docker image inspect --format "{{.Id}}"  ruckstack/ruckstack:v${VERSION})
  tmp/build_utils/file_join out/artifacts/mac/ruckstack out/artifacts/docker/ruckstack.image.tar $(docker image inspect --format "{{.Id}}"  ruckstack/ruckstack:v${VERSION})

  chmod 755 out/artifacts/linux/ruckstack
  chmod 755 out/artifacts/mac/ruckstack

  echo "Building release archives..."
  (cd out/artifacts/linux && tar -czf ruckstack-linux-${VERSION}.tar.gz ruckstack)
  (cd out/artifacts/mac && tar -czf ruckstack-mac-${VERSION}.tar.gz ruckstack)
  (cd out/artifacts/win && zip -q ruckstack-windows-${VERSION}.zip ruckstack.exe)
}

push_docker() {
  docker tag ruckstack/ruckstack:v${VERSION} ghcr.io/ruckstack/ruckstack:${1}
  docker push ghcr.io/ruckstack/ruckstack:${1}
}

clean() {
  echo "Cleaning..."
  rm -rf out
  rm -f builder/cli/install_root/resources/system-control
  rm -f builder/cli/install_root/resources/installer
  echo "Done"
}

if [ $# -eq 0 ]; then
  build_all
else
  "$@"
fi
