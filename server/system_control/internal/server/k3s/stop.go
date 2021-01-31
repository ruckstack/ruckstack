package k3s

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"github.com/ruckstack/ruckstack/server/system_control/internal/util"
	"github.com/shirou/gopsutil/v3/process"
	v1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"sync"
	"syscall"
	"time"
)

func Stop(ctx context.Context) error {
	ui.Printf("Shutting down k3s...")

	k3sProcess, err := util.GetProcessFromFile(k3sPidPath)
	if err != nil {
		return err
	}

	if k3sProcess == nil {
		ui.VPrintf("k3s is already shut down")

		if err := os.Remove(k3sPidPath); err != nil {
			if !os.IsNotExist(err) {
				ui.Printf("error deleting %s: %s", k3sPidPath, err)
			}
		}

		return nil
	}

	if err := setUnschedulable(true, ctx); err != nil {
		return fmt.Errorf("cannot cordon this node: %s", err)
	}

	if err := evictPods(ctx); err != nil {
		return err
	}

	if err := killK3sProcess(k3sProcess, ctx); err != nil {
		return err
	}

	if err := os.Remove(k3sPidPath); err != nil {
		if !os.IsNotExist(err) {
			ui.Printf("error deleting %s: %s", k3sPidPath, err)
		}
	}

	ui.Printf("Shutting down k3s...DONE")
	return nil
}

func evictPods(ctx context.Context) error {
	defer ui.StartProgressf("Shutting down local containers").Stop()

	list, err := kube.Client().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var waitGroup sync.WaitGroup

	for _, pod := range list.Items {
		ui.VPrintf("Considering pod %s for eviction...", kube.FullName(pod.ObjectMeta))
		canEvict := true

		for _, ownerRef := range pod.OwnerReferences {
			if ownerRef.Kind == "DaemonSet" {
				ui.VPrintf("Cannot evict daemonset pod %s", kube.FullName(pod.ObjectMeta))
				canEvict = false
			}
		}

		if pod.Status.Phase != v1.PodRunning {
			ui.VPrintf("Cannot evict %s pod %s", pod.Status.Phase, pod.Name)
			continue
		}

		if pod.Spec.NodeName != environment.NodeName {
			ui.VPrintf("Not evicting pod %s on node %s", kube.FullName(pod.ObjectMeta), pod.Spec.NodeName)
			continue
		}

		if canEvict {
			if err := kube.Client().CoreV1().Pods(pod.ObjectMeta.Namespace).Evict(ctx, &policy.Eviction{
				ObjectMeta: pod.ObjectMeta,
			}); err != nil {
				ui.Printf("Error evicting: %s", err)
			}
			waitForPodShutdown(pod.ObjectMeta, time.Second*60, ctx, &waitGroup)
		}
	}

	waitGroup.Wait()

	return nil
}

func waitForPodShutdown(pod metav1.ObjectMeta, timeout time.Duration, ctx context.Context, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)

	go func() {
		ui.VPrintf("Waiting for pod %s to shutdown...", kube.FullName(pod))
		defer ui.VPrintf("Waiting for pod %s to shutdown...DONE", kube.FullName(pod))
		defer waitGroup.Done()

		err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
			foundPod, err := kube.Client().CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if apierrs.IsNotFound(err) {
				return true, nil // done
			}

			if err != nil {
				return true, err // stop wait with error
			}

			if foundPod.Status.Phase == v1.PodPending {
				return true, nil // done
			}

			return false, nil
		})

		if err != nil {
			ui.Printf("timeout waiting for %s to shut down", kube.FullName(pod))
		}
	}()
}

func killK3sProcess(process *process.Process, ctx context.Context) error {
	err := process.SendSignal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("error sending sigterm to k3s: %s", err)
	}

	return nil

}

func setUnschedulable(unschedulable bool, ctx context.Context) error {
	_, err := kube.Client().CoreV1().Nodes().Patch(ctx, environment.NodeName, types.JSONPatchType, []byte(fmt.Sprintf("[{\"op\": \"replace\", \"path\": \"/spec/unschedulable\", \"value\":%t}]", unschedulable)), metav1.PatchOptions{})
	return err
}
