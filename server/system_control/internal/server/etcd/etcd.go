package etcd

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/embed"
	"net/url"
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

	etcdClientUrl, err := url.Parse("http://" + environment.LocalConfig.BindAddress + ":2379")
	if err != nil {
		return err
	}

	etcdPeerUrl, err := url.Parse("http://" + environment.LocalConfig.BindAddress + ":2380")
	if err != nil {
		return err
	}

	cfg := embed.NewConfig()
	cfg.Name = environment.PackageConfig.Id
	cfg.Dir = environment.ServerHome + "/data/etcd"
	cfg.Logger = "zap"
	//cfg.ClientAutoTLS = true
	//cfg.PeerAutoTLS = true
	cfg.LogOutputs = []string{environment.ServerHome + "/logs/etcd.log"}
	cfg.ACUrls = []url.URL{
		*etcdClientUrl,
	}
	cfg.LCUrls = []url.URL{
		*etcdClientUrl,
	}
	cfg.APUrls = []url.URL{
		*etcdPeerUrl,
	}
	cfg.LPUrls = []url.URL{
		*etcdPeerUrl,
	}
	cfg.InitialClusterToken = "todo"
	cfg.InitialCluster = environment.PackageConfig.Id + "=http://" + environment.LocalConfig.BindAddress + ":2380"
	cfg.ClusterState = "new"

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

	client, err := clientv3.NewFromURL(fmt.Sprintf("http://%s:2379", environment.LocalConfig.BindAddress))
	if err != nil {
		return err
	}
	status, err := client.Status(parent, fmt.Sprintf("http://%s:2379/health", environment.LocalConfig.BindAddress))
	if err != nil {
		return err
	}

	ui.Println(status)
	//ui.Fatal(<-etcdServer.Err())
	return nil
}
