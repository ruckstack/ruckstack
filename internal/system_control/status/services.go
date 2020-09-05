package status

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ruckstack/ruckstack/internal/system_control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type serviceInfo struct {
	name       string
	namespace  string
	kind       string
	lastStatus string
	pods       []string
}

var kubeClient *kubernetes.Clientset
var lastPodStatus = map[string]string{}
var ownerTree = map[string]*meta.OwnerReference{}
var allServices = map[string]*serviceInfo{}

func ShowServiceStatus(includeSystemService bool, watch bool) {
	fmt.Printf("Services in %s\n", util.GetPackageConfig().Name)
	fmt.Println("----------------------------------------------------")

	kubeClient = kubeclient.KubeClient()

	namespaces := []string{"default"}
	if includeSystemService {
		namespaces = []string{"kube-system", "default"}
	}
	namespaceDetails := map[string]string{"kube-system": "System Services", "default": "Application Services"}

	for _, namespace := range namespaces {
		namespaceDesc := namespaceDetails[namespace]

		fmt.Println(namespaceDesc)
		fmt.Println("----------------------------------------------------")

		replicaSetList, err := kubeClient.AppsV1().ReplicaSets(namespace).List(meta.ListOptions{})
		util.Check(err)
		for _, replicaSet := range replicaSetList.Items {
			for _, owner := range replicaSet.OwnerReferences {
				ownerTree[util.GetAbsoluteName(replicaSet.GetObjectMeta())] = &owner
			}
		}

		daemonSetList, err := kubeClient.AppsV1().DaemonSets(namespace).List(meta.ListOptions{})
		util.Check(err)
		for _, ds := range daemonSetList.Items {
			allServices[util.GetAbsoluteName(ds.GetObjectMeta())] = &serviceInfo{
				name:       ds.Name,
				namespace:  ds.Namespace,
				kind:       "DaemonSet",
				lastStatus: getDaemonSetStatus(&ds),
			}
		}

		deploymentList, err := kubeClient.AppsV1().Deployments(namespace).List(meta.ListOptions{})
		util.Check(err)
		for _, deployment := range deploymentList.Items {
			allServices[util.GetAbsoluteName(deployment.GetObjectMeta())] = &serviceInfo{
				name:       deployment.Name,
				namespace:  deployment.Namespace,
				kind:       "Deployment",
				lastStatus: getDeploymentStatus(&deployment),
			}
		}

		statefulSetList, err := kubeClient.AppsV1().StatefulSets(namespace).List(meta.ListOptions{})
		util.Check(err)
		for _, statefulSet := range statefulSetList.Items {
			allServices[util.GetAbsoluteName(statefulSet.GetObjectMeta())] = &serviceInfo{
				name:       statefulSet.Name,
				namespace:  statefulSet.Namespace,
				kind:       "StatefulSet",
				lastStatus: getStatefulSetStatus(&statefulSet),
			}
		}

		podList, err := kubeClient.CoreV1().Pods(namespace).List(meta.ListOptions{})
		util.Check(err)
		for _, pod := range podList.Items {
			lastPodStatus[util.GetAbsoluteName(pod.GetObjectMeta())] = getPodStatusDescription(&pod)

			owner := getOwnerService(&pod)
			if owner != nil {
				owner.pods = append(owner.pods, getPodStatus(&pod))
			}
		}

		var seenServiceNames []string
		for k := range allServices {
			seenServiceNames = append(seenServiceNames, k)
		}
		sort.Strings(seenServiceNames)

		for _, name := range seenServiceNames {
			status := allServices[name]
			fmt.Println(status.lastStatus)

			if len(status.pods) == 0 {
				fmt.Println("- No containers")
			}

			sort.Strings(status.pods)
			for _, podMessage := range status.pods {
				fmt.Println(" - " + podMessage)
			}
		}

		fmt.Println("")
	}

	if watch {
		fmt.Println("\nWatching for changes (ctrl-c to exit)...")

		stopper := make(chan struct{})
		defer close(stopper)

		var wg sync.WaitGroup
		wg.Add(4)

		go watchPods(&wg)
		go watchDaemonSets(&wg)
		go watchDeployments(&wg)
		go watchStatefulSets(&wg)

		wg.Wait()

		//TODO: flag for showing system services

	}
}

