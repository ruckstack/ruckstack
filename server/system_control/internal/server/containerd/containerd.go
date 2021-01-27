package containerd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"github.com/shirou/gopsutil/v3/process"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var containerdAddress = "/run/k3s/containerd/containerd.sock"
var containerdClient *containerd.Client
var containerdClientConnectErr error
var logger *log.Logger

/**
Starts the containerd manager and monitor. Actual containerd is started in k3s.
*/
func StartManager(ctx context.Context) error {
	ui.Println("Starting containerd manager...")

	logFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "containerd.manager.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening webserver.log: %s", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)
	logger.Println("Starting containerd manager...")

	go func() {
		select {
		case <-ctx.Done():
			logger.Println("Containerd shutting down...")
		}
	}()

	go func() {
		var err error
		for containerdClient == nil {
			logger.Printf("Connecting to containerd at %s...", containerdAddress)
			containerdClient, containerdClientConnectErr = containerd.New(containerdAddress)
			if containerdClientConnectErr == nil {
				logger.Print("Containerd client connected")
			} else {
				logger.Printf("Waiting for containerd...%s", containerdClientConnectErr)
				time.Sleep(2 * time.Second)
			}
		}

		err = LoadPackagedImages()
		if err != nil {
			logger.Printf("cannot load packaged images: %s", err)
		}
	}()

	monitor.Add(&monitor.Tracker{
		Name:  "Containerd Client",
		Check: checkContainerdHealth,
	})

	return nil
}

func LoadPackagedImages() error {
	logger.Printf("Loading packaged images...")
	defer logger.Printf("Loading packaged images...DONE")

	imagesDir := filepath.Join(environment.ServerHome, "data/agent/images")

	untarDirs, err := filepath.Glob(imagesDir + "/*.untar")
	if err != nil {
		return err
	}

	ctxContainerD := namespaces.WithNamespace(context.Background(), "k8s.io")

	for _, untarDir := range untarDirs {
		importedInfoPath := untarDir + "/imported.info"
		importDirectory := false
		currentHash := "UNKNOWN"

		manifestFile, err := os.Open(untarDir + "/manifest.json")
		if err == nil {
			hash := sha1.New()
			_, err = io.Copy(hash, manifestFile)
			if err != nil {
				return fmt.Errorf("cannot compute hash for %s: %s", untarDir+"/manifest.json", err)
			}

			currentHash = hex.EncodeToString(hash.Sum(nil)[:20])

			importedHash, err := ioutil.ReadFile(importedInfoPath)
			if err == nil {
				if string(importedHash) == currentHash {
					logger.Printf("Directory %s has already been imported", untarDir)
					importDirectory = false
				} else {
					logger.Printf("Directory %s has changed since the last time it was imported", untarDir)
					importDirectory = true
				}
			} else {
				if err == os.ErrNotExist {
					logger.Printf("Directory %s has not been imported before", untarDir)
					importDirectory = true
				} else {
					logger.Printf("Cannot check imported.info file: %s", err)
					importDirectory = true
				}
			}
		} else {
			if err == os.ErrNotExist {
				logger.Printf("Invalid untar dir %s. No manifest.json file", untarDir)
				continue
			} else {
				logger.Printf("Cannot check manifest.json file: %s", err)
				continue
			}
		}

		if !importDirectory {
			continue
		}

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
				logger.Fatal(err)
			}

			tarWriter.Close()
			pipeWriter.Close()

		}()

		logger.Printf("Importing images in %s...", untarDir)
		images, err := containerdClient.Import(ctxContainerD, pipeReader, containerd.WithAllPlatforms(true), containerd.WithImportCompression())
		if err != nil {
			return err
		}

		for _, image := range images {
			logger.Printf("Imported %s", image.Name)
		}

		err = ioutil.WriteFile(importedInfoPath, []byte(currentHash), 0644)
		if err != nil {
			logger.Printf("Cannot write %s: %s", importedInfoPath, err)
		}

		logger.Printf("Importing images in %s...DONE", untarDir)
	}

	return nil
}

func KillProcesses(ctx context.Context) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	var waitGroup sync.WaitGroup

	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			ui.Printf("Cannot get proc name for %d", proc.Pid)
			continue
		}
		if strings.HasPrefix(name, "containerd-shim") {
			cmdLine, err := proc.CmdlineSlice()
			if err != nil {
				ui.Printf("Cannot get command line for %d", proc.Pid)
				continue
			}
			for _, arg := range cmdLine {
				if arg == containerdAddress {
					global_util.ShutdownProcess(proc, 30*time.Second, &waitGroup, ctx)
				}
			}
		}
	}

	waitGroup.Wait()

	return nil
}
