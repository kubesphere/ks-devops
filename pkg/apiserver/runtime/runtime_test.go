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
