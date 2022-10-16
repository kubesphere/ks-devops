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

import "kubesphere.io/devops/pkg/store/store"

// FakeStore is a fake store
type FakeStore struct {
	data map[string]string
	err  error
}

// NewFakeStore creates an instance of the fake store
func NewFakeStore() *FakeStore {
	return &FakeStore{
		data: map[string]string{},
	}
}

// WithError sets the error for expect error
func (s *FakeStore) WithError(err error) *FakeStore {
	s.err = err
	return s
}

// Get is a fake method
func (s *FakeStore) Get(key string) string {
	return s.data[key]
}

// Set is a fake method
func (s *FakeStore) Set(key, value string) {
	s.data[key] = value
}

// Save is a fake method
func (s *FakeStore) Save() error {
	return s.err
}

// GetStages is a fake method
func (s *FakeStore) GetStages() string {
	return s.Get(store.DataKeyStage)
}

// SetStages is fake method
func (s *FakeStore) SetStages(stages string) {
	s.Set(store.DataKeyStage, stages)
}

// GetStatus is a fake method
func (s *FakeStore) GetStatus() string {
	return s.Get(store.DataKeyStatus)
}

// SetStatus is a fake method
func (s *FakeStore) SetStatus(status string) {
	s.Set(store.DataKeyStatus, status)
}

// GetStepLog is a fake method
func (s *FakeStore) GetStepLog(stage, step int) string {
	return s.Get(store.StepLogKey(stage, step))
}

// SetStepLog is a fake method
func (s *FakeStore) SetStepLog(stage, step int, log string) {
	s.Set(store.StepLogKey(stage, step), log)
}

// GetAllLog is a fake method
func (s *FakeStore) GetAllLog() string {
	return s.Get(store.DataKeyAllLog)
}

// SetAllLog is a fake method
func (s *FakeStore) SetAllLog(log string) {
	s.data[store.DataKeyAllLog] = log
}
