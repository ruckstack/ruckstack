package monitor

import (
	"log"
	"time"
)

var SystemReady = false

func watchOverall() {
	for true {
		log.Println("Checking overall health..")

		if KubeClientReady &&
			ReadyNodeCount > 0 &&
			DaemonSetsReady &&
			DeploymentsReady &&
			TraefikIp != "" {

			log.Printf("Healthy")
			SystemReady = true
		} else {
			log.Printf("Unhealthy")
			SystemReady = false
		}
		log.Println("Checking overall health..DONE")

		time.Sleep(10 * time.Second)

	}
}
