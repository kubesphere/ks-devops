/*
Copyright 2020 KubeSphere Authors

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

package k8s

import (
	"strings"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubesphere "kubesphere.io/devops/pkg/client/clientset/versioned"
)

type Client interface {
	Kubernetes() kubernetes.Interface
	KubeSphere() kubesphere.Interface
	ApiExtensions() apiextensionsclient.Interface
	Discovery() discovery.DiscoveryInterface
	Master() string
	Config() *rest.Config
}

type kubernetesClient struct {
	// kubernetes client interface
	k8s kubernetes.Interface
	// discovery client
	discoveryClient *discovery.DiscoveryClient

	// generated clientset
	ks            kubesphere.Interface
	apiextensions apiextensionsclient.Interface
	master        string
	config        *rest.Config
}

// NewKubernetesClientOrDie creates KubernetesClient and panic if there is an error
func NewKubernetesClientOrDie(options *KubernetesOptions) (client Client) {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		panic(err)
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	k := &kubernetesClient{
		k8s:             kubernetes.NewForConfigOrDie(config),
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(config),
		apiextensions:   apiextensionsclient.NewForConfigOrDie(config),
		master:          config.Host,
		config:          config,
	}

	if options.Master != "" {
		k.master = options.Master
	}
	// The https prefix is automatically added when using sa.
	// But it will not be set automatically when reading from kubeconfig
	// which may cause some problems in the client of other languages.
	if !strings.HasPrefix(k.master, "http://") && !strings.HasPrefix(k.master, "https://") {
		k.master = "https://" + k.master
	}
	return k
}

// NewKubernetesClientWithConfig creates a k8s client with the rest config
func NewKubernetesClientWithConfig(config *rest.Config) (client Client, err error) {
	if config == nil {
		return
	}

	var k kubernetesClient
	if k.k8s, err = kubernetes.NewForConfig(config); err != nil {
		return
	}

	if k.discoveryClient, err = discovery.NewDiscoveryClientForConfig(config); err != nil {
		return
	}

	if k.ks, err = kubesphere.NewForConfig(config); err != nil {
		return
	}

	if k.apiextensions, err = apiextensionsclient.NewForConfig(config); err != nil {
		return
	}

	k.config = config
	client = &k
	return
}

// NewKubernetesClientWithToken creates a k8s client with a bearer token
func NewKubernetesClientWithToken(token string, master string) (client Client, err error) {
	if token == "" {
		return
	}

	client, err = NewKubernetesClientWithConfig(&rest.Config{
		BearerToken: token,
		Host:        master,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	})
	return
}

// NewKubernetesClient creates a KubernetesClient
func NewKubernetesClient(options *KubernetesOptions) (client Client, err error) {
	if options == nil {
		return
	}

	var config *rest.Config
	if config, err = clientcmd.BuildConfigFromFlags("", options.KubeConfig); err != nil {
		return
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	if client, err = NewKubernetesClientWithConfig(config); err == nil {
		if k8sClient, ok := client.(*kubernetesClient); ok {
			k8sClient.config = config
			k8sClient.master = options.Master
		}
	}
	return
}

func (k *kubernetesClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *kubernetesClient) Discovery() discovery.DiscoveryInterface {
	return k.discoveryClient
}

func (k *kubernetesClient) ApiExtensions() apiextensionsclient.Interface {
	return k.apiextensions
}

// master address used to generate kubeconfig for downloading
func (k *kubernetesClient) Master() string {
	return k.master
}

func (k *kubernetesClient) Config() *rest.Config {
	return k.config
}

func (k *kubernetesClient) KubeSphere() kubesphere.Interface {
	return k.ks
}
