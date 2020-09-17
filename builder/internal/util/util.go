package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/builder/internal/builder/global"
	"github.com/ruckstack/ruckstack/common/ui"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	validate = validator.New()
)

func ExpectNoError(err error) {
	if err != nil {
		ui.Printf("Unexpected error %s", err)
		//panic(err)
		os.Exit(15)
	}
}

func Validate(obj interface{}) error {
	return validate.Struct(obj)
}

func DownloadFile(url string) (string, error) {

	cacheKey := regexp.MustCompile(`https?://.+?/`).ReplaceAllString(url, "")

	savePath := filepath.Join(global.BuildEnvironment.CacheDir, cacheKey)

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

	savePath, err := ioutil.TempDir(filepath.Join(global.BuildEnvironment.WorkDir), "extract")
	if err != nil {
		return "", err
	}

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

/**
Check for a WRAPPED_* environment variable that was set by ruckstack wrapper and return that if it was set.
Otherwise, return the nonWrappedValue
*/
func WrappedValue(name string, nonWrappedValue string) string {
	env := os.Getenv("WRAPPED_" + strings.ToUpper(name))
	if env == "" {
		return nonWrappedValue
	} else {
		return env
	}
}

/**
Returns true if ruckstack is running via the launcher
*/
func IsRunningLauncher() bool {
	return os.Getenv("RUCKSTACK_DOCKERIZED") == "true"
}
