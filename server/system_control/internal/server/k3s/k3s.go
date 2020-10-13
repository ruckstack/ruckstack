package k3s

import (
	"context"
	"flag"
	"fmt"
	"github.com/rancher/k3s/pkg/cli/cmds"
	"github.com/rancher/k3s/pkg/cli/server"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/urfave/cli"
)

var startupChannel chan error

func Start() error {

	//_, clusterIPNet, _ := net.ParseCIDR("10.42.0.0/16")
	//_, serviceIPNet, _ := net.ParseCIDR("10.43.0.0/16")

	cmds.ServerConfig = cmds.Server{
		DisableAgent: true,
		DataDir:      environment.ServerHome + "/data",
		ClusterCIDR:  "10.42.0.0/16",
		ServiceCIDR:  "10.43.0.0/16",
		ExtraControllerArgs: []string{
			fmt.Sprintf("flex-volume-plugin-dir=%s/data/kubelet-plugins/volume/exec", environment.ServerHome),
		},
		DefaultLocalStoragePath: environment.ServerHome + "/data/local-storage",

		AgentToken:               "",
		AgentTokenFile:           "",
		Token:                    "",
		TokenFile:                "",
		ClusterSecret:            "",
		ClusterDNS:               "",
		ClusterDomain:            "",
		HTTPSPort:                0,
		SupervisorPort:           0,
		APIServerPort:            0,
		APIServerBindAddress:     "",
		KubeConfigOutput:         "",
		KubeConfigMode:           "",
		TLSSan:                   nil,
		BindAddress:              "",
		ExtraAPIArgs:             nil,
		ExtraSchedulerArgs:       nil,
		ExtraCloudControllerArgs: nil,
		Rootless:                 false,
		DatastoreEndpoint:        "https://localhost:2379",
		DatastoreCAFile:          "",
		DatastoreCertFile:        "",
		DatastoreKeyFile:         "",
		AdvertiseIP:              "",
		AdvertisePort:            0,
		DisableScheduler:         false,
		ServerURL:                "",
		FlannelBackend:           "",
		DisableCCM:               false,
		DisableNPC:               false,
		DisableKubeProxy:         false,
		ClusterInit:              false,
		ClusterReset:             false,
		ClusterResetRestorePath:  "",
		EncryptSecrets:           false,
		StartupHooks:             nil,
		EtcdDisableSnapshots:     false,
		EtcdSnapshotDir:          "",
		EtcdSnapshotCron:         "",
		EtcdSnapshotRetention:    0,
	}

	app := cli.NewApp()
	ctx := cli.NewContext(app, &flag.FlagSet{}, nil)
	server.Run(ctx)

	return nil

}

