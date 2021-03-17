package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"strings"
)

func init() {
	Register(&TraefikDashboardPlugin{})
}

func (plugin *TraefikDashboardPlugin) Name() string {
	return "Traefik Dashboard"
}

func (plugin *TraefikDashboardPlugin) Start(router *gin.Engine, ctx context.Context) {
	proxy := &ServiceProxy{
		Namespace:   "kube-system",
		ServiceName: "traefik-dashboard",
		ModifyUrl: func(originalUrl string) string {
			return strings.Replace(originalUrl, "ops/traefik", "", 1)
		},
	}

	go proxy.Start(ctx)

	router.Any("ops/traefik/*url", proxy.RequestHandler)
}

type TraefikDashboardPlugin struct {
}
