package k3s

import (
	"context"
	"fmt"
	"github.com/rancher/k3s/pkg/daemons/config"
	"github.com/rancher/k3s/pkg/server"
	"github.com/rancher/kine/pkg/endpoint"
	"github.com/rancher/kine/pkg/tls"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path/filepath"
)

func Start(parent context.Context) error {

	k3sLogFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "k3s.server.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening k3s.server.log: %s", err)
	}

	////k3s doesn't use a context log, so it has to get the default output for logrus
	logrus.SetOutput(k3sLogFile)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	_, clusterIPNet, _ := net.ParseCIDR("10.42.0.0/16")
	_, serviceIPNet, _ := net.ParseCIDR("10.43.0.0/16")

	serverConfig := server.Config{
		DisableAgent: true,
		ControlConfig: config.Control{
			AdvertiseIP:          environment.LocalConfig.BindAddress,
			HTTPSPort:            6443,
			APIServerBindAddress: environment.LocalConfig.BindAddress,
			ClusterIPRange:       clusterIPNet,
			ServiceIPRange:       serviceIPNet,
			ClusterDNS:           net.ParseIP("10.43.0.10"),
			ClusterDomain:        "cluster.local",
			KubeConfigOutput:     filepath.Join(environment.ServerHome, "/config/kubeconfig.yaml"),
			KubeConfigMode:       "0640",
			DataDir:              environment.ServerHome + "/data/k3s",
			Skips: map[string]bool{
				"traefik":        true,
				"metrics-server": true,
			},
			DefaultLocalStoragePath: environment.ServerHome + "/data/local-storage",
			ExtraControllerArgs: []string{
				fmt.Sprintf("flex-volume-plugin-dir=%s/data/kubelet-plugins/volume/exec", environment.ServerHome),
				fmt.Sprintf("log-file=%s", filepath.Join(environment.ServerHome, "logs", "k3s.controller.log")),
				fmt.Sprintf("logtostderr=false"),
			},
			ExtraCloudControllerArgs: []string{
				fmt.Sprintf("log-file=%s", filepath.Join(environment.ServerHome, "logs", "k3s.cloudcontroller.log")),
				fmt.Sprintf("logtostderr=false"),
			},
			ExtraSchedulerAPIArgs: []string{
				fmt.Sprintf("log-file=%s", filepath.Join(environment.ServerHome, "logs", "k3s.scheduler.log")),
				fmt.Sprintf("logtostderr=false"),
			},
			ExtraAPIArgs: []string{
				fmt.Sprintf("log-file=%s", filepath.Join(environment.ServerHome, "logs", "k3s.apiserver.log")),
				fmt.Sprintf("logtostderr=false"),
			},

			Datastore: endpoint.Config{
				Endpoint: "http://localhost:2379",

				GRPCServer: nil,
				Listener:   "",
				Config:     tls.Config{},
			},
			BindAddress: "127.0.0.1",
			SANs: []string{
				"127.0.0.1",
				"192.168.1.27",
			},

			//AdvertisePort:        6443,
			//APIServerPort:        6444,

			//unset (so far) fields
			SupervisorPort:          0,
			AgentToken:              "",
			Token:                   "",
			NoCoreDNS:               false,
			Disables:                nil,
			NoScheduler:             false,
			NoLeaderElect:           false,
			JoinURL:                 "",
			FlannelBackend:          "",
			IPSECPSK:                "",
			DisableCCM:              false,
			DisableNPC:              false,
			DisableKubeProxy:        false,
			ClusterInit:             false,
			ClusterReset:            false,
			ClusterResetRestorePath: "",
			EncryptSecrets:          false,
			TLSMinVersion:           0,
			TLSCipherSuites:         nil,
			EtcdDisableSnapshots:    false,
			EtcdSnapshotDir:         "",
			EtcdSnapshotCron:        "",
			EtcdSnapshotRetention:   0,
			Runtime:                 nil,
		},

		DisableServiceLB: false,
		Rootless:         false,
		SupervisorPort:   0,
		StartupHooks:     nil,
	}

	if err := server.StartServer(parent, &serverConfig); err != nil {
		ui.Fatal(err)
	}

	go func() {
		select {
		case <-parent.Done():
			logrus.Warn("Stopping k3s...")
			logrus.Warn("Stopping k3s...DONE")
		}
	}()

	<-serverConfig.ControlConfig.Runtime.APIServerReady
	ui.VPrintln("APIServer is ready")

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
//			AgentToken:           "",
//			Token:                "",
//			NoCoreDNS:            false,
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
