#!/bin/bash

set -e

export RUCKSTACK_ANALYTICS=false

##Ideally this comes from $(out/linux/bin/ruckstack --version)
VERSION=1.1.1
export RUCKSTACK_WORK_DIR="$(pwd)/tmp/work_dir"

full_build() {
  clean
  compile_ops
  compile_go
  test
  build_artifacts
}

fast() {
  compile_go
  build_artifacts
}

compile_ops() {
  echo "Building /ops..."
  (cd ops-dashboard && npm run-script build)
}

compile_go() {
  echo "Building ruckstack ${VERSION}..."

  echo "Compiling system-control..."
  (export GOOS=linux && go build -o builder/internal/bundled/system-control server/system_control/cmd/main.go)

  echo "Compiling installer..."
  (export GOOS=linux && go build -o builder/internal/bundled/installer installer/cmd/main.go)

  echo "Compiling builder..."
  echo "Compiling builder...Linux..."
  (export GOOS=linux && export CGO_ENABLED=0 && go build -o out/artifacts/linux/ruckstack builder/cmd/main.go)

  echo "Compiling builder...Windows..."
  (export GOOS=windows && go build -o out/artifacts/win/ruckstack.exe builder/cmd/main.go)

  echo "Compiling builder...Mac..."
  (export GOOS=darwin && go build -o out/artifacts/mac/ruckstack builder/cmd/main.go)

  chmod 755 out/artifacts/linux/ruckstack
  chmod 755 out/artifacts/mac/ruckstack
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

build_artifacts() {
  echo "Building release archives..."
  (cd out/artifacts/linux && tar -czf ruckstack-linux-${VERSION}.tar.gz ruckstack)
  (cd out/artifacts/mac && tar -czf ruckstack-mac-${VERSION}.tar.gz ruckstack)
  (cd out/artifacts/win && zip -q ruckstack-windows-${VERSION}.zip ruckstack.exe)
}

build_docker() {
  echo "Building docker image '${1}'..."
  mkdir -p out/artifacts/docker
  docker build -t ghcr.io/ruckstack/ruckstack:${1} .
  docker save ghcr.io/ruckstack/ruckstack:${1} --output out/artifacts/docker/ruckstack.image.tar
}

push_docker() {
  docker push ghcr.io/ruckstack/ruckstack:${1}
}

clean() {
  echo "Cleaning..."
  rm -rf out

  rm -f builder/internal/bundled/system-control
  rm -f builder/internal/bundled/installer
  rm -rf server/system_control/internal/server/webserver/ops
  echo "Done"
}

if [ $# -eq 0 ]; then
  full_build
else
  "$@"
fi
