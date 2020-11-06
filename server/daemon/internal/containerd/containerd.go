package containerd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/prometheus/common/log"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/daemon/internal/k3s"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func LoadPackagedImages() error {
	log.Info("Loading packaged images...")
	defer log.Info("Loading packaged images...DONE")

	client := k3s.ContainerdClient

	imagesDir := filepath.Join(environment.ServerHome, "data/agent/images")

	untarDirs, err := filepath.Glob(imagesDir + "/*.untar")
	if err != nil {
		return err
	}

	ctxContainerD := namespaces.WithNamespace(context.Background(), "k8s.io")

	ui.VPrintf("Found %d untar directories", len(untarDirs))
	for _, untarDir := range untarDirs {

		pipeReader, pipeWriter := io.Pipe()

		tarWriter := tar.NewWriter(pipeWriter)

		go func() {
			if err := filepath.Walk(untarDir, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}

				relativePath, err := filepath.Rel(untarDir, path)
				if err != nil {
					return err
				}

				file, err := os.Open(path)
				if err != nil {
					return err
				}

				defer file.Close()

				if strings.HasSuffix(relativePath, "layer.tar.gz") {
					_, _ = file.Seek(-4, 2)

					sizeBuffer := new(bytes.Buffer)
					written, err := io.Copy(sizeBuffer, file)
					if err != nil {
						return err
					}
					if written != 4 {
						return fmt.Errorf("Didn't read size correctly")
					}
					uncompressedLength := binary.LittleEndian.Uint32(sizeBuffer.Bytes())

					_, _ = file.Seek(0, 0)
					gzipReader, err := gzip.NewReader(file)
					if err != nil {
						return err
					}

					header := &tar.Header{
						Name:    strings.Replace(relativePath, ".tar.gz", ".tar", 1),
						ModTime: info.ModTime(),
						Size:    int64(uncompressedLength),
						Mode:    0644,
					}

					err = tarWriter.WriteHeader(header)
					if err != nil {
						return err
					}

					_, err = io.Copy(tarWriter, gzipReader)
					if err != nil {
						return err
					}
					gzipReader.Close()
				} else {
					header := &tar.Header{
						Name:    relativePath,
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
				}

				return nil
			}); err != nil {
				ui.Fatal(err)
			}

			tarWriter.Close()
			pipeWriter.Close()

		}()

		images, err := client.Import(ctxContainerD, pipeReader, containerd.WithAllPlatforms(true))
		if err != nil {
			return err
		}
		fmt.Println(images)
	}

	return nil
}
