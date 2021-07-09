package filters

import (
	"kubesphere.io/devops/pkg/constants"
	"kubesphere.io/devops/pkg/server/errors"
	"net/http"
	"testing"
)

func Test_injectToken(t *testing.T) {
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name       string
		args       args
		assertFunc func(req *http.Request) error
	}{{
		name: "without header in request",
		args: args{
			req: &http.Request{},
		},
		assertFunc: func(req *http.Request) error {
			return nil
		},
	}, {
		name: "with auth header in request",
		args: args{
			req: &http.Request{
				Header: map[string][]string{
					"Authorization": {"bearer fake-token"},
				},
			},
		},
		assertFunc: func(req *http.Request) error {
			if token := req.Context().Value(constants.K8SToken); token == "" {
				return errors.New("cannot found k8s token from request context")
			}
			return nil
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRequestWithTokenContext := injectToken(tt.args.req)
			if err := tt.assertFunc(gotRequestWithTokenContext); err != nil {
				t.Errorf("injectToken() failed, error %v", err)
			}
		})
	}
}
