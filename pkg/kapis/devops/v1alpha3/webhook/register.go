package webhook

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RegisterWebhooks registers all webhooks into web service.
func RegisterWebhooks(genericClient client.Client, ws *restful.WebService) {
	webhookHandler := NewHandler(genericClient)
	ws.Route(ws.POST("/webhooks/jenkins").
		To(webhookHandler.ReceiveEventsFromJenkins).
		Doc("Webhook for receiving events from Jenkins").
		Returns(http.StatusOK, api.StatusOK, nil))
}
