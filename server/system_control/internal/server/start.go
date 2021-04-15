package server

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/containerd"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/k3s"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/proxy"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/webserver"
	"github.com/ruckstack/ruckstack/server/system_control/internal/util"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

var serverPidPath string

func init() {
	serverPidPath = filepath.Join(environment.ServerHome, "data", "server.pid")
}

func Start() error {

	serverProcess, err := util.GetProcessFromFile(serverPidPath)
	if err != nil {
		return fmt.Errorf("cannot check %s lock file: %s", serverPidPath, err)
	}
	if serverProcess != nil {
		err = serverProcess.SendSignal(syscall.Signal(0))
		if err == nil {
			ui.Fatalf("Server already running under process id %d. Cannot run a second instance", serverProcess.Pid)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ui.Printf("Received shutdown command")

		cancel()

		stopErr := Stop(false)
		if stopErr != nil {
			ui.Printf("error stopping server: %s", stopErr)
		}

		os.Exit(0)
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

	if err := ioutil.WriteFile(serverPidPath, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return err
	}

	log.Println("Starting server components...")

	if err := monitor.Start(ctx); err != nil {
		return fmt.Errorf("error starting monitor: %s", err)
	}

	if err := webserver.Start(ctx); err != nil {
		return fmt.Errorf("error starting webserver: %s", err)
	}

	if err := containerd.StartManager(ctx); err != nil {
		return fmt.Errorf("error starting containerd: %s", err)
	}

	if err := k3s.Start(ctx); err != nil {
		return fmt.Errorf("error starting k3s server: %s", err)
	}

	if err := proxy.Start(ctx); err != nil {
		return fmt.Errorf("error starting proxy server: %s", err)
	}

	ui.Println("Server started")
	ui.Printf("Additional logs are available through `%s logs` or in %s/logs", environment.SystemConfig.ManagerFilename, environment.ServerHome)
	ui.Printf("System can be watched with `%s status`", environment.SystemConfig.ManagerFilename)

	for true {
		time.Sleep(100000 * time.Hour)
	}

	return nil
}
