package containerd

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

func checkContainerdHealth(tracker *monitor.Tracker) {
	problemKey := "Containerd not healthy"
	componentKey := "containerd"

	for containerdClient == nil {
		tracker.FoundProblem(componentKey, problemKey, fmt.Sprintf("Client not connected: %s", containerdClientConnectErr))

		time.Sleep(5 * time.Second)
	}

	var watcher grpc_health_v1.Health_WatchClient
	var err error
	for watcher == nil {
		watcher, err = containerdClient.HealthService().Watch(context.Background(), &grpc_health_v1.HealthCheckRequest{})

		if err != nil {
			tracker.FoundProblem(componentKey, problemKey, fmt.Sprintf("Cannot create watcher: %s", err))
			time.Sleep(10 * time.Second)
		}
	}

	for true {
		healthCheckResponse, err := watcher.Recv()
		if err != nil {
			tracker.FoundProblem(componentKey, problemKey, fmt.Sprintf("Error checking health: %s", err))
			continue
		}

		if healthCheckResponse.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			tracker.ResolveProblem(componentKey, problemKey, healthCheckResponse.Status.String())
		} else {
			tracker.FoundProblem(componentKey, problemKey, healthCheckResponse.Status.String())
		}
	}

}
