package argocd

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/gitops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/kapis/common"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestRegisterRoutes(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	type args struct {
		service *restful.WebService
		options *common.Options
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, service *restful.WebService)
	}{{
		name: "normal case",
		args: args{
			service: runtime.NewWebService(v1alpha1.GroupVersion),
			options: &common.Options{GenericClient: fake.NewFakeClientWithScheme(schema)},
		},
		verify: func(t *testing.T, service *restful.WebService) {
			assert.Greater(t, len(service.Routes()), 0)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterRoutes(tt.args.service, tt.args.options, &config.ArgoCDOption{})
			tt.verify(t, tt.args.service)
		})
	}
}

func TestAPIs(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)
	err = v1.SchemeBuilder.AddToScheme(schema)
	assert.Nil(t, err)

	app := v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "app",
			Labels: map[string]string{
				v1alpha1.HealthStatusLabelKey: "Healthy",
				v1alpha1.SyncStatusLabelKey:   "Synced",
			},
		},
	}

	nonArgoClusterSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "non-argo-cluster",
			Namespace: "ns",
		},
	}
	invalidArgoClusterSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-argo-cluster",
			Namespace: "ns",
			Labels: map[string]string{
				"argocd.argoproj.io/secret-type": "cluster",
			},
		},
		Data: map[string][]byte{
			"server": []byte("server"),
		},
	}
	validArgoClusterSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argo-cluster",
			Namespace: "ns",
			Labels: map[string]string{
				"argocd.argoproj.io/secret-type": "cluster",
			},
		},
		Data: map[string][]byte{
			"name":   []byte("name"),
			"server": []byte("server"),
		},
	}

	type request struct {
		method string
		uri    string
		body   func() io.Reader
	}
	tests := []struct {
		name         string
		request      request
		responseCode int
		k8sclient    client.Client
		verify       func(t *testing.T, body []byte)
	}{{
		name: "get clusters, no expected data",
		request: request{
			method: http.MethodGet,
			uri:    "/clusters",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, nonArgoClusterSecret.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			assert.Equal(t, `[
 {
  "server": "https://kubernetes.default.svc",
  "name": "in-cluster"
 }
]`, string(body))
		},
	}, {
		name: "get clusters, have invalid data",
		request: request{
			method: http.MethodGet,
			uri:    "/clusters",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, invalidArgoClusterSecret.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			assert.Equal(t, `[
 {
  "server": "https://kubernetes.default.svc",
  "name": "in-cluster"
 }
]`, string(body))
		},
	}, {
		name: "get clusters, have the expected data",
		request: request{
			method: http.MethodGet,
			uri:    "/clusters",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, validArgoClusterSecret.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			assert.Equal(t, `[
 {
  "server": "https://kubernetes.default.svc",
  "name": "in-cluster"
 },
 {
  "server": "server",
  "name": "name"
 }
]`, string(body))
		},
	}, {
		name: "get applications summary",
		request: request{
			method: http.MethodGet,
			uri:    "/namespaces/ns/application-summary",
		},
		k8sclient:    fake.NewFakeClientWithScheme(schema, app.DeepCopy()),
		responseCode: http.StatusOK,
		verify: func(t *testing.T, body []byte) {
			assert.JSONEq(t, `{"total": 1, "healthStatus": { "Healthy": 1 }, "syncStatus": { "Synced": 1 }}`, string(body))
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsWithGroup := runtime.NewWebService(v1alpha1.GroupVersion)
			RegisterRoutes(wsWithGroup, &common.Options{GenericClient: tt.k8sclient}, &config.ArgoCDOption{})

			container := restful.NewContainer()
			container.Add(wsWithGroup)

			api := fmt.Sprintf("http://fake.com/kapis/gitops.kubesphere.io/%s%s", v1alpha1.GroupVersion.Version, tt.request.uri)
			var body io.Reader
			if tt.request.body != nil {
				body = tt.request.body()
			}
			req, err := http.NewRequest(tt.request.method, api, body)
			req.Header.Set("Content-Type", "application/json")
			assert.Nil(t, err)

			httpWriter := httptest.NewRecorder()
			container.Dispatch(httpWriter, req)
			assert.Equal(t, tt.responseCode, httpWriter.Code)

			if tt.verify != nil {
				tt.verify(t, httpWriter.Body.Bytes())
			}
		})
	}
}
