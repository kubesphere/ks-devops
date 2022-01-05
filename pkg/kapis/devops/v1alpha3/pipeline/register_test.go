package pipeline

import (
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/apiserver/runtime"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestAPIsExist(t *testing.T) {
	httpWriter := httptest.NewRecorder()
	wsWithGroup := runtime.NewWebService(v1alpha1.GroupVersion)

	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	RegisterRoutes(wsWithGroup, fake.NewFakeClientWithScheme(schema))
	restful.DefaultContainer.Add(wsWithGroup)

	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "get branches from the pipeline",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/pipelines/fake/branches",
		},
	}, {
		name: "get a branch from the pipeline",
		args: args{
			method: http.MethodGet,
			uri:    "/namespaces/fake/pipelines/fake/branches/fake",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha1"+tt.args.uri, nil)
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.NotEqual(t, httpWriter.Code, 404)
		})
	}
}
