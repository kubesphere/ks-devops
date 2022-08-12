/*
Copyright 2022 KubeSphere Authors

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

package v1alpha3

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_defaultListHandler_Comparator(t *testing.T) {
	d := defaultListHandler{}
	assert.NotNil(t, d.Comparator())
	assert.NotNil(t, d.Filter())
	assert.NotNil(t, d.Transformer())
}
