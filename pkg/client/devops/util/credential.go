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

package util

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	jcredential "github.com/jenkins-zh/jenkins-client/pkg/credential"
	devopsv1alpha3 "github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	v1 "k8s.io/api/core/v1"
)

// ConvertSecretToCredential converts a secret to Jenkins credential type
func ConvertSecretToCredential(secret *v1.Secret, saveKubeConfigAs string) (interface{}, error) {
	name := secret.GetName()

	switch secret.Type {
	case devopsv1alpha3.SecretTypeBasicAuth:
		username := string(secret.Data[devopsv1alpha3.BasicAuthUsernameKey])
		password := string(secret.Data[devopsv1alpha3.BasicAuthPasswordKey])
		return jcredential.NewUsernamePasswordCredential(name, username, password), nil
	case devopsv1alpha3.SecretTypeSSHAuth:
		username := string(secret.Data[devopsv1alpha3.SSHAuthUsernameKey])
		passphrase := string(secret.Data[devopsv1alpha3.SSHAuthPassphraseKey])
		privatekey := string(secret.Data[devopsv1alpha3.SSHAuthPrivateKey])
		return jcredential.NewSSHCredential(name, username, passphrase, privatekey), nil
	case devopsv1alpha3.SecretTypeSecretText:
		secretContent := string(secret.Data[devopsv1alpha3.SecretTextSecretKey])
		return jcredential.NewSecretTextCredential(name, secretContent), nil
	case devopsv1alpha3.SecretTypeKubeConfig:
		secretContent := string(secret.Data[devopsv1alpha3.KubeConfigSecretKey])
		// for backward compatibility, empty value means kubeconfig
		if saveKubeConfigAs == devopsv1alpha3.SecretTextString {
			return jcredential.NewSecretTextCredential(name, secretContent), nil
		}
		return jcredential.NewKubeConfigCredential(name, secretContent), nil
	default:
		err := fmt.Errorf("error unsupport credential type")
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}
}
