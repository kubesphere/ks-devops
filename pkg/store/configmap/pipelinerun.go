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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/store/store"
	"kubesphere.io/devops/pkg/utils/k8sutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigMapStore represents a key-value store base on ConfigMap
type ConfigMapStore struct {
	ctx       context.Context
	k8sClient client.Client

	cache *v1.ConfigMap
	owner metav1.OwnerReference
}

// NewConfigMapStore creates a PipelineRun data store
func NewConfigMapStore(ctx context.Context, key client.ObjectKey, k8sClient client.Client) (
	result store.ConfigMapStore, err error) {
	cm := &v1.ConfigMap{Data: map[string]string{}}
	cm.Namespace = key.Namespace
	cm.Name = key.Name
	if err = k8sClient.Get(ctx, key, cm); err != nil {
		err = client.IgnoreNotFound(err)
	}

	result = &ConfigMapStore{
		cache:     cm,
		k8sClient: k8sClient,
		ctx:       ctx,
	}
	return
}

// GetStages returns the stage data
func (s *ConfigMapStore) GetStages() string {
	return s.Get(store.DataKeyStage)
}

// SetStages stores the stage data
func (s *ConfigMapStore) SetStages(stages string) {
	s.Set(store.DataKeyStage, stages)
}

// GetStatus returns the status
func (s *ConfigMapStore) GetStatus() string {
	return s.Get(store.DataKeyStatus)
}

// SetStatus stores the status
func (s *ConfigMapStore) SetStatus(status string) {
	s.Set(store.DataKeyStatus, status)
}

// GetStepLog returns the step log
func (s *ConfigMapStore) GetStepLog(stage, step int) string {
	return s.Get(store.StepLogKey(stage, step))
}

// SetStepLog stores the step log
func (s *ConfigMapStore) SetStepLog(stage, step int, log string) {
	s.Set(store.StepLogKey(stage, step), log)
}

// GetAllLog returns the whole log
func (s *ConfigMapStore) GetAllLog() string {
	return s.Get(store.DataKeyAllLog)
}

// SetAllLog store the whole log
func (s *ConfigMapStore) SetAllLog(log string) {
	s.Set(store.DataKeyAllLog, log)
}

// Get returns the value by a key
func (s *ConfigMapStore) Get(key string) string {
	return s.cache.Data[key]
}

// Set puts a key and value
func (s *ConfigMapStore) Set(key, value string) {
	s.cache.Data[key] = value
}

// Save puts the data into ConfigMap via Kubernetes client
func (s *ConfigMapStore) Save() (err error) {
	// it might be a new ConfigMap or not
	if s.cache.GetResourceVersion() == "" {
		k8sutil.SetOwnerReference(s.cache, s.owner)
		err = s.k8sClient.Create(s.ctx, s.cache)
	} else {
		err = s.k8sClient.Update(s.ctx, s.cache)
	}
	return
}

// SetOwnerReference set the owner reference
func (s *ConfigMapStore) SetOwnerReference(owner metav1.OwnerReference) {
	s.owner = owner
}
