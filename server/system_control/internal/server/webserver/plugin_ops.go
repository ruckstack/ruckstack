package webserver

import (
	"context"
	"embed"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"strings"
)

//go:embed ops/*
var embeddedOpsFiles embed.FS

func init() {
	Register(&OpsPlugin{})
}

func (plugin *OpsPlugin) Name() string {
	return "Ops"
}

func (plugin *OpsPlugin) Start(router *gin.Engine, ctx context.Context) {
}

func serveOpsPage(ctx *gin.Context) {
	path := ctx.Request.URL.Path
	path = strings.TrimPrefix(path, "/")
	if path == "ops" {
		path = "ops/"
	}
	if path == "ops/" {
		path = "ops/index.html"
	}

	file, err := embeddedOpsFiles.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			if !strings.Contains(path, ".") { //assume it's a pushstate route
				file, err = embeddedOpsFiles.Open("ops/index.html")
			}
		}
	}

	if err == nil {
		defer file.Close()

		ctx.Status(200)
		ctx.Writer.WriteHeader(200)
		if strings.HasSuffix(path, ".css") {
			ctx.Writer.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".js") {
			ctx.Writer.Header().Set("Content-Type", "text/javascript")
		}

		_, err = io.Copy(ctx.Writer, file)
		if err == nil {
			return
		} else {
		}
	} else {
		if os.IsNotExist(err) {
			ctx.Error(err)
			ctx.Writer.WriteHeader(404)
			return
		} else {
			ctx.Error(err)
			ctx.Writer.WriteHeader(500)
			return

		}
	}
}

type OpsPlugin struct {
}
