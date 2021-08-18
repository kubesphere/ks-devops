package runtime

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestNewWebServiceWithoutGroup(t *testing.T) {
	ws := NewWebServiceWithoutGroup(schema.GroupVersion{
		Version: "v2",
	})

	assert.NotNil(t, ws)
	assert.Equal(t, "/v2", ws.RootPath())
}

func TestNewWebService(t *testing.T) {
	ws := NewWebService(schema.GroupVersion{
		Group:   "devops",
		Version: "v2",
	})

	assert.NotNil(t, ws)
	assert.Equal(t, ApiRootPath+"/devops/v2", ws.RootPath())
}
