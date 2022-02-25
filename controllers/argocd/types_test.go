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

package argocd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_parseKubeConfig(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want *config
	}{{
		name: "with token",
		args: args{
			data: []byte(`
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: server
  name: name
users:
- name: name
  user:
    token: token`),
		},
		want: &config{
			Clusters: []clusterConfig{{
				Name: "name",
				Cluster: clusterConfigConnection{
					SkipTLS: true,
					Server:  "server",
				},
			}},
			Users: []clusterConfigUser{{
				Name: "name",
				Auth: clusterConfigUserAuth{
					Token: "token",
				},
			}},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parseKubeConfig(tt.args.data), "parseKubeConfig(%v)", tt.args.data)
		})
	}
}
