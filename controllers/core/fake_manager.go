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

package core

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// FakeManager is for the test purpose
type FakeManager struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// Add is a fake method
func (f *FakeManager) Add(manager.Runnable) error {
	return nil
}

// Elected is a fake method
func (f *FakeManager) Elected() <-chan struct{} {
	return nil
}

// SetFields is a fake method
func (f *FakeManager) SetFields(interface{}) error {
	return nil
}

// AddMetricsExtraHandler is a fake method
func (f *FakeManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	return nil
}

// AddHealthzCheck is a fake method
func (f *FakeManager) AddHealthzCheck(string, healthz.Checker) error {
	return nil
}

// AddReadyzCheck is a fake method
func (f *FakeManager) AddReadyzCheck(string, healthz.Checker) error {
	return nil
}

// Start is a fake method
func (f *FakeManager) Start(<-chan struct{}) error {
	return nil
}

// GetConfig is a fake method
func (f *FakeManager) GetConfig() *rest.Config {
	return nil
}

// GetScheme is a fake method
func (f *FakeManager) GetScheme() *runtime.Scheme {
	return f.Scheme
}

// GetClient is a fake method
func (f *FakeManager) GetClient() client.Client {
	return f.Client
}

// GetFieldIndexer is a fake method
func (f *FakeManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

// GetCache is a fake method
func (f *FakeManager) GetCache() cache.Cache {
	return nil
}

// GetEventRecorderFor is a fake method
func (f *FakeManager) GetEventRecorderFor(name string) record.EventRecorder {
	return nil
}

// GetRESTMapper is a fake method
func (f *FakeManager) GetRESTMapper() meta.RESTMapper {
	return meta.FirstHitRESTMapper{}
}

// GetAPIReader is a fake method
func (f *FakeManager) GetAPIReader() client.Reader {
	return f.Client
}

// GetWebhookServer is a fake method
func (f *FakeManager) GetWebhookServer() *webhook.Server {
	return nil
}

// GetLogger is a fake method
func (f *FakeManager) GetLogger() logr.Logger {
	return log.NullLogger{}
}
