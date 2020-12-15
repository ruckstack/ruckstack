package main

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"os"
)

/**
This is an internal program used by the build process to append the packaged container onto the executables.
Called in BUILD.sh
*/
func main() {

	targetFilename := os.Args[1]
	zipContentFilename := os.Args[2]
	imageHash := os.Args[3]

	_, err := os.Stat(targetFilename)
	if err != nil {
		panic(fmt.Sprintf("Cannot stat %s: %s", targetFilename, err))
	}

	targetFile, err := os.OpenFile(targetFilename, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		panic(fmt.Sprintf("Cannot open %s: %s", targetFilename, err))
	}
	defer targetFile.Close()

	zipContentFile, err := os.OpenFile(zipContentFilename, os.O_RDONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("Cannot open %s: %s", zipContentFilename, err))
	}
	defer zipContentFile.Close()

	startOffset, _ := targetFile.Seek(0, io.SeekEnd)

	zipWriter := zip.NewWriter(targetFile)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, 9)
	})
	zipWriter.SetOffset(startOffset)
	defer zipWriter.Close()

	zipFileWriter, err := zipWriter.Create(imageHash)
	if err != nil {
		panic(fmt.Sprintf("Cannot create zip entry: %s", err))
	}

	_, err = io.Copy(zipFileWriter, zipContentFile)
	if err != nil {
		panic(fmt.Sprintf("Cannot write to zip: %s", err))
	}
}
