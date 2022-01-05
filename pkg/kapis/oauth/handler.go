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
	"fmt"

	"github.com/emicklei/go-restful"

	"kubesphere.io/devops/pkg/api"
	"kubesphere.io/devops/pkg/models/auth"
)

const (
	KindTokenReview = "TokenReview"
)

type Spec struct {
	Token string `json:"token" description:"access token"`
}

type Status struct {
	Authenticated bool                   `json:"authenticated" description:"is authenticated"`
	User          map[string]interface{} `json:"user,omitempty" description:"user info"`
}

type TokenReview struct {
	APIVersion string  `json:"apiVersion" description:"Kubernetes API version"`
	Kind       string  `json:"kind" description:"kind of the API object"`
	Spec       *Spec   `json:"spec,omitempty"`
	Status     *Status `json:"status,omitempty" description:"token review status"`
}

type LoginRequest struct {
	Username string `json:"username" description:"username"`
	Password string `json:"password" description:"password"`
}

func (request *TokenReview) Validate() error {
	if request.Spec == nil || request.Spec.Token == "" {
		return fmt.Errorf("token must not be null")
	}
	return nil
}

type handler struct {
	tokenOperator auth.TokenManagementInterface
}

func newHandler(tokenOperator auth.TokenManagementInterface) *handler {
	return &handler{tokenOperator: tokenOperator}
}

// TokenReview Implement webhook authentication interface
// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
func (h *handler) TokenReview(req *restful.Request, resp *restful.Response) {
	var tokenReview TokenReview

	err := req.ReadEntity(&tokenReview)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	if err = tokenReview.Validate(); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	authenticated, err := h.tokenOperator.Verify(tokenReview.Spec.Token)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	success := TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: KindTokenReview,
		Status: &Status{
			Authenticated: true,
			User:          map[string]interface{}{"username": authenticated.GetName(), "uid": authenticated.GetUID()},
		},
	}

	resp.WriteEntity(success)
}
