package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var traefikNotListening = "Traefik is not listening on an IP address"
var traefikNotStarted = "Traefik is not running"
var traefikServiceName = "kube-system.traefik"

func checkTraefik(tracker *monitor.Tracker) {
	factory := informers.NewSharedInformerFactory(kube.Client(), 0)
	informer := factory.Core().V1().Services().Informer()

	tracker.FoundProblem(traefikServiceName, traefikNotStarted, "Service not started")
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newService := newObj.(*core.Service)
			tracker.Logf("Detected updated service %s", kube.FullName(newService.ObjectMeta))

			checkService(newService, tracker)
		},

		AddFunc: func(obj interface{}) {
			service := obj.(*core.Service)
			tracker.Logf("Detected added service %s", kube.FullName(service.ObjectMeta))

			checkService(service, tracker)
		},

		DeleteFunc: func(obj interface{}) {
			service := obj.(*core.Service)
			tracker.Logf("Detected deleted service %s", kube.FullName(service.ObjectMeta))

			if IsTraefik(service) {
				tracker.FoundProblem(traefikServiceName, traefikNotStarted, "Traefik service deleted")
			}

			checkService(service, tracker)
		},
	})
	informer.Run(tracker.Context.Done())

}

func checkService(service *core.Service, tracker *monitor.Tracker) {
	if IsTraefik(service) {
		tracker.ResolveProblem(traefikServiceName, traefikNotStarted, "")

		if service.Spec.ClusterIP == "" {
			tracker.FoundProblem(traefikServiceName, traefikNotListening, "")
		} else {
			tracker.ResolveProblem(traefikServiceName, traefikNotListening, fmt.Sprintf("Traefik is listening on %s", service.Spec.ClusterIP))
		}
	}
}

func IsTraefik(service *core.Service) bool {
	return kube.FullName(service.ObjectMeta) == traefikServiceName
}
