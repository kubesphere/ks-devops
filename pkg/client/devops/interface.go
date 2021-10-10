/*
Copyright 2020 KubeSphere Authors

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

package devops

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
)

type Interface interface {
	CredentialOperator

	BuildGetter

	PipelineOperator

	ProjectPipelineOperator

	ProjectOperator

	ConfigurationOperator
}

func GetDevOpsStatusCode(devopsErr error) int {
	errStr := strings.TrimPrefix(devopsErr.Error(), "unexpected status code: ")
	if code, err := strconv.Atoi(errStr); err == nil {
		message := http.StatusText(code)
		if !govalidator.IsNull(message) {
			return code
		}
	}
	if jErr, ok := devopsErr.(*ErrorResponse); ok {
		return jErr.Response.StatusCode
	}
	return http.StatusInternalServerError
}

type ErrorResponse struct {
	Body     []byte
	Response *http.Response
	Message  string
}

func (e *ErrorResponse) Error() string {
	var u string
	var method string
	if e.Response != nil && e.Response.Request != nil {
		req := e.Response.Request
		u = fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.RequestURI())
		method = req.Method
	}
	return fmt.Sprintf("%s %s: %d %s", method, u, e.Response.StatusCode, e.Message)
}
