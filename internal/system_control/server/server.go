package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/ruckstack/ruckstack/internal/system_control/k3s"
	"github.com/ruckstack/ruckstack/internal/system_control/server/monitor"
	"github.com/ruckstack/ruckstack/internal/system_control/server/webserver"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
)

func Start() error {
	//serverReady = make(chan bool, 1)
	//

	packageConfig, err := util.GetPackageConfig()
	if err != nil {
		return err
	}

	fmt.Printf("Starting %s version %s\n", packageConfig.Name, packageConfig.Version)

	if err := os.MkdirAll(filepath.Join(util.InstallDir(), "logs"), 0755); err != nil {
		return err
	}
	serverLog, err := os.OpenFile(filepath.Join(util.InstallDir(), "logs", "server.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer serverLog.Close()
	log.SetOutput(serverLog)

	fmt.Printf("    Server log: %s\n", serverLog.Name())
	fmt.Printf("    K3S log: %s\n", filepath.Join(util.InstallDir(), "logs", "k3s.log"))

	if err := ioutil.WriteFile(filepath.Join(util.InstallDir(), "data", "server.pid"), []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
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
