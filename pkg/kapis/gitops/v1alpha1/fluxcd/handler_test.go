// Copyright 2022 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package fluxcd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/ks-devops/pkg/api/gitops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/kapis/gitops/v1alpha1/gitops"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_handler_applicationGet(t *testing.T) {
	createRequest := func(uri string) *restful.Request {
		fakeRequest := httptest.NewRequest(http.MethodGet, uri, nil)
		request := restful.NewRequest(fakeRequest)
		return request
	}

	type fields struct {
		Client client.Client
	}
	type args struct {
		req     *restful.Request
		secrets v1.SecretList
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		wantResponse []v1alpha1.ApplicationDestination
	}{
		{
			name: "get a member clusters",
			args: args{
				req: createRequest("/clusters"),
				secrets: v1.SecretList{
					Items: []v1.Secret{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "member-cluster",
								Namespace: "ns",
								Labels: map[string]string{
									"app.kubernetes.io/managed-by": "fluxcd-controller",
								},
							},
						},
					},
				},
			},
			wantResponse: []v1alpha1.ApplicationDestination{
				{
					Name: "in-cluster",
				},
				{
					Name: "member-cluster",
				},
			},
		},
		{
			name: "get host cluster",
			args: args{
				req: createRequest("/clusters"),
				secrets: v1.SecretList{
					Items: nil,
				},
			},
			wantResponse: []v1alpha1.ApplicationDestination{
				{
					Name: "in-cluster",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
			utilruntime.Must(v1.AddToScheme(scheme.Scheme))
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithLists(tt.args.secrets.DeepCopy()).Build()
			h := handler{Handler: &gitops.Handler{Client: fakeClient}}
			req := tt.args.req
			recorder := httptest.NewRecorder()
			resp := restful.NewResponse(recorder)
			resp.SetRequestAccepts(restful.MIME_JSON)
			h.getClusters(req, resp)
			assert.Equal(t, 200, recorder.Code)
			wantResponseBytes, err := json.Marshal(tt.wantResponse)
			assert.Nil(t, err)
			assert.JSONEq(t, string(wantResponseBytes), recorder.Body.String())
		})
	}
}