func watchPods(wg *sync.WaitGroup) {
	defer wg.Done()
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Pods().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newPod := newObj.(*core.Pod)

			owner := getOwnerService(newPod)
			if owner == nil {
				//no service for pod
				return
			}

			lastStatus := lastPodStatus[util.GetAbsoluteName(newPod.GetObjectMeta())]
			if lastStatus == "" || lastStatus != getPodStatusDescription(newPod) {
				displayPodChanged(newPod)
			} else {
				//unchanged status
			}
		},

		AddFunc: func(obj interface{}) {
			pod := obj.(*core.Pod)

			owner := getOwnerService(pod)
			if owner == nil {
				//not a service pod
				return
			}

			lastStatus := lastPodStatus[util.GetAbsoluteName(pod.GetObjectMeta())]
			if lastStatus != "" {
				//already seen before
				return
			}

			displayPodChanged(pod)
		},

		DeleteFunc: func(obj interface{}) {
			pod := obj.(*core.Pod)

			owner := getOwnerService(pod)
			if owner == nil {
				// No service for pod
				return
			}

			fmt.Printf("Container %s on %s has been shut down\n", pod.Name, pod.Spec.NodeName)
		},
	})
	informer.Run(stopper)
}

func getPodStatusDescription(newPod *core.Pod) string {
	if newPod.Status.Phase != core.PodRunning {
		return string(newPod.Status.Phase)
	}

	for _, containerStatus := range newPod.Status.ContainerStatuses {

		if !containerStatus.Ready {
			if containerStatus.State.Terminated != nil {
				return "Stopping"
			} else {
				return "Starting"
			}

		}
	}

	return string(newPod.Status.Phase)
}

func watchDaemonSets(wg *sync.WaitGroup) {
	defer wg.Done()
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Apps().V1().DaemonSets().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newDaemonSet := newObj.(*apps.DaemonSet)

			serviceInfo := allServices[util.GetAbsoluteName(newDaemonSet.GetObjectMeta())]
			if serviceInfo == nil {
				//not a tracked service
				return
			}
			lastStatus := serviceInfo.lastStatus
			newStatus := getDaemonSetStatus(newDaemonSet)
			if lastStatus == "" || lastStatus != newStatus {
				if newDaemonSet.Namespace == "kube-system" {
					fmt.Print("System service ")
				} else {
					fmt.Print("Application service ")
				}
				fmt.Println(newStatus)
				serviceInfo.lastStatus = newStatus
			} else {
				//unchanged status
			}
		},

		AddFunc: func(obj interface{}) {
			//not logging added services
		},

		DeleteFunc: func(obj interface{}) {
			//not logging deleted services
		},
	})
	informer.Run(stopper)
}

func watchDeployments(wg *sync.WaitGroup) {
	defer wg.Done()
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Apps().V1().Deployments().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newDeployment := newObj.(*apps.Deployment)

			serviceInfo := allServices[util.GetAbsoluteName(newDeployment.GetObjectMeta())]
			if serviceInfo == nil {
				//not a tracked service
				return
			}

			lastStatus := serviceInfo.lastStatus
			newStatus := getDeploymentStatus(newDeployment)
			if lastStatus == "" || lastStatus != newStatus {
				if newDeployment.Namespace == "kube-system" {
					fmt.Print("System service ")
				} else {
					fmt.Print("Application service ")
				}
				fmt.Println(newStatus)
				serviceInfo.lastStatus = newStatus
			} else {
				//unchanged status
			}
		},

		AddFunc: func(obj interface{}) {
			//not logging added services
		},

		DeleteFunc: func(obj interface{}) {
			//not logging deleted services
		},
	})
	informer.Run(stopper)
}

