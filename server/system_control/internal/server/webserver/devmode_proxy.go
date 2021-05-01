package webserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type DevModeProxy struct {
	targetHost string
	targetPort int

	reverseProxy *httputil.ReverseProxy
}

func (proxy *DevModeProxy) RequestHandler(ctx *gin.Context) {
	proxy.reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
}

func (proxy *DevModeProxy) Start(ctx context.Context) {
	proxyUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", proxy.targetHost, proxy.targetPort))
	if err == nil {
		logger.Printf("Error parsing target url: %s", err)
	}
	proxy.reverseProxy = httputil.NewSingleHostReverseProxy(proxyUrl)
	proxy.reverseProxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	proxy.reverseProxy.ErrorHandler = func(response http.ResponseWriter, request *http.Request, err error) {
		logger.Printf("PROXY ERROR: %s", err)

		if strings.Contains(err.Error(), "connection refused") {
			response.WriteHeader(504)
		}
		_, _ = response.Write([]byte(err.Error()))
	}

	proxy.reverseProxy.ModifyResponse = func(response *http.Response) error {
		if response.StatusCode == 502 || response.StatusCode == 503 || response.StatusCode == 504 {
			return errors.New("Proxy Error")
		}

		return nil
	}
}
