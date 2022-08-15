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

package sonarqube

import (
	"fmt"
	"strings"

	sonargo "github.com/kubesphere/sonargo/sonar"
	"k8s.io/klog"
)

// Client represents the client of SonarQube
type Client struct {
	client *sonargo.Client
}

// NewSonarQubeClient creates a client with options
func NewSonarQubeClient(options *Options) (*Client, error) {
	var endpoint string

	if strings.HasSuffix(options.Host, "/") {
		endpoint = fmt.Sprintf("%sapi/", options.Host)
	} else {
		endpoint = fmt.Sprintf("%s/api/", options.Host)
	}

	sonar, err := sonargo.NewClientWithToken(endpoint, options.Token)
	if err != nil {
		klog.Errorf("failed to connect to sonarqube service, %+v", err)
		return nil, err
	}

	return &Client{client: sonar}, err
}

// SonarQube returns sonarqube client
// Also we can wrap some methods to avoid direct use sonar client
func (s *Client) SonarQube() *sonargo.Client {
	return s.client
}
