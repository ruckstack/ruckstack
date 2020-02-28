package monitor

import (
	"github.com/ruckstack/ruckstack/internal/system-control/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"time"
)

var kubeClient *kubernetes.Clientset

var (
	knownProblems = map[string]string{}
	seenProblems  = map[string]bool{}
)

var ServerStatus = struct {
	TraefikIp   string
	SystemReady bool
}{}

func StartMonitor() {
	log.Println("Starting monitor...")

	for !kubeclient.ConfigExists() {
		log.Printf("Monitor waiting for %s", kubeclient.KubeconfigFile())

		time.Sleep(10 * time.Second)
	}

	kubeClient = kubeclient.KubeClient()

	foundProblem("Monitors not started", "System starting")

	go watchKubernetes()
	go watchNodes()
	go watchDaemonSets()
	go watchDeployments()
	go watchServices()
	go watchOverall()

	log.Println("Starting monitor...Complete")

}

func fullName(obj metav1.ObjectMeta) string {
	if obj.Namespace == "" {
		return obj.Name
	}
	return obj.Namespace + "." + obj.Name
}

func foundProblem(problemKey string, description string) {
	seenProblems[problemKey] = true

	existingDesc, problemExists := knownProblems[problemKey]
	if !problemExists || existingDesc != description {
		message := problemKey
		if description != "" {
			message += " -- " + description
		}
		log.Println("PROBLEM: " + message)
	}

	knownProblems[problemKey] = description
}

func resolveProblem(problemKey string, resolvedMessage string) {
	_, problemExists := knownProblems[problemKey]
	if problemExists {
		delete(knownProblems, problemKey)
		log.Println("RESOLVED: " + resolvedMessage)
	} else {
		if !seenProblems[problemKey] {
			log.Println("RESOLVED: " + resolvedMessage)
			seenProblems[problemKey] = true
		}
	}
}
