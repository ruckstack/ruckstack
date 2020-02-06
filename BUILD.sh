#!/bin/bash

##Ideally this comes from $(out/linux/bin/ruckstack --version)
VERSION=0.5.0

build() {
  echo "Building ruckstack ${VERSION}..."

  echo "Compiling system-control..."
  (export GOOS=linux && go build -o out/system-control cmd/system-control/system-control.go)

  echo "Compiling installer..."
  (export GOOS=linux && go build -o out/installer cmd/installer/installer.go)

  echo "Collecting ruckstack resources..."
  (go-bindata -o internal/ruckstack/builder/resources/bindata/bindata.go -pkg bindata \
          internal/ruckstack/builder/resources/install-dir/... \
          internal/ruckstack/builder/resources/new-project/... \
          out/system-control \
          out/installer \
  )

  echo "Compiling ruckstack..."
  (export GOOS=windows && go build -o out/dist/win/bin/ruckstack.exe cmd/ruckstack/ruckstack.go)
  (export GOOS=linux && go build -o out/dist/linux/ruckstack/bin/ruckstack cmd/ruckstack/ruckstack.go)

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
  rm -rf internal/system-control/resources/bin
  echo "Done"
}

if [ $# -eq 0 ]
then
    build
else
    "$@"
fi
