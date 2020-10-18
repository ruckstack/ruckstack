package status

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/internal/util"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kubeclient"
	batch "k8s.io/api/batch/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var seenJobs = map[string]bool{}

func ShowJobStatus(includeSystemJobs bool, watch bool) error {
	packageConfig := environment.PackageConfig

	fmt.Printf("Jobs in %s\n", packageConfig.Name)
	fmt.Println("----------------------------------------------------")

	kubeClient, err := kubeclient.KubeClient(environment.ServerHome)
	if err != nil {
		return err
	}

	namespaces := []string{"default"}
	if includeSystemJobs {
		namespaces = []string{"kube-system", "default"}
	}
	namespaceDetails := map[string]string{"kube-system": "System Jobs", "default": "Application Jobs"}

	for _, namespace := range namespaces {
		namespaceDesc := namespaceDetails[namespace]

		fmt.Println(namespaceDesc)
		fmt.Println("----------------------------------------------------")

		jobList, err := kubeClient.BatchV1().Jobs(namespace).List(context.Background(), meta.ListOptions{})
		if err != nil {
			return err
		}

		for _, job := range jobList.Items {
			printJobStatus(&job)

			seenJobs[util.GetAbsoluteName(job.GetObjectMeta())] = true
		}

		fmt.Println("")
	}

	if watch {
		fmt.Println("\nWatching for changes (ctrl-c to exit)...")

		watchJobs()

	}

	return nil
}

func printJobStatus(job *batch.Job) {
	fmt.Print(job.Name + ": ")

	if job.Status.Active > 0 {
		fmt.Println("RUNNING")
	} else {
		condition := job.Status.Conditions[0]
		fmt.Printf("%s at %s %s\n", condition.Type, condition.LastTransitionTime, condition.Message)
	}
}

func watchJobs() {
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Batch().V1().Jobs().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newJob := newObj.(*batch.Job)

			printJobStatus(newJob)
		},

		AddFunc: func(obj interface{}) {
			job := obj.(*batch.Job)

			if seenJobs[util.GetAbsoluteName(job.GetObjectMeta())] {
				return
			}

			printJobStatus(job)
		},
	})
	informer.Run(stopper)
}
