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
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

func Test_maskCredential(t *testing.T) {
	type args struct {
		secret *v1.Secret
	}
	tests := []struct {
		name string
		args args
		want *v1.Secret
	}{{
		name: "Mask basic auth secret",
		args: args{
			secret: &v1.Secret{
				Type: v1alpha3.SecretTypeBasicAuth,
				Data: map[string][]byte{
					v1alpha3.BasicAuthPasswordKey: []byte("fake password"),
					v1alpha3.BasicAuthUsernameKey: []byte("fake username"),
				},
			},
		},
		want: &v1.Secret{
			Type: v1alpha3.SecretTypeBasicAuth,
			Data: map[string][]byte{
				v1alpha3.BasicAuthPasswordKey: []byte(""),
				v1alpha3.BasicAuthUsernameKey: []byte("fake username"),
			},
		},
	}, {
		name: "Mask ssh auth secret",
		args: args{
			secret: &v1.Secret{
				Type: v1alpha3.SecretTypeSSHAuth,
				Data: map[string][]byte{
					v1alpha3.SSHAuthPassphraseKey: []byte("fake password"),
					v1alpha3.SSHAuthPrivateKey:    []byte("fake private key"),
					v1alpha3.SSHAuthUsernameKey:   []byte("fake username"),
				},
			},
		},
		want: &v1.Secret{
			Type: v1alpha3.SecretTypeSSHAuth,
			Data: map[string][]byte{
				v1alpha3.SSHAuthPassphraseKey: []byte(""),
				v1alpha3.SSHAuthPrivateKey:    []byte(""),
				v1alpha3.SSHAuthUsernameKey:   []byte("fake username"),
			},
		},
	}, {
		name: "Mask secret text secret",
		args: args{
			secret: &v1.Secret{
				Type: v1alpha3.SecretTypeSecretText,
				Data: map[string][]byte{
					v1alpha3.SecretTextSecretKey: []byte("fake secret text"),
				},
			},
		},
		want: &v1.Secret{
			Type: v1alpha3.SecretTypeSecretText,
			Data: map[string][]byte{
				v1alpha3.SecretTextSecretKey: []byte(""),
			},
		},
	}, {
		name: "Mask kubeconfig secret",
		args: args{
			secret: &v1.Secret{
				Type: v1alpha3.SecretTypeKubeConfig,
				Data: map[string][]byte{
					v1alpha3.KubeConfigSecretKey: []byte("fake kubeconfig"),
				},
			},
		},
		want: &v1.Secret{
			Type: v1alpha3.SecretTypeKubeConfig,
			Data: map[string][]byte{
				v1alpha3.KubeConfigSecretKey: []byte(""),
			},
		},
	}, {
		name: "Nil secret",
		args: args{
			secret: nil,
		},
		want: nil,
	}, {
		name: "Other secret",
		args: args{
			secret: &v1.Secret{
				Type: v1.SecretType("fake-type"),
				Data: map[string][]byte{
					"fake_key": []byte("fake_value"),
				},
			},
		},
		want: &v1.Secret{
			Type: v1.SecretType("fake-type"),
			Data: map[string][]byte{
				"fake_key": []byte("fake_value"),
			},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskCredential(tt.args.secret); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("maskCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}
