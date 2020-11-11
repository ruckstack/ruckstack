package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"time"
)

func checkKubeClient(tracker *monitor.Tracker) {
	cannotConnect := "Cannot connect to local k3s"

	for true {
		serverVersion, err := kube.Client().ServerVersion()
		if err == nil {
			tracker.ResolveProblem("k3s", cannotConnect, fmt.Sprintf("Connected to local k3s version %s", serverVersion.String()))
		} else {
			tracker.FoundProblem("k3s", cannotConnect, err.Error())
		}

		time.Sleep(10 * time.Second)
	}
}
