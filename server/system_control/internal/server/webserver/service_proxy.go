package webserver

import (
	"context"
	"crypto/tls"
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
	Namespace     string
	ServiceName   string
	Protocol      string
	ModifyUrl     func(string) string
	ModifyRequest func(request *http.Request)

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
	if proxy.Protocol == "" {
		proxy.Protocol = "http"
	}

	k3s.WatchService(k3s.ServiceWatcher{
		Namespace:   proxy.Namespace,
		ServiceName: proxy.ServiceName,
		FireIpChanged: func(newIp string) {
			logger.Printf("Service %s.%s IP is now %s. Updated proxy", proxy.Namespace, proxy.ServiceName, newIp)

			if newIp == "" {
				proxy.reverseProxy = nil
				return
			}

			internalUrl, err := url.Parse(fmt.Sprintf("%s://%s", proxy.Protocol, newIp))
			if err != nil {
				logger.Printf("ERROR: %s", err)
			}

			proxy.reverseProxy = httputil.NewSingleHostReverseProxy(internalUrl)

			proxy.reverseProxy.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}

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

				finalPath := request.URL.Path
				if proxy.ModifyUrl != nil {
					finalPath = proxy.ModifyUrl(finalPath)
					finalPath = strings.ReplaceAll(finalPath, "//", "/")
				}

				if strings.HasPrefix(finalPath, "http") {
					request.URL.Path = finalPath
				} else {
					request.URL, err = internalUrl.Parse(finalPath)
					if err != nil {
						panic(err)
					}
				}

				request.Header.Del("Authorization")

				if proxy.ModifyRequest != nil {
					proxy.ModifyRequest(request)
				}
			}

			proxy.reverseProxy.ModifyResponse = func(response *http.Response) error {
				if response.StatusCode == 502 || response.StatusCode == 503 || response.StatusCode == 504 {
					return errors.New("Gateway Error")
				}
				if response.StatusCode == 401 {
					response.StatusCode = 500
				}
				return nil
			}
		},
	}, ctx)

}
