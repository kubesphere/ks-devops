package v1alpha3

import (
	"context"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	fakeclientset "kubesphere.io/devops/pkg/client/clientset/versioned/fake"
	fakedevops "kubesphere.io/devops/pkg/client/devops/fake"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/constants"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestAPIsExist(t *testing.T) {
	httpWriter := httptest.NewRecorder()
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	AddToContainer(restful.DefaultContainer, fakedevops.NewFakeDevops(nil),
		k8s.NewFakeClientSets(k8sfake.NewSimpleClientset(), nil, nil, "", nil,
			fakeclientset.NewSimpleClientset(&v1alpha3.DevOpsProject{
				ObjectMeta: metav1.ObjectMeta{Name: "fake"},
			})),
		fake.NewFakeClientWithScheme(schema))

	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "credential list",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/credentials",
		},
	}, {
		name: "create a credential",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake/credentials",
		},
	}, {
		name: "get a credential",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/credentials/fake",
		},
	}, {
		name: "update a credential",
		args: args{
			method: http.MethodPut,
			uri:    "/devops/fake/credentials/fake",
		},
	}, {
		name: "delete a credential",
		args: args{
			method: http.MethodDelete,
			uri:    "/devops/fake/credentials/fake",
		},
	}, {
		name: "get pipeline list",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines",
		},
	}, {
		name: "create a pipeline",
		args: args{
			method: http.MethodPost,
			uri:    "/devops/fake/pipelines",
		},
	}, {
		name: "get a pipeline",
		args: args{
			method: http.MethodGet,
			uri:    "/devops/fake/pipelines/fake",
		},
	}, {
		name: "update a pipeline",
		args: args{
			method: http.MethodPut,
			uri:    "/devops/fake/pipelines/fake",
		},
	}, {
		name: "delete a pipeline",
		args: args{
			method: http.MethodDelete,
			uri:    "/devops/fake/pipelines/fake",
		},
	}, {
		name: "get devops list",
		args: args{
			method: http.MethodGet,
			uri:    "/workspaces/fake/devops",
		},
	}, {
		name: "create a devops",
		args: args{
			method: http.MethodPost,
			uri:    "/workspaces/fake/devops",
		},
	}, {
		name: "get a devops",
		args: args{
			method: http.MethodGet,
			uri:    "/workspaces/fake/devops/fake",
		},
	}, {
		name: "update a devops",
		args: args{
			method: http.MethodPut,
			uri:    "/workspaces/fake/devops/fake",
		},
	}, {
		name: "delete a devops",
		args: args{
			method: http.MethodDelete,
			uri:    "/workspaces/fake/devops/fake",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/kapis/devops.kubesphere.io/v1alpha3"+tt.args.uri, nil)
			httpRequest = httpRequest.WithContext(context.WithValue(context.TODO(), constants.K8SToken, constants.ContextKeyK8SToken("")))
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.NotEqual(t, httpWriter.Code, 404)
		})
	}
}
