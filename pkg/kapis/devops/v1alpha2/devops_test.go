package v1alpha2

import (
	"github.com/stretchr/testify/assert"
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

	for i, item := range table {
		t.Run(item.message, func(t *testing.T) {
			pipelineName, ns := parseNameFilterFromQuery(item.query)
			assert.Equal(t, item.expectedPipelineName, pipelineName)
			if ns != item.expectNamespace {
				t.Fatalf("invalid namespace, index: %d, message: %s", i, item.message)
			}
		})
	}
}
