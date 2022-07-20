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

type FakeManager struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (f *FakeManager) Add(manager.Runnable) error {
	return nil
}

func (f *FakeManager) Elected() <-chan struct{} {
	return nil
}

func (f *FakeManager) SetFields(interface{}) error {
	return nil
}

func (f *FakeManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	return nil
}

func (f *FakeManager) AddHealthzCheck(string, healthz.Checker) error {
	return nil
}

func (f *FakeManager) AddReadyzCheck(string, healthz.Checker) error {
	return nil
}

func (f *FakeManager) Start(<-chan struct{}) error {
	return nil
}

func (f *FakeManager) GetConfig() *rest.Config {
	return nil
}

func (f *FakeManager) GetScheme() *runtime.Scheme {
	return f.Scheme
}

func (f *FakeManager) GetClient() client.Client {
	return f.Client
}

func (f *FakeManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (f *FakeManager) GetCache() cache.Cache {
	return nil
}

func (f *FakeManager) GetEventRecorderFor(name string) record.EventRecorder {
	return nil
}

func (f *FakeManager) GetRESTMapper() meta.RESTMapper {
	return meta.FirstHitRESTMapper{}
}

func (f *FakeManager) GetAPIReader() client.Reader {
	return f.Client
}

func (f *FakeManager) GetWebhookServer() *webhook.Server {
	return nil
}

func (f *FakeManager) GetLogger() logr.Logger {
	return log.NullLogger{}
}
