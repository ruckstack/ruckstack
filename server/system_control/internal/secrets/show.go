package secrets

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"unicode/utf8"
)

func Show(secretName string, secretKey string, systemSecrets bool) error {

	ctx := context.Background()

	client := kube.Client()

	namespace := "default"
	if systemSecrets {
		namespace = "kube-system"
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error accessing secure data: %s", err)
	}

	data, found := secret.Data[secretKey]
	if !found {
		return fmt.Errorf("no key %s in %s", secretKey, secretName)
	}
	if !utf8.Valid(data) {
		return fmt.Errorf("%s.%s contains binary data. Cannot display", secretName, secretKey)
	}

	ui.Println(string(data))

	return nil
}
