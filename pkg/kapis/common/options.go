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
	"kubesphere.io/devops/pkg/apiserver/query"
	"kubesphere.io/devops/pkg/kapis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
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

	// PageNumberQueryParameter is a restful query parameter of the page number
	//
	// Deprecated
	PageNumberQueryParameter = restful.QueryParameter("pageNumber", "The number of paging").DataType("int")
	// PageSizeQueryParameter is a restful query parameter of the page size
	//
	// Deprecated
	PageSizeQueryParameter = restful.QueryParameter("pageSize", "The size of each paging data").DataType("int")

	// NameQueryParameter is a restful query parameter of the name filter
	NameQueryParameter = restful.QueryParameter(query.ParameterName, "Filter by name, containing match pattern")
	// PageQueryParameter is a restful query parameter of the page number filter
	PageQueryParameter = restful.QueryParameter(query.ParameterPage, "Which page you want to query. Default value is 1")
	// LimitQueryParameter is a restful query parameter of the page list filter
	LimitQueryParameter = restful.QueryParameter(query.ParameterLimit, "Which size per page you want to query. Default value is 10")
	// SortByQueryParameter is a restful query parameter of the order by
	SortByQueryParameter = restful.QueryParameter(query.ParameterOrderBy, `Order by field. Default value is "creationTimestamp"`)
	// AscendingQueryParameter is a restful query parameter of the order
	AscendingQueryParameter = restful.QueryParameter(query.ParameterAscending, "Sort order. Default is false(descending)")
)

// GetPathParameter returns the path parameter value from a request
func GetPathParameter(req *restful.Request, param *restful.Parameter) string {
	return req.PathParameter(param.Data().Name)
}

// GetQueryParameter returns the query parameter value from a request
func GetQueryParameter(req *restful.Request, param *restful.Parameter) string {
	return req.QueryParameter(param.Data().Name)
}

// GetPageParameters returns the page number and page size from a request
// the default value of page number is 1, page size is 10 if the number string is invalid
func GetPageParameters(req *restful.Request) (pageNumber, pageSize int) {
	pageNumberStr := req.QueryParameter(PageNumberQueryParameter.Data().Name)
	pageSizeStr := req.QueryParameter(PageSizeQueryParameter.Data().Name)

	var err error
	if pageNumber, err = strconv.Atoi(pageNumberStr); err != nil {
		pageNumber = 1
	}
	if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
		pageSize = 10
	}
	return
}

// Response is a common response method
func Response(req *restful.Request, res *restful.Response, object interface{}, err error) {
	if err != nil {
		kapis.HandleError(req, res, err)
	} else {
		_ = res.WriteEntity(object)
	}
}
