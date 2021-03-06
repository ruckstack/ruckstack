package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func DownloadFile(url string) (string, error) {

	cacheKey := regexp.MustCompile(`https?://`).ReplaceAllString(url, "")

	savePath := environment.CachePath("download/general/" + cacheKey)
	_, err := os.Stat(savePath)
	if err == nil {
		ui.VPrintf("Packaged file %s", filepath.Base(savePath))
		return savePath, nil
	} else {
		if os.IsNotExist(err) {
			//ok, will download
		} else {
			return "", err
		}

	}

	saveDir, _ := filepath.Split(savePath)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create directory %s: %s", saveDir, err)
	}

	_, err = os.Stat(savePath)
	if err == nil {
		ui.VPrintf("Already downloaded %s to %s", filepath.Base(savePath), savePath)
		return savePath, nil
	}

	defer ui.StartProgressf("Downloading %s", url).Stop()
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

	foundFile := false
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

				foundFile = true
				break
			}
		}
	}

	if !foundFile {
		return "", fmt.Errorf("cannot find %s in %s", wantedFile, gzipSource)
	}
	return savePath, nil
}

func CopyFile(sourcefile fs.File, dest string) (err error) {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	//if err == nil {
	//	sourceinfo, err := os.Stat(source)
	//	if err != nil {
	//		err = os.Chmod(dest, sourceinfo.Mode())
	//	}
	//
	//}

	return
}

func CopyDir(source fs.FS, dest string) (err error) {

	return fs.WalkDir(source, ".", func(path string, obj fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		destinationfilepointer := dest + "/" + path

		if obj.IsDir() {
			err = os.MkdirAll(dest, 0755)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			fileObj, err := source.Open(path)
			err = CopyFile(fileObj, destinationfilepointer)
			if err != nil {
				return err
			}
			_ = fileObj.Close()
		}

		return nil
	})
}
