package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"

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
