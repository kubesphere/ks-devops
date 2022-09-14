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

package fake

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFakeStore(t *testing.T) {
	store := NewFakeStore()
	assert.NotNil(t, store)

	assert.Empty(t, store.Get("fake"))
	store.Set("fake", "fake")
	assert.Equal(t, "fake", store.Get("fake"))

	assert.Empty(t, store.GetAllLog())
	store.SetAllLog("log")
	assert.Equal(t, "log", store.GetAllLog())

	assert.Empty(t, store.GetStages())
	store.SetStages("stages")
	assert.Equal(t, "stages", store.GetStages())

	assert.Empty(t, store.GetStatus())
	store.SetStatus("status")
	assert.Equal(t, "status", store.GetStatus())

	assert.Empty(t, store.GetStepLog(1, 1))
	store.SetStepLog(1, 1, "step")
	assert.Equal(t, "step", store.GetStepLog(1, 1))

	assert.Nil(t, store.Save())
	assert.NotNil(t, store.WithError(errors.New("fake")).Save())
}
