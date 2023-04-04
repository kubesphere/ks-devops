// Created by zhbinary on 2023/3/31.
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var (
	GroupVersion = schema.GroupVersion{Group: "installer.kubesphere.io", Version: "v1alpha1"}
)

type InstallerV1alpha1Interface interface {
	RESTClient() rest.Interface
}

type InstallerV1alpha1Client struct {
	client rest.Interface
}

func (i *InstallerV1alpha1Client) RESTClient() rest.Interface {
	return i.client
}

func NewForConfig4Installer(c *rest.Config) (*InstallerV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &InstallerV1alpha1Client{client}, nil
}

func setConfigDefaults(config *rest.Config) error {
	gv := GroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}
