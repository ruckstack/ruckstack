package global_util

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func TarDirectory(sourceDir string, targetFilename string, compress bool) error {
	targetFile, err := os.OpenFile(targetFilename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	var tarWriter *tar.Writer
	if compress {
		gzipWriter := gzip.NewWriter(targetFile)
		defer gzipWriter.Close()

		tarWriter = tar.NewWriter(gzipWriter)
	} else {
		tarWriter = tar.NewWriter(targetFile)

	}
	defer tarWriter.Close()

	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		savePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		savePath = strings.ReplaceAll(savePath, "\\", "/")
		if !strings.HasPrefix(savePath, "/") {
			savePath = "/" + savePath
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		defer file.Close()

		header := &tar.Header{
			Name:    savePath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    0644,
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil

}

func UntarFile(sourceFilePath string, targetDir string, compressed bool) error {
	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return err
	}

	var tarReader *tar.Reader
	if compressed {
		gzipReader, err := gzip.NewReader(sourceFile)
		if err != nil {
			return err
		}
		defer gzipReader.Close()

		tarReader = tar.NewReader(gzipReader)
	} else {
		tarReader = tar.NewReader(sourceFile)
	}

	for {
		header, err := tarReader.Next()

		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		target := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {

		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
