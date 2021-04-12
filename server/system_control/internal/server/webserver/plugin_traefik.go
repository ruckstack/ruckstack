package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"strings"
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

	router.NoRoute(func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.URL.Path, "/ops") {
			serveOpsPage(ctx)
		} else {
			proxy.RequestHandler(ctx)
		}
	})
}

type TraefikPlugin struct {
}
