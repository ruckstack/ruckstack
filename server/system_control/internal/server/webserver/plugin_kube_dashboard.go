package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"net/http"
	"strings"
)

var authToken string

func init() {
	Register(&KubeDashboardPlugin{})
}

func (plugin *KubeDashboardPlugin) Name() string {
	return "Kubernetes Dashboard"
}

func (plugin *KubeDashboardPlugin) Start(router *gin.Engine, ctx context.Context) {
	proxy := &ServiceProxy{
		Namespace:   "ops",
		ServiceName: "kubernetes-dashboard",
		ModifyUrl: func(originalUrl string) string {
			return strings.Replace(originalUrl, "ops/kube", "", 1)
		},
		ModifyRequest: func(request *http.Request) {
			if authToken != "" {
				request.Header.Add("Authorization", "Bearer "+authToken)
			}
		},
	}

	go proxy.Start(ctx)
	go watchSecret(ctx)

	router.Any("ops/kube", proxy.RequestHandler)
	router.Any("ops/kube/*url", proxy.RequestHandler)
}

func watchSecret(ctx context.Context) {
	kubeClient := kube.Client()

	fullSecretName := "ops.admin-user-token"

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Secrets().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			secret := newObj.(*core.Secret)
			if !strings.HasPrefix(kube.FullName(secret.ObjectMeta), fullSecretName) {
				return
			}

			authToken = string(secret.Data["token"])
		},

		AddFunc: func(obj interface{}) {
			secret := obj.(*core.Secret)
			if !strings.HasPrefix(kube.FullName(secret.ObjectMeta), fullSecretName) {
				return
			}

			authToken = string(secret.Data["token"])
		},

		DeleteFunc: func(obj interface{}) {
			secret := obj.(*core.Secret)
			if !strings.HasPrefix(kube.FullName(secret.ObjectMeta), fullSecretName) {
				return
			}

			authToken = ""
		},
	})
	informer.Run(ctx.Done())
}

type KubeDashboardPlugin struct {
}
