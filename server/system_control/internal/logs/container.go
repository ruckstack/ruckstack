package logs

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"math"
	"regexp"
	"strings"
	"time"
)

func ShowContainerLogs(systemContainer bool, containerId string, watch bool, previous bool, since string) error {
	if previous {
		//cannot watch previous container logs
		watch = false
	}

	namespace := "default"
	if systemContainer {
		namespace = "kube-system"
	}

	client := kube.Client()

	logOptions := &core.PodLogOptions{
		Follow: watch,
	}

	fmt.Print("Logs for")

	if previous {
		fmt.Print(" PREVIOUS")
		logOptions.Previous = previous
	}

	fmt.Printf(" container %s", containerId)
	if strings.ToLower(since) != "all" {
		duration, err := time.ParseDuration(since)
		if err != nil {
			return err
		}

		sinceSeconds := int64(math.Abs(duration.Seconds()))
		logOptions.SinceSeconds = &sinceSeconds

		fmt.Printf(" since %s", time.Now().Add(time.Duration(-1*sinceSeconds)*time.Second).Format(time.RFC822))
	}

	if watch {
		fmt.Println(" (ctrl-c to exit)...")
	} else {
		fmt.Println("")
	}
	fmt.Println("-----------------------------------------")

	if err := outputLogs(namespace, containerId, false, logOptions, client); err != nil {
		return err
	}

	return nil
}

func outputLogs(namespace string, containerId string, includeContainerId bool, logOptions *core.PodLogOptions, client *kubernetes.Clientset) error {
	logs := client.CoreV1().Pods(namespace).GetLogs(containerId, logOptions)
	logStream, err := logs.Stream(context.Background())
	if err != nil {
		return err
	}
	defer logStream.Close()

	colorRemover := regexp.MustCompile("\\x1b\\[[0-9;]*m")

	scanner := bufio.NewScanner(logStream)
	for scanner.Scan() {
		line := scanner.Text()
		if includeContainerId {
			line = "[" + containerId + "] " + line
		}
		fmt.Println(colorRemover.ReplaceAllString(line, ""))
	}

	return nil
}
