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

package configmap

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/store/store"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestConfigMapStore(t *testing.T) {
	var err error
	var cmStore store.ConfigMapStore
	cmStore, err = NewConfigMapStore(context.Background(), types.NamespacedName{
		Namespace: "ns",
		Name:      "name",
	}, fake.NewClientBuilder().Build())
	assert.NotNil(t, cmStore)
	assert.Nil(t, err)

	cmStore.SetOwnerReference(v1.OwnerReference{})
	assert.Nil(t, cmStore.Save())

	assert.Empty(t, cmStore.GetStages())
	cmStore.SetStages("stages")
	assert.Equal(t, "stages", cmStore.GetStages())

	assert.Empty(t, cmStore.GetStatus())
	cmStore.SetStatus("status")
	assert.Equal(t, "status", cmStore.GetStatus())

	assert.Empty(t, cmStore.GetStepLog(1, 2))
	cmStore.SetStepLog(1, 2, "step")
	assert.Equal(t, "step", cmStore.GetStepLog(1, 2))

	assert.Empty(t, cmStore.GetAllLog())
	cmStore.SetAllLog("log")
	assert.Equal(t, "log", cmStore.GetAllLog())

	assert.Nil(t, cmStore.Save())
}
