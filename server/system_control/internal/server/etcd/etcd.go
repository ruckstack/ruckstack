package etcd

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"go.etcd.io/etcd/embed"
	"time"
)

func Start() error {
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

	cfg := embed.NewConfig()
	cfg.Name = environment.PackageConfig.Id
	cfg.Dir = environment.ServerHome + "/data/etcd"

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return err
	}
	defer e.Close()
	select {
	case <-e.Server.ReadyNotify():
		ui.Printf("Server is ready!")
	case <-time.After(60 * time.Second):
		e.Server.Stop() // trigger a shutdown
		ui.Printf("Server took too long to start!")
	}
	//ui.Fatal(<-e.Err())
	return nil
}
