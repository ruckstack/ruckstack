package server

import (
	"fmt"
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/ruckstack/ruckstack/server/system_control/internal/k3s"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/webserver"
)

func Start() error {
	//serverReady = make(chan bool, 1)
	//

	packageConfig, err := common2.GetPackageConfig()
	if err != nil {
		return err
	}

	fmt.Printf("Starting %s version %s\n", packageConfig.Name, packageConfig.Version)

	if err := os.MkdirAll(filepath.Join(common2.InstallDir(), "logs"), 0755); err != nil {
		return err
	}
	serverLog, err := os.OpenFile(filepath.Join(common2.InstallDir(), "logs", "server.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer serverLog.Close()
	log.SetOutput(serverLog)

	fmt.Printf("    Server log: %s\n", serverLog.Name())
	fmt.Printf("    K3S log: %s\n", filepath.Join(common2.InstallDir(), "logs", "k3s.log"))

	if err := ioutil.WriteFile(filepath.Join(common2.InstallDir(), "data", "server.pid"), []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("Recieved sig %s", sig.String())
		k3s.Stop()
		done <- true
	}()

	log.Println("Starting Server...")

	go monitor.StartMonitor()

	go webserver.StartWebserver()
	go k3s.Start()

	<-done
	log.Println("Exiting")

	return nil
}
