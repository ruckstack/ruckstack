package global_util

import (
	"archive/zip"
	"github.com/ruckstack/ruckstack/common/ui"
	"io"
	"os"
	"path"
	"path/filepath"
)

func Unzip(installPath string, zipReader *zip.ReadCloser) (err error) {
	ui.Print(".....")

	for i, file := range zipReader.File {
		fullname := path.Join(installPath, file.Name)
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

			if i%10 == 0 {
				ui.Print(".")
			}
		}
	}
	return
}