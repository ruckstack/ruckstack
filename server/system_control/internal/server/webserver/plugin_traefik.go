package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
)

func init() {
	Register(&TraefikPlugin{})
}

func (plugin *TraefikPlugin) Name() string {
	return "Traefik Proxy"
}

func (plugin *TraefikPlugin) Start(router *gin.Engine, ctx context.Context) {
	proxy := &ServiceProxy{
		Namespace:   "kube-system",
		ServiceName: "traefik",
	}

	go proxy.Start(ctx)

	router.NoRoute(proxy.RequestHandler)
}

type TraefikPlugin struct {
}