func watchStatefulSets(wg *sync.WaitGroup) {
	defer wg.Done()
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Apps().V1().StatefulSets().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newStatefulSet := newObj.(*apps.StatefulSet)

			serviceInfo := allServices[util.GetAbsoluteName(newStatefulSet.GetObjectMeta())]
			if serviceInfo == nil {
				//not a tracked service
				return
			}

			lastStatus := serviceInfo.lastStatus
			newStatus := getStatefulSetStatus(newStatefulSet)
			if lastStatus == "" || lastStatus != newStatus {
				if newStatefulSet.Namespace == "kube-system" {
					fmt.Print("System service ")
				} else {
					fmt.Print("Application service ")
				}
				fmt.Println(newStatus)
				serviceInfo.lastStatus = newStatus
			} else {
				//unchanged status
			}
		},

		AddFunc: func(obj interface{}) {
			//not logging added services
		},

		DeleteFunc: func(obj interface{}) {
			//not logging deleted services
		},
	})
	informer.Run(stopper)
}
func getOwnerService(pod *core.Pod) *serviceInfo {
	owners := pod.OwnerReferences
	if owners != nil {
		for _, owner := range owners {
			altOwner := ownerTree[pod.Namespace+"/"+owner.Name]
			if altOwner != nil {
				owner = *altOwner
			}

			serviceInfo := allServices[pod.Namespace+"/"+owner.Name]
			if serviceInfo != nil {
				return serviceInfo
			}
		}
	}
	return nil
}

func getDaemonSetStatus(daemonSet *apps.DaemonSet) string {
	returnMessage := fmt.Sprintf("%s: ", daemonSet.Name)

	if daemonSet.Status.NumberReady == 0 {
		returnMessage += "UNAVAILABLE. No containers are ready"
	} else if daemonSet.Status.NumberUnavailable > 0 {
		returnMessage += fmt.Sprintf("DEGRADED. Not running on %d nodes", daemonSet.Status.NumberUnavailable)
	} else {
		returnMessage += "HEALTHY"
	}

	return returnMessage
}

func getDeploymentStatus(deployment *apps.Deployment) string {
	returnMessage := fmt.Sprintf("%s: ", deployment.Name)

	if deployment.Status.AvailableReplicas == 0 {
		returnMessage += "UNAVAILABLE. No containers are ready"
	} else if *deployment.Spec.Replicas < deployment.Status.AvailableReplicas {
		returnMessage += fmt.Sprintf("DEGRADED. Only %d of %d expected containers are ready", deployment.Spec.Replicas, deployment.Status.AvailableReplicas)
	} else {
		returnMessage += "HEALTHY"
	}

	return returnMessage
}

func getStatefulSetStatus(statefulSet *apps.StatefulSet) string {
	returnMessage := fmt.Sprintf("%s: ", statefulSet.Name)

	if statefulSet.Status.ReadyReplicas == 0 {
		returnMessage += "UNAVAILABLE. No containers are ready"
	} else if *statefulSet.Spec.Replicas < statefulSet.Status.ReadyReplicas {
		returnMessage += fmt.Sprintf("DEGRADED. Only %d of %d expected containers are ready", statefulSet.Spec.Replicas, statefulSet.Status.ReadyReplicas)
	} else {
		returnMessage += "HEALTHY"
	}

	return returnMessage
}

func getPodStatus(pod *core.Pod) string {
	returnMessage := fmt.Sprintf("Container %s on %s: %s", pod.Name, pod.Spec.NodeName, getPodStatusDescription(pod))

	//if statefulSet.Status.ReadyReplicas == 0 {
	//	returnMessage += "UNAVAILABLE. No instances ready"
	//} else if *statefulSet.Spec.Replicas < statefulSet.Status.ReadyReplicas {
	//	returnMessage += fmt.Sprintf("DEGRADED. Only %d of configured %d are ready", statefulSet.Spec.Replicas, statefulSet.Status.ReadyReplicas)
	//} else {
	//	returnMessage += "RUNNING"
	//}

	return returnMessage
}

func displayPodChanged(pod *core.Pod) {
	if pod.Spec.NodeName == "" {
		fmt.Printf("Container %s is waiting for a node assignment\n", pod.Name)
	} else {
		podStatus := getPodStatusDescription(pod)
		lastPodStatus[util.GetAbsoluteName(pod.GetObjectMeta())] = podStatus

		fmt.Printf("Container %s on %s is now %s\n", pod.Name, pod.Spec.NodeName, podStatus)
	}
}
