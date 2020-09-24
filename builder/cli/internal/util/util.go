package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func DownloadFile(url string) (string, error) {

	cacheKey := regexp.MustCompile(`https?://.+?/`).ReplaceAllString(url, "")

	savePath := environment.CachePath("download/general/" + cacheKey)

	saveDir, _ := filepath.Split(savePath)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create directory %s: %s", saveDir, err)
	}

	_, err := os.Stat(savePath)
	if err == nil {
		ui.Println(savePath + " already exists. Not re-downloading")
		return savePath, nil
	}

	ui.Println("Downloading " + url + "...")
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot download %s: %s", url, err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("cannot download %s: %s", url, resp.Status)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("cannot create %s: %s", savePath, err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot write %s: %s", savePath, err)
	}

	return savePath, nil
}

func ExtractFromGzip(gzipSource string, wantedFile string) (string, error) {
	gzipFile, err := os.OpenFile(gzipSource, os.O_RDONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("cannot open file to extract: %s", err)
	}

	uncompressedStream, err := gzip.NewReader(gzipFile)
	if err != nil {
		return "", fmt.Errorf("cannot uncompress: %s", err)
	}

	tarReader := tar.NewReader(uncompressedStream)

	savePath := environment.TempPath("gzip-extract-*")

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		switch header.Typeflag {
		case tar.TypeReg:
			if header.Name == wantedFile {
				outFile, err := os.Create(savePath)
				if err != nil {
					return "", err
				}

				defer outFile.Close()
				_, err = io.Copy(outFile, tarReader)
				if err != nil {
					return "", err
				}

				break
			}
		}
	}

	return savePath, nil
}

func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}

func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				ui.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				ui.Println(err)
			}
		}

	}
	return
}
