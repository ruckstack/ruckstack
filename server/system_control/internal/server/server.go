package server

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/containerd"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/k3s"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/webserver"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

func Start() error {

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	packageConfig := environment.PackageConfig

	fmt.Printf("Starting %s version %s\n", packageConfig.Name, packageConfig.Version)

	serverLog, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "server.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer serverLog.Close()
	log.SetOutput(serverLog)

	fmt.Printf("    Server log: %s\n", serverLog.Name())
	fmt.Printf("    K3S log: %s\n", filepath.Join(environment.ServerHome, "logs", "k3s.log"))

	if err := ioutil.WriteFile(filepath.Join(environment.ServerHome, "data", "server.pid"), []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return err
	}

	log.Println("Starting server components...")

	if err := monitor.Start(ctx); err != nil {
		return fmt.Errorf("error starting monitor: %s", err)
	}

	if err := webserver.Start(ctx); err != nil {
		return fmt.Errorf("error starting webserver: %s", err)
	}

	if err := containerd.Start(ctx); err != nil {
		return fmt.Errorf("error starting containerd: %s", err)
	}

	//if err := etcd.Start(ctx); err != nil {
	//	return fmt.Errorf("error starting etcd: %s", err)
	//}

	if err := k3s.Start(ctx); err != nil {
		return fmt.Errorf("error starting k3s server: %s", err)
	}

	//if err := containerd.LoadPackagedImages(); err != nil {
	//	return fmt.Errorf("error loading images: %s", err)
	//}

	//if err := agent.Start(ctx); err != nil {
	//	ui.Fatalf("error starting k3s agent: %s", err)
	//}

	//go monitor.StartMonitor()

	ui.Println("Server started")
	ui.Printf("Additional logs are available through `%s logs` or in %s/logs", environment.PackageConfig.ManagerFilename, environment.ServerHome)
	ui.Printf("System can be watched with `%s status`", environment.PackageConfig.ManagerFilename)

	select {
	case <-ctx.Done():
		ui.Println("Server shutting down...")
		ui.VPrintf("Shutdown reason: %s", ctx.Err())
		return nil
	}
}
