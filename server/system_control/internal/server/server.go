package server

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/agent"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/containerd"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/ruckstack/ruckstack/server/system_control/internal/server/k3s"
)

func Start() error {
	//serverReady = make(chan bool, 1)
	//

	packageConfig := environment.PackageConfig

	fmt.Printf("Starting %s version %s\n", packageConfig.Name, packageConfig.Version)

	if err := os.MkdirAll(filepath.Join(environment.ServerHome, "logs"), 0755); err != nil {
		return err
	}
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

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("Recieved sig %s", sig.String())
		Stop()
		done <- true
	}()

	log.Println("Starting Server...")

	if err := containerd.Start(); err != nil {
		ui.Fatalf("error starting containerd: %s", err)
	}
	ui.Println("Containerd started successfully")

	//go monitor.StartMonitor()

	//go webserver.StartWebserver()

	if err := k3s.Start(); err != nil {
		ui.Fatalf("error starting k8s server: %s", err)
	}
	ui.Println("K3s server started successfully")

	if err := agent.Start(); err != nil {
		ui.Fatalf("error starting k3s agent: %s", err)
	}
	ui.Println("K3s agent started successfully")

	<-done
	log.Println("Exiting")

	return nil
}

func Stop() {
	k3s.Stop()
}
