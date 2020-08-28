package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/shared"
	"io"
	"os"
	"path/filepath"
)

var validate = validator.New()

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func ExpectNoError(err error) {
	if err != nil {
		fmt.Printf("Unexpected error %s", err)
		//panic(err)
		os.Exit(15)
	}
}

func CheckWithMessage(err error, message string, messageParams ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, message, messageParams...)
		fmt.Fprintln(os.Stderr)

		Check(err)
	}
}

func Validate(obj interface{}) error {
	return validate.Struct(obj)
}

func ExtractFromGzip(gzipSource string, wantedFile string, buildEnv *shared.BuildEnvironment) string {
	gzipFile, err := os.OpenFile(gzipSource, os.O_RDONLY, 0644)
	Check(err)

	uncompressedStream, err := gzip.NewReader(gzipFile)
	Check(err)

	tarReader := tar.NewReader(uncompressedStream)
	Check(err)

	savePath := filepath.Join(buildEnv.WorkDir, "extract", wantedFile)
	err = os.MkdirAll(filepath.Dir(savePath), 0755)
	Check(err)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		Check(err)

		switch header.Typeflag {
		case tar.TypeReg:
			if header.Name == wantedFile {
				outFile, err := os.Create(savePath)
				Check(err)

				defer outFile.Close()
				_, err = io.Copy(outFile, tarReader)
				Check(err)

				break
			}
		}
	}

	return savePath
}
