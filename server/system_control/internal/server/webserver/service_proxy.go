package webserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/k3s"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ServiceProxy struct {
	Namespace   string
	ServiceName string
	ModifyUrl   func(string) string

	reverseProxy *httputil.ReverseProxy
}

func (proxy *ServiceProxy) RequestHandler(ctx *gin.Context) {
	if proxy.reverseProxy != nil && monitor.ServerStatus.SystemReady {
		proxy.reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
	} else {
		err := showSiteDownPage(ctx.Writer)
		if err != nil {
			_ = ctx.Error(err)
		}
	}
}

func (proxy *ServiceProxy) Start(ctx context.Context) {
	k3s.WatchService(k3s.ServiceWatcher{
		Namespace:   proxy.Namespace,
		ServiceName: proxy.ServiceName,
		FireIpChanged: func(newIp string) {
			logger.Printf("Service %s.%s IP is now %s. Updated proxy", proxy.Namespace, proxy.ServiceName, newIp)

			if newIp == "" {
				proxy.reverseProxy = nil
				return
			}

			internalUrl, err := url.Parse(fmt.Sprintf("http://%s", newIp))
			if err != nil {
				logger.Printf("ERROR: %s", err)
			}

			proxy.reverseProxy = httputil.NewSingleHostReverseProxy(internalUrl)
			proxy.reverseProxy.ErrorHandler = func(response http.ResponseWriter, request *http.Request, err error) {
				if err.Error() == "Gateway Error" {
					if err := showSiteDownPage(response); err != nil {
						logger.Printf("ERROR: %s", err)
					}
				}
			}

			proxy.reverseProxy.Director = func(request *http.Request) {
				if request.URL.Path == "/favicon.ico" {
					return
				}

				if proxy.ModifyUrl != nil {
					modifiedUrl := proxy.ModifyUrl(request.URL.Path)
					modifiedUrl = strings.ReplaceAll(modifiedUrl, "//", "/")
					request.URL, err = internalUrl.Parse(modifiedUrl)
					if err != nil {
						panic(err)
					}
				}

				if !strings.HasPrefix(request.URL.Path, "http") {
					request.URL, err = internalUrl.Parse(request.URL.Path)
					if err != nil {
						panic(err)
					}
				}
			}

			proxy.reverseProxy.ModifyResponse = func(response *http.Response) error {
				if response.StatusCode == 502 || response.StatusCode == 503 || response.StatusCode == 504 {
					return errors.New("Gateway Error")
				}
				return nil
			}
		},
	}, ctx)

}
