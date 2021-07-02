package app

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type k8sClientFactory interface {
	Get() (kubernetes.Interface, error)
}

// make sure DefaultK8sClientFactory implemented the interface k8sClientFactory
var _ k8sClientFactory = &DefaultK8sClientFactory{}

// DefaultK8sClientFactory is the default k8s client factory
type DefaultK8sClientFactory struct {
}

func (d *DefaultK8sClientFactory) Get() (client kubernetes.Interface, err error) {
	var config *rest.Config
	if config, err = clientcmd.BuildConfigFromFlags("", ""); err == nil {
		client, err = kubernetes.NewForConfig(config)
	}
	return
}
