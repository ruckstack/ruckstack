package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/global"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var validate = validator.New()

func ExpectNoError(err error) {
	if err != nil {
		fmt.Printf("Unexpected error %s", err)
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
		log.Println(savePath + " already exists. Not re-downloading")
		return savePath, nil
	}

	log.Println("Downloading " + url + "...")
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
