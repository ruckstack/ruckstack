package monitor

import (
	"log"
	"time"
)

var KubeClientReady = false

func watchKubernetes() {
	for true {
		serverVersion, err := kubeClient.ServerVersion()
		if err == nil {
			log.Printf("Connected to Kubernetes version %s", serverVersion)
			KubeClientReady = true

		} else {
			log.Printf("Error connecting to kubernetes: %s", err)
			KubeClientReady = false
		}

		time.Sleep(10 * time.Second)
	}
}
