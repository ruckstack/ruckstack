package configmap

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Show(configName string, configKey string, systemConfig bool) error {

	ctx := context.Background()

	client := kube.Client()

	namespace := "default"
	if systemConfig {
		namespace = "kube-system"
	}

	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, configName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error accessing configuration: %s", err)
	}

	data, found := configMap.Data[configKey]
	if !found {
		return fmt.Errorf("no key %s in %s", configKey, configName)
	}

	ui.Println(data)

	return nil
}
