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
