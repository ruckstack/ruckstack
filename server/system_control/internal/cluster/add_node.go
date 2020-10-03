package cluster

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func AddNode() error {

	localConfig := environment.LocalConfig
	if localConfig.Join.Server != "" {
		panic("Must run this command from the primary machine in your cluster")
	}

	tokenFileContent, err := ioutil.ReadFile(filepath.Join(environment.ServerHome, "/data/server/token"))
	if err != nil {
		return err
	}

	kubeConfigFileContent, err := ioutil.ReadFile(filepath.Join(environment.ServerHome, "/config/kubeconfig.yaml"))
	if err != nil {
		return err
	}

	token := new(config.AddNodeToken)
	token.Token = strings.TrimSpace(string(tokenFileContent))
	token.Server = localConfig.BindAddress
	token.KubeConfig = strings.TrimSpace(string(kubeConfigFileContent))

	var yamlToken bytes.Buffer

	tokenEncoder := yaml.NewEncoder(&yamlToken)
	if err := tokenEncoder.Encode(token); err != nil {
		return err
	}

	encodedToken := base64.StdEncoding.EncodeToString(yamlToken.Bytes())

	fmt.Printf("To add additional nodes to the cluster, run the installer and choose \"Join an existing cluster\".\n\n")
	fmt.Printf("Join Token:\n")
	fmt.Println(encodedToken)

	return nil
}
