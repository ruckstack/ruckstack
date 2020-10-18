package containerd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/services/server"
	"github.com/containerd/containerd/services/server/config"
	"github.com/containerd/containerd/sys"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/containerd/containerd/diff/walking/plugin"
	_ "github.com/containerd/containerd/gc/scheduler"
	_ "github.com/containerd/containerd/runtime/restart/monitor"
	_ "github.com/containerd/containerd/services/containers"
	_ "github.com/containerd/containerd/services/content"
	_ "github.com/containerd/containerd/services/diff"
	_ "github.com/containerd/containerd/services/events"
	_ "github.com/containerd/containerd/services/healthcheck"
	_ "github.com/containerd/containerd/services/images"
	_ "github.com/containerd/containerd/services/introspection"
	_ "github.com/containerd/containerd/services/leases"
	_ "github.com/containerd/containerd/services/namespaces"
	_ "github.com/containerd/containerd/services/opt"
	_ "github.com/containerd/containerd/services/snapshots"
	_ "github.com/containerd/containerd/services/tasks"
	_ "github.com/containerd/containerd/services/version"

	_ "github.com/containerd/containerd/metrics/cgroups"
	_ "github.com/containerd/containerd/runtime/v1/linux"
	_ "github.com/containerd/containerd/runtime/v2"
	_ "github.com/containerd/containerd/runtime/v2/runc/options"
	_ "github.com/containerd/containerd/snapshots/native"
	_ "github.com/containerd/containerd/snapshots/overlay"
)

var containerdServer *server.Server
var client *containerd.Client
var SocketFile string

func init() {
	SocketFile = fmt.Sprintf("/run/%s/containerd.sock", environment.PackageConfig.Id)
	if !environment.IsRunningAsRoot {
		SocketFile = fmt.Sprintf("/tmp/%s/containerd.sock", environment.PackageConfig.Id)

	}
	ui.VPrintf("Using containerd socket file %s", SocketFile)
	if err := os.MkdirAll(filepath.Dir(SocketFile), 0755); err != nil {
		ui.Fatalf("cannot create containerd socket dir: %s", err)
	}
}

/**
Starts containerd and opens a client to it
*/
func Start(parent context.Context) error {
	if containerdServer != nil {
		ui.VPrintln("Containerd already started")
		return nil
	}

	ui.Println("Starting containerd...")
	defer ui.Println("Starting containerd...DONE")

	containerdLogFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "containerd.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening containerd.log: %s", err)
	}

	containerdGrpcLogFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "containerd.grpc.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening containerd.grpc.log: %s", err)
	}

	ctx := log.WithLogger(parent, logrus.NewEntry(&logrus.Logger{
		Out:          containerdLogFile,
		Formatter:    &logrus.JSONFormatter{},
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ReportCaller: true,
	}))
	containerdLog := log.GetLogger(ctx)

	grpclog.SetLoggerV2(grpclog.NewLoggerV2(containerdGrpcLogFile, containerdGrpcLogFile, containerdGrpcLogFile))

	containerdLog.Info("Starting containerd...")
	go func() {
		select {
		case <-parent.Done():
			if containerdServer != nil {
				containerdLog.Info("Stopping containerd...")
				ui.Println("Stopping containerd...")

				containerdServer.Stop()

				containerdLog.Info("Stopping containerd...DONE")
				ui.Println("Stopping containerd...DONE")
			}
		}
	}()

	containerdConfig := &config.Config{
		Root:      environment.ServerHome + "/data/containerd/root",
		State:     environment.ServerHome + "/data/containerd/state",
		PluginDir: environment.ServerHome + "/data/containerd/plugins",
		GRPC: config.GRPCConfig{
			Address: SocketFile,
			UID:     1000,
			GID:     1000,
		},
		Metrics: config.MetricsConfig{
			GRPCHistogram: false,
		},
		OOMScore: 0,
		//RequiredPlugins: []string{
		//	"leases",
		//},
	}
	containerdServer, _ = server.New(ctx, containerdConfig)

	listener, err := sys.GetLocalListener(containerdConfig.GRPC.Address, containerdConfig.GRPC.UID, containerdConfig.GRPC.GID)
	if err != nil {
		return fmt.Errorf("error creating containerd socket: %s", err)
	}

	go func() {
		defer listener.Close()

		containerdLog.Infof("Starting etcd listener at %s", listener.Addr().String())
		err := containerdServer.ServeGRPC(listener)
		if err != nil {
			containerdLog.Errorf("Error starting containerd listener: %s", err)
			ui.Fatalf("Error starting containerd listener: %s", err)
		}
	}()

	client, err = containerd.New(SocketFile)

	if err != nil {
		return err
	}

	return nil
}

func Stop() {
	containerdServer.Stop()

}

func LoadPackagedImages() error {
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
