package webserver

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var logger *log.Logger

var sslCertFilePath = environment.ServerHome + "/data/ssl-cert.crt"
var sslKeyFilePath = environment.ServerHome + "/data/ssl-private.key"

var plugins []WebserverPlugin

func Register(plugin WebserverPlugin) {
	plugins = append(plugins, plugin)
}

func Start(ctx context.Context) error {

	logFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "webserver.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening webserver.log: %s", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)

	routerLogFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "webserver.access.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening webserver.access.log: %s", err)
	}

	logger.Println("Starting webserver")
	ui.Println("Starting webserver...")

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: routerLogFile,
	}))

	for _, plugin := range plugins {
		logger.Printf("starting plugin %s", plugin)
		go plugin.Start(router, ctx)
	}

	httpsSupported := false
	_, err = os.Stat(sslKeyFilePath)
	if err == nil {
		_, err = os.Stat(sslCertFilePath)
		if err == nil {
			go func() {
				httpsSupported = true
				logger.Println("Starting listener on port 443")
				if err := http.ListenAndServeTLS(":443",
					sslCertFilePath,
					sslKeyFilePath,
					router); err != nil {
					e := fmt.Errorf("error starting webserver listener on port 443: %s", err)
					logger.Println(e)
					//ui.Fatal(e)
				}
			}()
		} else {
			logger.Printf("Not starting https, cannot use certificate in %s: %s", sslCertFilePath, err)
		}
	} else {
		logger.Printf("Not starting https, cannot use key in %s: %s", sslKeyFilePath, err)
	}

	go func() {
		var handler http.Handler
		if httpsSupported {
			handler = http.HandlerFunc(redirectToHttps)
		} else {
			handler = router
		}

		logger.Println("Starting listener on port 80")
		if err := http.ListenAndServe(":80", handler); err != nil {
			logger.Println(fmt.Errorf("error starting webserver listener on port 80: %s", err))
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			logger.Println("Stopping webserver...")
			logger.Println("Stopping webserver...DONE")
		}
	}()

	return nil
}

func serveLocalFile(res http.ResponseWriter, url string) error {
	siteDownFile, err := os.Open(environment.ServerHome + "/data/web" + url)
	if err == nil {
		defer siteDownFile.Close()

		_, err = io.Copy(res, siteDownFile)
		if err != nil {
			return err
		}
	} else {
		res.WriteHeader(404)
	}
	return nil
}

func showSiteDownPage(res http.ResponseWriter) error {
	return serveLocalFile(res, "/site-down.html")
}

func redirectToHttps(responseWriter http.ResponseWriter, requeset *http.Request) {
	target := "https://" + requeset.Host + requeset.URL.Path
	if len(requeset.URL.RawQuery) > 0 {
		target += "?" + requeset.URL.RawQuery
	}
	http.Redirect(responseWriter, requeset, target, http.StatusMovedPermanently)
}

type WebserverPlugin interface {
	Name() string
	Start(router *gin.Engine, ctx context.Context)
}
