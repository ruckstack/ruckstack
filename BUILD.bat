@echo off
if "%OS%" == "Windows_NT" setlocal

setlocal enabledelayedexpansion

REM ##Ideally this comes from $(out/linux/bin/ruckstack --version)
set VERSION=0.8.1

echo "Building ruckstack %VERSION%..."

echo "Compiling system-control..."
set GOOS=linux
go build -o out/system-control cmd/system-control/system-control.go

echo "Compiling installer..."
set GOOS=linux
go build -o out/installer cmd/installer/installer.go

echo "Collecting ruckstack resources..."
go-bindata -o internal/ruckstack/builder/resources/bindata/bindata.go -pkg bindata ^
         internal/ruckstack/builder/resources/install-dir/... ^
         internal/ruckstack/builder/resources/new-project/... ^
         out/system-control ^
         out/installer


echo "Compiling ruckstack..."
set GOOS=windows
go build -o out/dist/win/bin/ruckstack.exe cmd/ruckstack/ruckstack.go

set GOOS=linux
go build -o out/dist/linux/ruckstack/bin/ruckstack cmd/ruckstack/ruckstack.go

echo "Done"
