/*
Copyright 2022 The KubeSphere Authors.

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

package v1alpha2

import (
	"net/http"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

func TestParseNameFilterFromQuery(t *testing.T) {
	table := []struct {
		query                string
		pipeline             *v1alpha3.Pipeline
		expectNamespace      string
		expectedPipelineName string
		message              string
	}{{
		query:                "type:pipeline;organization:jenkins;pipeline:serverjkq4c/*",
		pipeline:             &v1alpha3.Pipeline{},
		expectNamespace:      "serverjkq4c",
		expectedPipelineName: "",
		message:              "query all pipelines with filter *",
	}, {
		query:                "type:pipeline;organization:jenkins;pipeline:cccc/*abc*",
		pipeline:             &v1alpha3.Pipeline{},
		expectNamespace:      "cccc",
		expectedPipelineName: "abc",
		message:              "query all pipelines with filter abc",
	}, {
		query:                "type:pipeline;organization:jenkins;pipeline:pai-serverjkq4c/*",
		pipeline:             &v1alpha3.Pipeline{},
		expectNamespace:      "pai-serverjkq4c",
		expectedPipelineName: "",
		message:              "query all pipelines with filter *",
	}, {
		query:                "type:pipeline;organization:jenkins;pipeline:defdef",
		pipeline:             &v1alpha3.Pipeline{},
		expectNamespace:      "defdef",
		expectedPipelineName: "",
		message:              "query all pipelines with filter *",
	}}

	for _, item := range table {
		t.Run(item.message, func(t *testing.T) {
			pipelineName, ns := parseNameFilterFromQuery(item.query)
			assert.Equal(t, item.expectedPipelineName, pipelineName)
			assert.Equal(t, item.expectNamespace, ns)
		})
	}
}

func TestBuildPipelineSearchQueryParam(t *testing.T) {
	httpReq, _ := http.NewRequest(http.MethodGet, "http://localhost?start=10&limit=10", nil)
	req := &restful.Request{
		Request: httpReq,
	}
	nameReg := ""
	query := buildPipelineSearchQueryParam(req, nameReg)
	assert.NotNil(t, query)
	assert.Equal(t, 10, query.Pagination.Offset)
	assert.Equal(t, 10, query.Pagination.Limit)

	// use parameter: 'page'
	httpReq, _ = http.NewRequest(http.MethodGet, "http://localhost?page=2&limit=20", nil)
	req = &restful.Request{
		Request: httpReq,
	}
	query = buildPipelineSearchQueryParam(req, nameReg)
	assert.NotNil(t, query)
	assert.Equal(t, 20, query.Pagination.Offset)
	assert.Equal(t, 20, query.Pagination.Limit)

	// mixed with parameter 'start' and 'page', take 'page` as high priority
	httpReq, _ = http.NewRequest(http.MethodGet, "http://localhost?page=2&limit=20&start=100", nil)
	req = &restful.Request{
		Request: httpReq,
	}
	query = buildPipelineSearchQueryParam(req, nameReg)
	assert.NotNil(t, query)
	assert.Equal(t, 20, query.Pagination.Offset)
	assert.Equal(t, 20, query.Pagination.Limit)
}
