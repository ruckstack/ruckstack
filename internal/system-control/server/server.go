package server

import (
	"github.com/ruckstack/ruckstack/internal/system-control/k3s"
	"github.com/ruckstack/ruckstack/internal/system-control/server/monitor"
	"github.com/ruckstack/ruckstack/internal/system-control/server/webserver"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Start() {
	//serverReady = make(chan bool, 1)
	//
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
}
