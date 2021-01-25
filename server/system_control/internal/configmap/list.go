package configmap

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func List(systemConfigs bool) error {

	ctx := context.Background()

	client := kube.Client()

	namespace := "default"
	if systemConfigs {
		namespace = "kube-system"
	}

	configs, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error accessing configuration data: %s", err)
	}

	ui.Println("Available Configuration Data:")
	ui.Println("-----------------------")
	for _, foundConfig := range configs.Items {
		ui.Printf("%s", foundConfig.Name)
		for key, _ := range foundConfig.Data {
			ui.Printf(" - %s", key)
		}
		ui.Println()
	}

	return nil

}
