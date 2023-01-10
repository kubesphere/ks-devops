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

// Store is a fake store
type Store struct {
	data map[string]string
	err  error
}

// NewFakeStore creates an instance of the fake store
func NewFakeStore() *Store {
	return &Store{
		data: map[string]string{},
	}
}

// WithError sets the error for expect error
func (s *Store) WithError(err error) *Store {
	s.err = err
	return s
}

// Get is a fake method
func (s *Store) Get(key string) string {
	return s.data[key]
}

// Set is a fake method
func (s *Store) Set(key, value string) {
	s.data[key] = value
}

// Save is a fake method
func (s *Store) Save() error {
	return s.err
}

// GetStages is a fake method
func (s *Store) GetStages() string {
	return s.Get(store.DataKeyStage)
}

// SetStages is fake method
func (s *Store) SetStages(stages string) {
	s.Set(store.DataKeyStage, stages)
}

// GetStatus is a fake method
func (s *Store) GetStatus() string {
	return s.Get(store.DataKeyStatus)
}

// SetStatus is a fake method
func (s *Store) SetStatus(status string) {
	s.Set(store.DataKeyStatus, status)
}

// GetStepLog is a fake method
func (s *Store) GetStepLog(stage, step int) string {
	return s.Get(store.StepLogKey(stage, step))
}

// SetStepLog is a fake method
func (s *Store) SetStepLog(stage, step int, log string) {
	s.Set(store.StepLogKey(stage, step), log)
}

// GetAllLog is a fake method
func (s *Store) GetAllLog() string {
	return s.Get(store.DataKeyAllLog)
}

// SetAllLog is a fake method
func (s *Store) SetAllLog(log string) {
	s.data[store.DataKeyAllLog] = log
}
