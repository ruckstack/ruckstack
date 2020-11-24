package global_util

import (
	"archive/zip"
	"github.com/ruckstack/ruckstack/common/ui"
	"io"
	"os"
	"path"
	"path/filepath"
)

func UnzipFile(zipFile string, outputDir string) (err error) {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	return Unzip(zipReader, outputDir)
}

func Unzip(zipContent *zip.ReadCloser, outputDir string) (err error) {
	ui.Println(".....")

	for _, file := range zipContent.File {
		fullname := path.Join(outputDir, file.Name)
		fileInfo := file.FileInfo()
		if fileInfo.IsDir() {
			os.MkdirAll(fullname, fileInfo.Mode().Perm())
		} else {
			_, err := os.Stat(fullname)
			if err == nil {
				os.Remove(fullname)
			}

			os.MkdirAll(filepath.Dir(fullname), 0755)
			perms := fileInfo.Mode().Perm()
			out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
			if err != nil {
				return err
			}
			rc, err := file.Open()
			if err != nil {
				return err
			}
			_, err = io.CopyN(out, rc, fileInfo.Size())
			if err != nil {
				return err
			}
			rc.Close()
			out.Close()

			mtime := fileInfo.ModTime()
			if err := os.Chtimes(fullname, mtime, mtime); err != nil {
				return err
			}
		}
	}
	return
}
