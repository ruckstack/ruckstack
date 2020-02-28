package monitor

import (
	"fmt"
	"time"
)

func watchKubernetes() {
	CANNOT_CONNECT := "Cannot connect to local kubernetes"

	resolveProblem("Monitors not started", "Monitors are starting up")
	foundProblem(CANNOT_CONNECT, "System starting")

	for true {
		serverVersion, err := kubeClient.ServerVersion()
		if err == nil {
			resolveProblem(CANNOT_CONNECT, fmt.Sprintf("Connected to local kubernetes version %s", serverVersion.String()))

		} else {
			foundProblem(CANNOT_CONNECT, err.Error())
		}

		time.Sleep(10 * time.Second)
	}
}
