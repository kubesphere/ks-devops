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

package api

import (
	"net/http"
	"runtime"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

// Avoid emitting errors that look like valid HTML. Quotes are okay.
var sanitizer = strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;")

func HandleInternalError(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusInternalServerError, req, response, err)
}

// HandleBadRequest writes http.StatusBadRequest and log error
func HandleBadRequest(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusBadRequest, req, response, err)
}

func HandleNotFound(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusNotFound, req, response, err)
}

func HandleForbidden(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusForbidden, req, response, err)
}

func HandleUnauthorized(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusUnauthorized, req, response, err)
}

func HandleTooManyRequests(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusTooManyRequests, req, response, err)
}

func HandleConflict(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusConflict, req, response, err)
}

func HandleError(request *restful.Request, response *restful.Response, err error) {
	var statusCode int
	switch t := err.(type) {
	case errors.APIStatus:
		statusCode = int(t.Status().Code)
	case restful.ServiceError:
		statusCode = t.Code
	default:
		statusCode = http.StatusInternalServerError
	}
	handle(statusCode, request, response, err)
}

func handle(statusCode int, req *restful.Request, response *restful.Response, err error) {
	_, fn, line, _ := runtime.Caller(2)
	klog.Errorf("%s:%d %v", fn, line, err)
	http.Error(response, sanitizer.Replace(err.Error()), statusCode)
}
