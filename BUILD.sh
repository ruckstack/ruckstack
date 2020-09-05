#!/bin/bash

set -e

##Ideally this comes from $(out/linux/bin/ruckstack --version)
VERSION=0.8.3

build() {
  echo "Building ruckstack ${VERSION}..."

  echo "Compiling system-control..."
  (export GOOS=linux && go build -o out/system_control cmd/system_control/main.go)

  echo "Compiling installer..."
  (export GOOS=linux && go build -o out/installer cmd/installer/main.go)

  echo "Collecting ruckstack resources..."
  (go-bindata -o internal/ruckstack/builder/resources/bindata/bindata.go -pkg bindata \
          internal/ruckstack/builder/resources/install_dir/... \
          internal/ruckstack/builder/resources/new_project/... \
          out/system_control \
          out/installer \
  )

  echo "Compiling ruckstack..."
  (export GOOS=windows && go build -o out/dist/win/bin/ruckstack.exe cmd/ruckstack/main.go)
  (export GOOS=linux && go build -o out/dist/linux/ruckstack/bin/ruckstack cmd/ruckstack/main.go)

  echo "Creating windows distribution..."
  cp ./LICENSE out/dist/win
  cp -r dist out/dist/win
  (cd out/dist/win && zip -q -r ../../ruckstack-win-${VERSION}.zip *)

  echo "Creating linux distribution..."
  cp ./LICENSE out/dist/linux/ruckstack
  cp -r dist out/dist/linux/ruckstack
  chmod 755 out/dist/linux/ruckstack/bin/ruckstack
  (cd out/dist/linux && tar cfz ../../ruckstack-linux-${VERSION}.tgz ruckstack)


  echo "Done"
}

clean() {
  echo "Cleaning..."
  rm -rf out
  rm -rf internal/ruckstack/resources/bindata
  rm -rf internal/system_control/resources/bin
  echo "Done"
}

if [ $# -eq 0 ]
then
    build
else
    "$@"
fi
