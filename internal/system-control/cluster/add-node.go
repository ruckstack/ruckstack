package cluster

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/ruckstack/ruckstack/internal/installer"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func AddNode() {

	localConfig, err := util.GetLocalConfig()
	util.Check(err)
	if localConfig.Join.Server != "" {
		panic("Must run this command from the primary machine in your cluster")
	}

	tokenFileContent, err := ioutil.ReadFile(filepath.Join(util.InstallDir(), "/data/server/token"))
	util.Check(err)

	kubeConfigFileContent, err := ioutil.ReadFile(filepath.Join(util.InstallDir(), "/config/kubeconfig.yaml"))
	util.Check(err)

	token := new(installer.AddNodeToken)
	token.Token = strings.TrimSpace(string(tokenFileContent))
	token.Server = localConfig.BindAddress
	token.KubeConfig = strings.TrimSpace(string(kubeConfigFileContent))

	var yamlToken bytes.Buffer

	tokenEncoder := yaml.NewEncoder(&yamlToken)
	util.Check(tokenEncoder.Encode(token))

	encodedToken := base64.StdEncoding.EncodeToString(yamlToken.Bytes())

	fmt.Printf("To add additional nodes to the cluster, run the installer and choose \"Join an existing cluster\".\n\n")
	fmt.Printf("Join Token:\n")
	fmt.Println(encodedToken)
}
