package secrets

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func List(systemSecrets bool) error {

	ctx := context.Background()

	client := kube.Client()

	namespace := "default"
	if systemSecrets {
		namespace = "kube-system"
	}

	secrets, err := client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error accessing secure data: %s", err)
	}

	ui.Println("Available Secure Configuration Data:")
	ui.Println("-----------------------")
	for _, foundSecret := range secrets.Items {
		ui.Printf("%s", foundSecret.Name)
		for key, _ := range foundSecret.Data {
			ui.Printf(" - %s", key)
		}
		ui.Println()
	}

	return nil

}
