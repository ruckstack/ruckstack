package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"net/http"
)

func (plugin *OpsApiPlugin) Start(router *gin.Engine, ctx context.Context) {
	router.GET("ops/api/status", getStatus)
	router.GET("ops/api/status/detailed", getDetailedStatus)
	router.GET("ops/api/me", getMe)
	router.GET("ops/api/login", getMe) //auth required version of getMe
	router.GET("ops/api/logout", getLogout)
}

func getLogout(ctx *gin.Context) {
	ctx.Header("WWW-Authenticate", realmHeader)
	ctx.AbortWithStatus(http.StatusUnauthorized)
}

func getMe(ctx *gin.Context) {
	username, _ := ctx.Get(gin.AuthUserKey)
	ctx.JSON(200, gin.H{
		"username": username,
	})
}

func getStatus(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"healthy":   monitor.ServerStatus.SystemReady,
		"name":      environment.PackageConfig.Name,
		"version":   environment.PackageConfig.Version,
		"support":   environment.PackageConfig.Support,
		"buildTime": environment.PackageConfig.BuildTime,
	})
}

func getDetailedStatus(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"healthy":   monitor.ServerStatus.SystemReady,
		"name":      environment.PackageConfig.Name,
		"version":   environment.PackageConfig.Version,
		"support":   environment.PackageConfig.Support,
		"buildTime": environment.PackageConfig.BuildTime,
		"trackers":  monitor.ServerStatus.Trackers,
	})
}

func init() {
	Register(&OpsApiPlugin{})
}

func (plugin *OpsApiPlugin) Name() string {
	return "Ops API"
}

type OpsApiPlugin struct {
}
