package webserver

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"golang.org/x/crypto/bcrypt"
	"io"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var logger *log.Logger

var sslCertFilePath = environment.ServerHome + "/data/ssl-cert.crt"
var sslKeyFilePath = environment.ServerHome + "/data/ssl-private.key"

var opsUsers = map[string][]byte{}
var realmHeader = "Basic realm=" + strconv.Quote("Authorization Required")

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

	go startPasswordWatch(ctx)
	router.Use(authHandler)

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
		res.WriteHeader(200)

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

func startPasswordWatch(ctx context.Context) {
	client := kube.Client()

	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Core().V1().Secrets().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			secret := obj.(*core.Secret)
			if kube.FullName(secret.ObjectMeta) != "ops.ops-users" {
				return
			}

			opsUsers = secret.Data
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			secret := newObj.(*core.Secret)
			if kube.FullName(secret.ObjectMeta) != "ops.ops-users" {
				return
			}

			opsUsers = secret.Data

		},
		DeleteFunc: func(obj interface{}) {
			secret := obj.(*core.Secret)
			if kube.FullName(secret.ObjectMeta) != "ops.ops-users" {
				return
			}

			opsUsers = map[string][]byte{}
		},
	})

	informer.Run(ctx.Done())
}

func requiresAuth(ctx *gin.Context) bool {
	path := ctx.Request.URL.Path
	if !strings.HasPrefix(path, "/ops") {
		return false
	}

	if path == "/ops" ||
		path == "/ops/" ||
		path == "/ops/status" ||
		path == "/ops/api/status" ||
		path == "/ops/api/me" ||
		strings.HasPrefix(path, "/ops/assets") ||
		strings.HasPrefix(path, "/ops/public") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".css.map") ||
		strings.HasSuffix(path, ".js.map") ||
		strings.HasSuffix(path, ".ico") {
		return false
	}

	return true
}

func authHandler(ctx *gin.Context) {
	user, password, _ := ctx.Request.BasicAuth()

	found := false
	if user != "" && password != "" {
		var knownHash []byte
		knownHash, found = opsUsers[user]

		if found {
			err := bcrypt.CompareHashAndPassword(knownHash, []byte(password))
			found = err == nil
		}
	}

	if found {
		//set auth key since we know who they are
		ctx.Set(gin.AuthUserKey, user)
	}

	if !found && requiresAuth(ctx) {
		if user != "" {
			n := rand.Intn(8 * 1000)
			time.Sleep(time.Duration(n) * time.Millisecond)
		}
		// Credentials doesn't match, we return 401 and abort handlers chain.
		ctx.Header("WWW-Authenticate", realmHeader)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

}

type WebserverPlugin interface {
	Name() string
	Start(router *gin.Engine, ctx context.Context)
}
