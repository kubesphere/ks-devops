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

package jclient

import (
	"github.com/jenkins-zh/jenkins-client/pkg/core"

	"github.com/kubesphere/ks-devops/pkg/client/devops"
	"github.com/kubesphere/ks-devops/pkg/client/devops/jenkins"
)

// JenkinsClient represents a client of Jenkins
type JenkinsClient struct {
	Core             core.JenkinsCore
	SaveKubeConfigAs string
	jenkins          *jenkins.Jenkins // For refactor purpose only
}

var _ devops.Interface = &JenkinsClient{}

// NewJenkinsClient creates a Jenkins client
func NewJenkinsClient(options *jenkins.Options) (*JenkinsClient, error) {
	jenkinsCore := core.JenkinsCore{
		URL:      options.Host,
		UserName: options.Username,
		Token:    options.ApiToken,
	}

	devopsClient, _ := jenkins.NewDevopsClient(options) // For refactor purpose only
	return &JenkinsClient{
		Core:             jenkinsCore,
		jenkins:          devopsClient, // For refactor purpose only
		SaveKubeConfigAs: options.SaveKubeConfigAs,
	}, nil
}
