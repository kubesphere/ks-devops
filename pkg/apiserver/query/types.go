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

package query

import (
	"strconv"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/devops/pkg/utils/sliceutil"
)

const (
	ParameterName          = "name"
	ParameterLabelSelector = "labelSelector"
	ParameterFieldSelector = "fieldSelector"
	ParameterStart         = "start"
	ParameterPage          = "page"
	ParameterLimit         = "limit"
	ParameterOrderBy       = "sortBy"
	ParameterAscending     = "ascending"

	DefaultLimit = 10
	DefaultPage  = 1
)

// Query represents api search terms
type Query struct {
	Pagination *Pagination

	// sort result in which field, default to FieldCreationTimeStamp
	SortBy Field

	// sort result in ascending or descending order, default to descending
	Ascending bool

	//
	Filters map[Field]Value

	LabelSelector string
}

type Pagination struct {
	// items per page
	Limit int

	// offset
	Offset int
}

var NoPagination = newPagination(-1, 0)

// make sure that pagination is valid
func newPagination(limit int, offset int) *Pagination {
	if limit < 0 {
		limit = DefaultLimit
	}
	if offset < 0 {
		offset = 0
	}
	return &Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

func (q *Query) Selector() labels.Selector {
	if selector, err := labels.Parse(q.LabelSelector); err != nil {
		return labels.Everything()
	} else {
		return selector
	}
}

func (p *Pagination) GetValidPagination(total int) (startIndex, endIndex int) {
	// out of range
	if p.Limit < 0 || p.Offset < 0 || p.Offset > total {
		return 0, 0
	}

	startIndex = p.Offset
	endIndex = startIndex + p.Limit

	if endIndex > total {
		endIndex = total
	}

	return startIndex, endIndex
}

func New() *Query {
	return &Query{
		Pagination: NoPagination,
		SortBy:     "",
		Ascending:  false,
		Filters:    map[Field]Value{},
	}
}

type Filter struct {
	Field Field
	Value Value
}

func ParseQueryParameter(request *restful.Request) *Query {
	query := New()
	var start int
	var page int
	var limit int
	var err error

	limit, err = strconv.Atoi(request.QueryParameter(ParameterLimit))
	// equivalent to undefined, use the default value
	if err != nil || limit < -1 {
		limit = DefaultLimit
	}
	if pageStr := request.QueryParameter(ParameterPage); pageStr != "" {
		if page, err = strconv.Atoi(pageStr);err != nil {
			page = DefaultPage
		}
	} else {
		if start, err = strconv.Atoi(request.QueryParameter(ParameterStart)); err != nil {
			start = 0
		}
	}

	if page >= 1 {
		query.Pagination = newPagination(limit, (page-1)*limit)
	} else if start >= 0 {
		query.Pagination = newPagination(limit, start)
	}

	query.SortBy = Field(defaultString(request.QueryParameter(ParameterOrderBy), FieldCreationTimeStamp))

	ascending, err := strconv.ParseBool(defaultString(request.QueryParameter(ParameterAscending), "false"))
	if err != nil {
		query.Ascending = false
	} else {
		query.Ascending = ascending
	}

	query.LabelSelector = request.QueryParameter(ParameterLabelSelector)

	for key, values := range request.Request.URL.Query() {
		if !sliceutil.HasString([]string{
			ParameterPage,
			ParameterLimit,
			ParameterOrderBy,
			ParameterAscending,
			ParameterLabelSelector,
			ParameterFieldSelector,
		}, key) {
			// support multiple query condition
			for _, value := range values {
				query.Filters[Field(key)] = Value(value)
			}
		}
	}

	return query
}

func defaultString(value, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
