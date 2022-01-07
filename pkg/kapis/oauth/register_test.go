package oauth

import (
	"bytes"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIsExist(t *testing.T) {
	httpWriter := httptest.NewRecorder()

	AddToContainer(restful.DefaultContainer, nil)

	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "authenticate",
		args: args{
			method: http.MethodPost,
			uri:    "/authenticate",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com/oauth"+tt.args.uri, bytes.NewBuffer([]byte("{}")))
			httpRequest.Header.Set("Content-Type", "application/json")
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.NotEqual(t, httpWriter.Code, 404)
		})
	}
}
