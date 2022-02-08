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

package secretutil

import (
	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

type credentialMask func(*v1.Secret) *v1.Secret

var defaultMasque = []byte("")
var credentialMaskHolder = map[v1.SecretType]credentialMask{}

func basicAuthCredentialMask(secret *v1.Secret) *v1.Secret {
	secret.Data[v1alpha3.BasicAuthPasswordKey] = defaultMasque
	return secret
}

func sshAuthCredentialMask(secret *v1.Secret) *v1.Secret {
	secret.Data[v1alpha3.SSHAuthPassphraseKey] = defaultMasque
	secret.Data[v1alpha3.SSHAuthPrivateKey] = defaultMasque
	return secret
}

func secretTextCredentialMask(secret *v1.Secret) *v1.Secret {
	secret.Data[v1alpha3.SecretTextSecretKey] = defaultMasque
	return secret
}

func kubeconfigCredentialMask(secret *v1.Secret) *v1.Secret {
	secret.Data[v1alpha3.KubeConfigSecretKey] = defaultMasque
	return secret
}

// MaskCredential masks sensetive data inside credential.
func MaskCredential(secret *v1.Secret) *v1.Secret {
	if secret == nil || secret.Data == nil {
		return secret
	}
	credentialMask := credentialMaskHolder[secret.Type]
	if credentialMask != nil {
		return credentialMask(secret)
	}
	return secret
}

func init() {
	// register credential masks
	credentialMaskHolder[v1alpha3.SecretTypeBasicAuth] = basicAuthCredentialMask
	credentialMaskHolder[v1alpha3.SecretTypeSSHAuth] = sshAuthCredentialMask
	credentialMaskHolder[v1alpha3.SecretTypeSecretText] = secretTextCredentialMask
	credentialMaskHolder[v1alpha3.SecretTypeKubeConfig] = kubeconfigCredentialMask
}
