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

// Parts of the following structs originally come from https://github.com/argoproj/argo-cd/

package argocd

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

// ClusterConfig is the configuration attributes. This structure is subset of the go-client
// rest.Config with annotations added for marshalling.
type ClusterConfig struct {
	// Server requires Basic authentication
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	// Server requires Bearer authentication. This client will not attempt to use
	// refresh tokens for an OAuth2 flow.
	// TODO: demonstrate an OAuth2 compatible client.
	BearerToken string `json:"bearerToken,omitempty"`

	// TLSClientConfig contains settings to enable transport layer security
	TLSClientConfig `json:"tlsClientConfig"`
}

// TLSClientConfig contains settings to enable transport layer security
type TLSClientConfig struct {
	// Insecure specifies that the server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool `json:"insecure"`
	// ServerName is passed to the server for SNI and is used in the client to check server
	// certificates against. If ServerName is empty, the hostname used to contact the
	// server is used.
	ServerName string `json:"serverName,omitempty"`
	// CertData holds PEM-encoded bytes (typically read from a client certificate file).
	// CertData takes precedence over CertFile
	CertData []byte `json:"certData,omitempty"`
	// KeyData holds PEM-encoded bytes (typically read from a client certificate key file).
	// KeyData takes precedence over KeyFile
	KeyData []byte `json:"keyData,omitempty"`
	// CAData holds PEM-encoded bytes (typically read from a root certificates bundle).
	// CAData takes precedence over CAFile
	CAData []byte `json:"caData,omitempty"`
}

func parseKubeConfig(data []byte) (cfg *config) {
	cfg = &config{}
	err := yaml.Unmarshal(data, cfg)
	fmt.Println(err)
	return
}

type config struct {
	Clusters []clusterConfig     `yaml:"clusters,omitempty"`
	Users    []clusterConfigUser `yaml:"users,omitempty"`
}

type clusterConfig struct {
	Name    string                  `yaml:"name,omitempty"`
	Cluster clusterConfigConnection `yaml:"cluster,omitempty"`
}

type clusterConfigConnection struct {
	SkipTLS bool   `yaml:"insecure-skip-tls-verify,omitempty"`
	Server  string `yaml:"server,omitempty"`
	CA      string `yaml:"certificate-authority-data:,omitempty"`
}

type clusterConfigUser struct {
	Name string                `yaml:"name,omitempty"`
	Auth clusterConfigUserAuth `yaml:"user,omitempty"`
}

type clusterConfigUserAuth struct {
	Token      string `yaml:"token,omitempty"`
	ClientCert string `yaml:"client-certificate-data,omitempty"`
	ClientKey  string `yaml:"client-key-data,omitempty"`
}
