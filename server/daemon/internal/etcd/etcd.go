package etcd

import (
	"context"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"go.etcd.io/etcd/embed"
	"time"
)

var etcdServer *embed.Etcd

func Start(parent context.Context) error {
	var err error

	if etcdServer != nil {
		ui.VPrintln("Etcd already started")
		return nil
	}

	ui.Println("Starting etcd...")

	//Name:                e.name,
	//	InitialOptions:      options,
	//		ForceNewCluster:     forceNew,
	//		ListenClientURLs:    fmt.Sprintf(e.clientURL() + ",https://127.0.0.1:2379"),
	//		ListenMetricsURLs:   "http://127.0.0.1:2381",
	//		ListenPeerURLs:      e.peerURL(),
	//		AdvertiseClientURLs: e.clientURL(),
	//		DataDir:             dataDir(e.config),
	//		ServerTrust: executor.ServerTrust{
	//		CertFile:       e.config.Runtime.ServerETCDCert,
	//		KeyFile:        e.config.Runtime.ServerETCDKey,
	//		ClientCertAuth: true,
	//		TrustedCAFile:  e.config.Runtime.ETCDServerCA,
	//	},
	//		PeerTrust: executor.PeerTrust{
	//		CertFile:       e.config.Runtime.PeerServerClientETCDCert,
	//		KeyFile:        e.config.Runtime.PeerServerClientETCDKey,
	//		ClientCertAuth: true,
	//		TrustedCAFile:  e.config.Runtime.ETCDPeerCA,
	//	},
	//		ElectionTimeout:   5000,
	//		HeartbeatInterval: 500,

	go func() {
		select {
		case <-parent.Done():
			if etcdServer != nil {
				ui.Println("Stopping etcd...")
				etcdServer.Close()
				ui.Println("Stopping etcd...DONE")
			}
		}
	}()

	cfg := embed.NewConfig()
	cfg.Name = environment.PackageConfig.Id
	cfg.Dir = environment.ServerHome + "/data/etcd"
	cfg.Logger = "zap"
	cfg.LogOutputs = []string{environment.ServerHome + "/logs/etcd.log"}

	etcdServer, err = embed.StartEtcd(cfg)
	if err != nil {
		return err
	}

	select {
	case <-etcdServer.Server.ReadyNotify():
		ui.Println("Starting etcd...DONE")
	case <-time.After(60 * time.Second):
		etcdServer.Server.Stop() // trigger a shutdown
		ui.Fatalf("Timeout starting etcd...DONE")
	}

	//ui.Fatal(<-etcdServer.Err())
	return nil
}
