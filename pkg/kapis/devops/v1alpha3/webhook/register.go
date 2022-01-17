/*
Copyright 2022 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
