package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
)

func init() {
	Register(&OpsPlugin{})
}

func (plugin *OpsPlugin) Name() string {
	return "Ops"
}

func (plugin *OpsPlugin) Start(router *gin.Engine, ctx context.Context) {
	router.StaticFile("ops/", environment.ServerHome+"/data/web/ops/index.html")
	router.StaticFile("ops/test.html", environment.ServerHome+"/data/web/ops/test.html")
}

type OpsPlugin struct {
}
