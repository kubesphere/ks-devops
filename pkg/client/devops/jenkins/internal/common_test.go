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

package internal

import (
	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	devopsv1alpha3 "kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"testing"
)

func TestCommonSituation(t *testing.T) {
	// make sure these functions do not panic
	// I add these test cases because it's possible that users just do give the git source
	AppendGitlabSourceToEtree(nil, nil)
	AppendGithubSourceToEtree(nil, nil)
	AppendBitbucketServerSourceToEtree(nil, nil)
	AppendGitSourceToEtree(nil, nil)
	AppendSingleSvnSourceToEtree(nil, nil)
	AppendSvnSourceToEtree(nil, nil)
}

func TestSkipJenkinsNotification(t *testing.T) {
	PRForks := &devopsv1alpha3.DiscoverPRFromForks{
		Strategy: 1,
		Trust:    1,
	}

	// github
	source := etree.NewDocument().CreateElement("source")
	AppendGithubSourceToEtree(source, &devopsv1alpha3.GithubSource{DiscoverPRFromForks: PRForks})
	githubSource := GetGithubSourcefromEtree(source)
	assert.False(t, githubSource.AcceptJenkinsNotification)

	// gitlab
	source = etree.NewDocument().CreateElement("source")
	AppendGitlabSourceToEtree(source, &devopsv1alpha3.GitlabSource{DiscoverPRFromForks: PRForks})
	gitlabSource := GetGitlabSourceFromEtree(source)
	assert.False(t, gitlabSource.AcceptJenkinsNotification)

	// bitbucketServer
	source = etree.NewDocument().CreateElement("source")
	AppendBitbucketServerSourceToEtree(source, &devopsv1alpha3.BitbucketServerSource{DiscoverPRFromForks: PRForks})
	bitbucketServerSource := GetBitbucketServerSourceFromEtree(source)
	assert.False(t, bitbucketServerSource.AcceptJenkinsNotification)
}

func TestAcceptJenkinsNotification(t *testing.T) {
	PRForks := &devopsv1alpha3.DiscoverPRFromForks{
		Strategy: 1,
		Trust:    1,
	}

	// github
	source := etree.NewDocument().CreateElement("source")
	AppendGithubSourceToEtree(source, &devopsv1alpha3.GithubSource{DiscoverPRFromForks: PRForks, AcceptJenkinsNotification: true})
	githubSource := GetGithubSourcefromEtree(source)
	assert.True(t, githubSource.AcceptJenkinsNotification)

	// gitlab
	source = etree.NewDocument().CreateElement("source")
	AppendGitlabSourceToEtree(source, &devopsv1alpha3.GitlabSource{DiscoverPRFromForks: PRForks, AcceptJenkinsNotification: true})
	gitlabSource := GetGitlabSourceFromEtree(source)
	assert.True(t, gitlabSource.AcceptJenkinsNotification)

	// bitbucketServer
	source = etree.NewDocument().CreateElement("source")
	AppendBitbucketServerSourceToEtree(source, &devopsv1alpha3.BitbucketServerSource{DiscoverPRFromForks: PRForks, AcceptJenkinsNotification: true})
	bitbucketServerSource := GetBitbucketServerSourceFromEtree(source)
	assert.True(t, bitbucketServerSource.AcceptJenkinsNotification)
}
