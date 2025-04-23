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

package git

import (
	"context"
	"fmt"

	goscm "github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClientFactory responsible for creating a git client
type ClientFactory struct {
	provider  string
	secretRef *v1.SecretReference
	k8sClient client.Client

	Server string
}

// NewClientFactory creates an instance of the ClientFactory
func NewClientFactory(provider string, secretRef *v1.SecretReference, k8sClient client.Client) *ClientFactory {
	return &ClientFactory{
		provider:  provider,
		secretRef: secretRef,
		k8sClient: k8sClient,
	}
}

// GetClient returns the git client with auth
func (c *ClientFactory) GetClient() (client *goscm.Client, err error) {
	provider := c.provider
	switch c.provider {
	case "bitbucket_cloud":
		provider = "bitbucketcloud"
	case "bitbucket-server":
		provider = "bitbucketserver"
	}

	if c.Server == "https://api.bitbucket.org" || c.Server == "https://bitbucket.org" {
		provider = "bitbucketcloud"
	}

	var token string
	username := ""
	if c.secretRef != nil {
		if token, username, _, err = c.GetTokenFromSecret(c.secretRef); err != nil {
			return
		}
	}
	client, err = factory.NewClient(provider, c.Server, token, func(scmClient *goscm.Client) {
		scmClient.Username = username
	})
	return
}

func (c *ClientFactory) GetTokenFromSecret(secretRef *v1.SecretReference) (token, username string, privateKey []byte, err error) {
	var gitSecret *v1.Secret
	if gitSecret, err = c.getSecret(secretRef); err != nil {
		return
	}

	switch gitSecret.Type {
	case v1.SecretTypeBasicAuth, v1alpha3.SecretTypeBasicAuth:
		token = string(gitSecret.Data[v1.BasicAuthPasswordKey])
		username = string(gitSecret.Data[v1.BasicAuthUsernameKey])
	case v1.SecretTypeOpaque:
		token = string(gitSecret.Data[v1.ServiceAccountTokenKey])
	case v1alpha3.SecretTypeSecretText:
		username = string(gitSecret.Data[v1.BasicAuthUsernameKey])
		token = string(gitSecret.Data[v1alpha3.SecretTextSecretKey])
	case v1alpha3.SecretTypeSSHAuth:
		privateKey = gitSecret.Data[v1alpha3.SSHAuthPrivateKey]
		token = string(gitSecret.Data[v1alpha3.SSHAuthPassphraseKey])
		username = string(gitSecret.Data[v1alpha3.SSHAuthUsernameKey])
	}
	return
}

// getSecret returns the secret, taking the namespace from GitRepository if it is empty
func (c *ClientFactory) getSecret(ref *v1.SecretReference) (secret *v1.Secret, err error) {
	secret = &v1.Secret{}
	ns := ref.Namespace

	if err = c.k8sClient.Get(context.TODO(), types.NamespacedName{
		Namespace: ns, Name: ref.Name,
	}, secret); err != nil {
		err = fmt.Errorf("cannot get secret %v, error is: %v", secret, err)
	}
	return
}
