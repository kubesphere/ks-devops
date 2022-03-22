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

package common

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/devops/pkg/kapis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Options contain options needed by creating handlers.
type Options struct {
	GenericClient client.Client
}

var (
	// DevopsPathParameter is a path parameter definition for devops.
	DevopsPathParameter = restful.PathParameter("devops", "DevOps project name")
	// NamespacePathParameter is a path parameter definition for namespace
	NamespacePathParameter = restful.PathParameter("namespace", "The namespace name")
)

// GetPathParameter returns the path parameter value from a request
func GetPathParameter(req *restful.Request, param *restful.Parameter) string {
	return req.PathParameter(param.Data().Name)
}

// GetQueryParameter returns the query parameter value from a request
func GetQueryParameter(req *restful.Request, param *restful.Parameter) string {
	return req.QueryParameter(param.Data().Name)
}

// Response is a common response method
func Response(req *restful.Request, res *restful.Response, object interface{}, err error) {
	if err != nil {
		kapis.HandleError(req, res, err)
	} else {
		_ = res.WriteEntity(object)
	}
}
