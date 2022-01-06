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
