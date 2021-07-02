/*
Copyright 2020 The KubeSphere Authors.

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

package oauth

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"

	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/constants"
	"kubesphere.io/devops/pkg/models/auth"
)

func AddToContainer(c *restful.Container, tokenOperator auth.TokenManagementInterface) error {
	ws := &restful.WebService{}
	ws.Path("/oauth").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	handler := newHandler(tokenOperator)

	// Implement webhook authentication interface
	// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
	ws.Route(ws.POST("/authenticate").
		Doc("TokenReview attempts to authenticate a token to a known user. Note: TokenReview requests may be "+
			"cached by the webhook token authenticator plugin in the kube-apiserver.").
		Reads(TokenReview{}).
		To(handler.TokenReview).
		Returns(http.StatusOK, api.StatusOK, TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AuthenticationTag}))
	c.Add(ws)
	return nil
}
