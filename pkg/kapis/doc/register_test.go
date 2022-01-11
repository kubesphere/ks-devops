package doc

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

	AddSwaggerService(nil, restful.DefaultContainer)

	type args struct {
		method string
		uri    string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "check /apidocs.json",
		args: args{
			method: http.MethodGet,
			uri:    "/apidocs.json",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest, _ := http.NewRequest(tt.args.method,
				"http://fake.com"+tt.args.uri, bytes.NewBuffer([]byte("{}")))
			httpRequest.Header.Set("Content-Type", "application/json")
			restful.DefaultContainer.Dispatch(httpWriter, httpRequest)
			assert.Equal(t, 200, httpWriter.Code)
		})
	}
}