//func StartX() error {
//	log.Println("Starting K3S...")
//
//	//if err := os.MkdirAll(environment.ServerHome+"/logs", os.FileMode(0755)); err != nil {
//	//	return fmt.Errorf("cannot create logs directory: %s", err)
//	//}
//
//	//k3sLogs, err := os.OpenFile(environment.ServerHome+"/logs/k3s.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
//	//if err != nil {
//	//	return err
//	//}
//	//
//	//defer k3sLogs.Close()
//
//	localConfig := environment.LocalConfig
//
//	_, clusterIPNet, _ := net.ParseCIDR("10.42.0.0/16")
//	_, serviceIPNet, _ := net.ParseCIDR("10.43.0.0/16")
//
//	var startupHooks []func(context context.Context, channel <-chan struct{}, x string) error
//	startupHooks = append(startupHooks, onStartupHook)
//
//	startupChannel = make(chan error)
//	ctx := context.Background()
//
//	err := server.StartServer(ctx, &server.Config{
//		DisableAgent:     true,
//		DisableServiceLB: false,
//		ControlConfig: config.Control{
//			AdvertisePort:        6443,
//			AdvertiseIP:          localConfig.BindAddress,
//			HTTPSPort:            6443,
//			SupervisorPort:       0,
//			APIServerPort:        6444,
//			APIServerBindAddress: localConfig.BindAddress,
//			AgentToken:           "",
//			Token:                "",
//			ClusterIPRange:       clusterIPNet,
//			ServiceIPRange:       serviceIPNet,
//			ClusterDNS:           net.ParseIP("10.43.0.10"),
//			ClusterDomain:        "cluster.local",
//			NoCoreDNS:            false,
//			KubeConfigOutput:     kubeclient.KubeconfigFile,
//			KubeConfigMode:       "0640",
//			DataDir:              environment.ServerHome + "/data/k3s",
//			Skips: map[string]bool{
//				"traefik": true,
//			},
//			Disables: nil,
//			//Datastore:                endpoint.Config{},
//			NoScheduler:              false,
//			ExtraAPIArgs:             nil,
//			ExtraControllerArgs:      nil,
//			ExtraCloudControllerArgs: nil,
//			ExtraSchedulerAPIArgs:    nil,
//			NoLeaderElect:            false,
//			JoinURL:                  "",
//			FlannelBackend:           "",
//			IPSECPSK:                 "",
//			DefaultLocalStoragePath:  environment.ServerHome + "/data/local-storage",
//			DisableCCM:               false,
//			DisableNPC:               false,
//			DisableKubeProxy:         false,
//			ClusterInit:              false,
//			ClusterReset:             false,
//			ClusterResetRestorePath:  "",
//			EncryptSecrets:           false,
//			TLSMinVersion:            0,
//			TLSCipherSuites:          nil,
//			EtcdDisableSnapshots:     false,
//			EtcdSnapshotDir:          "",
//			EtcdSnapshotCron:         "",
//			EtcdSnapshotRetention:    0,
//			BindAddress:              "",
//			SANs:                     nil,
//			Runtime:                  nil,
//		},
//		Rootless:       false,
//		SupervisorPort: 0,
//		StartupHooks:   startupHooks,
//	})
//	//select {
//	//case <-ctx.Done():
//	//	fmt.Sprintln("Startup DONE")
//	//	return ctx.Err()
//	//case err := <-startupChannel:
//	//	fmt.Sprintf("Got error %s", err)
//	//	return err
//	//}
//	//k3sCommand := "server"
//	//if localConfig.Join.Server != "" {
//	//	log.Printf("Joining server %s", localConfig.Join.Server)
//	//	k3sCommand = "agent"
//	//}
//	//
//	//k3sArgs := []string{
//	//	k3sCommand,
//	//	"--node-external-ip", localConfig.BindAddress,
//	//	"--data-dir", environment.ServerHome + "/data",
//	//	"--kubelet-arg", "root-dir=" + environment.ServerHome + "/data/kubelet",
//	//	//"--flannel-conf", common.ServerHome() + "/config/flannel.env",
//	//	"--flannel-iface", bindAddressIface,
//	//}
//	//
//	//if localConfig.Join.Server == "" {
//	//	k3sArgs = append(k3sArgs,
//	//		"--bind-address", localConfig.BindAddress,
//	//		"--no-deploy", "traefik",
//	//		"--default-local-storage-path", environment.ServerHome+"/data/local-storage",
//	//		"--write-kubeconfig", kubecConfigFile,
//	//		"--write-kubeconfig-mode", "640",
//	//	)
//	//} else {
//	//	k3sArgs = append(k3sArgs,
//	//		"--server", "https://"+localConfig.Join.Server+":6443",
//	//		"--token", localConfig.Join.Token,
//	//		"--node-ip", localConfig.BindAddress,
//	//	)
//	//}
//	//k3sStartCommand = exec.Command(environment.ServerHome+"/lib/k3s", k3sArgs...)
//	//k3sStartCommand.Stdout = k3sLogs
//	//k3sStartCommand.Stderr = k3sLogs
//	//if err := k3sStartCommand.Start(); err != nil {
//	//	return err
//	//}
//
//	//stat, err := os.Stat(kubecConfigFile)
//	//if err != nil {
//	//	return err
//	//}
//	//
//	//for stat == nil {
//	//	time.Sleep(10 * time.Second)
//	//	stat, err = os.Stat(kubecConfigFile)
//	//}
//	//
//	//if err := environment.PackageConfig.CheckFilePermissions("config/kubeconfig.yaml", environment.LocalConfig, environment.ServerHome); err != nil {
//	//	return err
//	//}
//
//	log.Println("Starting K3S...Complete")
//
//	return err
//}

func Stop() {
	//log.Println("Shutting down...")
	//kill := k3sStartCommand.Process.Kill()
	//if kill != nil {
	//	log.Printf("Error stopping k3s: %s", kill.Error())
	//}
	//err := k3sStartCommand.Wait()
	//if err != nil {
	//	log.Printf("Error waiting for stopping k3s: %s", err.Error())
	//}
}

func onStartupHook(ctx context.Context, apiServerReady <-chan struct{}, adminPath string) error {

	//fmt.Println("-----")
	//fmt.Println(ctx)
	//fmt.Println(apiServerReady)
	//fmt.Println(adminPath)
	//fmt.Println("-----")
	//
	//select {
	//case <-ctx.Done():
	//	fmt.Sprintln("Startup Hook DONE")
	//	return ctx.Err()
	//case value := <-apiServerReady:
	//	fmt.Printf("Admin server is ready %s", value)
	//
	//	startupChannel <- nil
	//}

	return nil

}
