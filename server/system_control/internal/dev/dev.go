package dev

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func Enable() error {
	ui.Println("WARNING: 'Development Mode' is designed to help with the development and testing of this system.")
	ui.Println("WARNING: It is NOT intended for production systems and may impact performance and/or security.")
	ui.Println("")

	if !ui.PromptForBoolean("Enable Development Mode", &ui.FalseBoolean) {
		return nil
	}

	ctx := context.Background()

	client := kube.Client()
	config, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "dev-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot read current dev config: %s", err)
	}

	if config.Data == nil {
		config.Data = map[string]string{}
	}

	enabled := config.Data["enabled"]
	if enabled == "true" {
		ui.Println("Development mode already enabled")
		if err := SaveDevMode(true); err != nil {
			return err
		}

		return nil
	}

	config.Data["enabled"] = "true"

	_, err = client.CoreV1().ConfigMaps("kube-system").Update(ctx, config, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cannot enable development mode: %s", err)
	}
	if err := SaveDevMode(true); err != nil {
		return err
	}

	ui.Println("Development mode enabled")

	return nil
}

func SaveDevMode(newMode bool) error {
	environment.ClusterConfig.DevModeEnabled = newMode
	if err := environment.ClusterConfig.Save(environment.ServerHome, environment.PackageConfig, environment.LocalConfig); err != nil {
		return fmt.Errorf("error saving configuration file: %s", err)
	}
	return nil
}

func Disable() error {
	if !ui.PromptForBoolean("Disable Development Mode", &ui.FalseBoolean) {
		return nil
	}

	ctx := context.Background()

	client := kube.Client()
	config, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "dev-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot read current dev config: %s", err)
	}
	enabled := config.Data["enabled"]
	if enabled == "false" {
		ui.Println("Development mode already disabled")
		if err := SaveDevMode(false); err != nil {
			return err
		}

		return nil
	}

	config.Data["enabled"] = "false"

	_, err = client.CoreV1().ConfigMaps("kube-system").Update(ctx, config, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cannot disable development mode: %s", err)
	}
	if err := SaveDevMode(false); err != nil {
		return err
	}

	ui.Println("Development mode disabled")

	return nil
}

func Reroute(serviceName string, targetHost string, targetPort int) error {
	ctx := context.Background()

	client := kube.Client()
	config, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "dev-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot read current dev config: %s", err)
	}

	proxyConfigs := map[string]ProxyConfig{}

	proxyString := config.Data["proxy"]
	if proxyString != "" {
		decoder := yaml.NewDecoder(strings.NewReader(proxyString))
		if err := decoder.Decode(proxyConfigs); err != nil {
			return fmt.Errorf("cannt parse proxy config: %s", err)
		}
	}

	proxyConfigs[serviceName] = ProxyConfig{
		TargetHost: targetHost,
		TargetPort: targetPort,
	}

	output := new(bytes.Buffer)
	encoder := yaml.NewEncoder(output)
	if err := encoder.Encode(proxyConfigs); err != nil {
		return fmt.Errorf("cannot write proxy config: %s", err)
	}
	config.Data["proxy"] = output.String()

	_, err = client.CoreV1().ConfigMaps("kube-system").Update(ctx, config, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cannot save proxy config: %s", err)
	}

	ui.Printf("Requests for service '%s' will now be proxied to http://%s:%d", serviceName, targetHost, targetPort)

	return nil
}

func RemoveRoute(serviceName string) error {
	ctx := context.Background()

	client := kube.Client()
	config, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "dev-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot read current dev config: %s", err)
	}

	proxyConfigs := map[string]ProxyConfig{}

	proxyString := config.Data["proxy"]
	if proxyString != "" {
		decoder := yaml.NewDecoder(strings.NewReader(proxyString))
		if err := decoder.Decode(proxyConfigs); err != nil {
			return fmt.Errorf("cannt parse proxy config: %s", err)
		}
	}

	_, ok := proxyConfigs[serviceName]
	if !ok {
		return fmt.Errorf("ERROR: service '%s' is not rerouted", serviceName)
	}

	delete(proxyConfigs, serviceName)

	output := new(bytes.Buffer)
	encoder := yaml.NewEncoder(output)
	if err := encoder.Encode(proxyConfigs); err != nil {
		return fmt.Errorf("cannot write proxy config: %s", err)
	}
	config.Data["proxy"] = output.String()

	_, err = client.CoreV1().ConfigMaps("kube-system").Update(ctx, config, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cannot save proxy config: %s", err)
	}

	ui.Printf("Requests for service '%s' will now be served by the packaged service.", serviceName)

	return nil
}

func ShowRoutes() error {
	ctx := context.Background()

	client := kube.Client()
	config, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "dev-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot read current dev config: %s", err)
	}

	proxyConfigs := map[string]ProxyConfig{}

	proxyString := config.Data["proxy"]
	if proxyString != "" {
		decoder := yaml.NewDecoder(strings.NewReader(proxyString))
		if err := decoder.Decode(proxyConfigs); err != nil {
			return fmt.Errorf("cannt parse proxy config: %s", err)
		}
	}

	ui.Println("Services rerouted to external systems")
	ui.Println("-----------------------------------------")

	if len(proxyConfigs) == 0 {
		ui.Println("NONE")
	} else {
		for serviceName, proxyConfig := range proxyConfigs {
			ui.Printf("%s -> http://%s:%d\n", serviceName, proxyConfig.TargetHost, proxyConfig.TargetPort)
		}

		ui.Println("")
		ui.Printf("Reroutes can be removed with:\n  %s dev remove-route --service <service>\n", environment.SystemConfig.ManagerFilename)
	}

	return nil
}

type ProxyConfig struct {
	TargetHost string `yaml:"targetHost"`
	TargetPort int    `yaml:"targetPort"`
}
