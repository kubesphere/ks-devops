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

package jenkins

import (
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/client/devops/jenkins/internal"
	"reflect"
	"testing"

	devopsv1alpha3 "kubesphere.io/devops/pkg/api/devops/v1alpha3"
)

func Test_NoScmPipelineConfig(t *testing.T) {
	inputs := []*devopsv1alpha3.NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
		},
		{
			Name:        "",
			Description: "",
			Jenkinsfile: "node{echo 'hello'}",
		},
		{
			Name:              "",
			Description:       "",
			Jenkinsfile:       "node{echo 'hello'}",
			DisableConcurrent: true,
		},
	}
	for _, input := range inputs {
		outputString, err := createPipelineConfigXml(input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parsePipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_NoScmPipelineConfig_Discarder(t *testing.T) {
	inputs := []*devopsv1alpha3.NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "3", NumToKeep: "5",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "3", NumToKeep: "",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "", NumToKeep: "21321",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "", NumToKeep: "",
			},
		},
	}
	for _, input := range inputs {
		outputString, err := createPipelineConfigXml(input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parsePipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_NoScmPipelineConfig_Param(t *testing.T) {
	inputs := []*devopsv1alpha3.NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Parameters: []devopsv1alpha3.ParameterDefinition{
				{
					Name:         "d",
					DefaultValue: "a\nb",
					Type:         "choice",
					Description:  "fortest",
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Parameters: []devopsv1alpha3.ParameterDefinition{
				{
					Name:         "a",
					DefaultValue: "abc",
					Type:         "string",
					Description:  "fortest",
				},
				{
					Name:         "b",
					DefaultValue: "false",
					Type:         "boolean",
					Description:  "fortest",
				},
				{
					Name:         "c",
					DefaultValue: "password \n aaa",
					Type:         "text",
					Description:  "fortest",
				},
				{
					Name:         "d",
					DefaultValue: "a\nb",
					Type:         "choice",
					Description:  "fortest",
				},
			},
		},
	}
	for _, input := range inputs {
		outputString, err := createPipelineConfigXml(input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parsePipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_NoScmPipelineConfig_Trigger(t *testing.T) {
	inputs := []*devopsv1alpha3.NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Cron: "1 1 1 * * *",
			},
		},

		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			RemoteTrigger: &devopsv1alpha3.RemoteTrigger{
				Token: "abc",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Cron: "1 1 1 * * *",
			},
			RemoteTrigger: &devopsv1alpha3.RemoteTrigger{
				Token: "abc",
			},
		},
	}

	for index, input := range inputs {
		outputString, err := createPipelineConfigXml(input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parsePipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("index: %d, input [%+v] output [%+v] should equal ", index, input, output)
		}
	}
}

func Test_MultiBranchPipelineConfig(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource:   &devopsv1alpha3.GitSource{},
		},
		{
			Name:         "",
			Description:  "for test",
			ScriptPath:   "Jenkinsfile",
			SourceType:   "github",
			GitHubSource: &devopsv1alpha3.GithubSource{},
		},
		{
			Name:            "",
			Description:     "for test",
			ScriptPath:      "Jenkinsfile",
			SourceType:      "single_svn",
			SingleSvnSource: &devopsv1alpha3.SingleSvnSource{},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "svn",
			SvnSource:   &devopsv1alpha3.SvnSource{},
		},
		{
			Name:         "",
			Description:  "for test",
			ScriptPath:   "Jenkinsfile",
			SourceType:   "gitlab",
			GitlabSource: &devopsv1alpha3.GitlabSource{},
		},
	}
	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_MultiBranchPipelineConfig_Discarder(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "1",
				NumToKeep:  "2",
			},
			GitSource: &devopsv1alpha3.GitSource{},
		},
	}
	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_MultiBranchPipelineConfig_TimerTrigger(t *testing.T) {
	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			GitSource: &devopsv1alpha3.GitSource{},
		},
	}
	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_MultiBranchPipelineConfig_Source(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			GitSource: &devopsv1alpha3.GitSource{
				Url:              "https://github.com/kubesphere/devops",
				CredentialId:     "git",
				DiscoverBranches: true,
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "gitlab",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			GitlabSource: &devopsv1alpha3.GitlabSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "gitlab",
				ServerName:           "default-gitlab",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverTags:         true,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				CloneOption: &devopsv1alpha3.GitCloneOption{
					Timeout: 10,
					Depth:   10,
				},
				RegexFilter: "*-dev",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "gitlab",
			GitlabSource: &devopsv1alpha3.GitlabSource{
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    2,
				},
				//CloneOption: &devopsv1alpha3.GitCloneOption{
				//	Depth:   -1,
				//	Timeout: -1,
				//},
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "gitlab",
			GitlabSource: &devopsv1alpha3.GitlabSource{
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    3,
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "gitlab",
			GitlabSource: &devopsv1alpha3.GitlabSource{
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    4,
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "bitbucket_server",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			BitbucketServerSource: &devopsv1alpha3.BitbucketServerSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
			},
		},

		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "svn",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			SvnSource: &devopsv1alpha3.SvnSource{
				Remote:       "https://api.svn.com/bcd",
				CredentialId: "svn",
				Excludes:     "truck",
				Includes:     "tag/*",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "single_svn",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			SingleSvnSource: &devopsv1alpha3.SingleSvnSource{
				Remote:       "https://api.svn.com/bcd",
				CredentialId: "svn",
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("\ninput [%+v] \noutput [%+v] \nshould equal ", input.GitlabSource.CloneOption, output.GitlabSource.CloneOption)
		}
	}
}

func Test_MultiBranchPipelineCloneConfig(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource: &devopsv1alpha3.GitSource{
				Url:              "https://github.com/kubesphere/devops",
				CredentialId:     "git",
				DiscoverBranches: true,
				CloneOption: &devopsv1alpha3.GitCloneOption{
					Shallow: false,
					Depth:   3,
					Timeout: 20,
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				CloneOption: &devopsv1alpha3.GitCloneOption{
					Shallow: false,
					Depth:   3,
					Timeout: 20,
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "gitlab",
			GitlabSource: &devopsv1alpha3.GitlabSource{
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				CloneOption: &devopsv1alpha3.GitCloneOption{
					Depth:   -1,
					Timeout: -1,
				},
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}

		// we'll give it a default value if it's negative
		if input.GitlabSource != nil && input.GitlabSource.CloneOption != nil {
			if input.GitlabSource.CloneOption.Timeout < 0 {
				input.GitlabSource.CloneOption.Timeout = 10
			}
			if input.GitlabSource.CloneOption.Depth < 0 {
				input.GitlabSource.CloneOption.Depth = 1
			}
		}

		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input.GitlabSource, output.GitlabSource)
		}
	}

}

func Test_MultiBranchPipelineRegexFilter(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource: &devopsv1alpha3.GitSource{
				Url:              "https://github.com/kubesphere/devops",
				CredentialId:     "git",
				DiscoverBranches: true,
				RegexFilter:      ".*",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}

}

func Test_MultiBranchPipelineMultibranchTrigger(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &devopsv1alpha3.MultiBranchJobTrigger{
				CreateActionJobsToTrigger: "abc",
				DeleteActionJobsToTrigger: "ddd",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &devopsv1alpha3.MultiBranchJobTrigger{
				CreateActionJobsToTrigger: "abc",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &devopsv1alpha3.MultiBranchJobTrigger{
				DeleteActionJobsToTrigger: "ddd",
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}

}

func TestForkPullRequestDiscoveryTrait(t *testing.T) {
	// for gitlab cases
	pipeline, err := parseMultiBranchPipelineConfigXml(noTrustForGitlabJobXML)
	assert.Nil(t, err)
	assert.Equal(t, "gitlab", pipeline.SourceType)

	pipeline, err = parseMultiBranchPipelineConfigXml(withTrustForGitlabJobXML)
	assert.Nil(t, err)
	assert.Equal(t, "gitlab", pipeline.SourceType)
	assert.Equal(t, internal.PRDiscoverTrustEveryone.Value(), pipeline.GitlabSource.DiscoverPRFromForks.Trust)

	// for github cases
	pipeline, err = parseMultiBranchPipelineConfigXml(noTrustForGitHubJobXML)
	assert.Nil(t, err)
	assert.Equal(t, "github", pipeline.SourceType)

	pipeline, err = parseMultiBranchPipelineConfigXml(withTrustForGitHubJobXML)
	assert.Nil(t, err)
	assert.Equal(t, "github", pipeline.SourceType)
	assert.Equal(t, internal.PRDiscoverTrustEveryone.Value(), pipeline.GitHubSource.DiscoverPRFromForks.Trust)

	// for bitbucket cases
	pipeline, err = parseMultiBranchPipelineConfigXml(noTrustForBitbucketJobXML)
	assert.Nil(t, err)
	assert.Equal(t, "bitbucket_server", pipeline.SourceType)

	pipeline, err = parseMultiBranchPipelineConfigXml(withTrustForBitbucketJobXML)
	assert.Nil(t, err)
	assert.Equal(t, "bitbucket_server", pipeline.SourceType)
	assert.Equal(t, internal.BitbucketPRDiscoverTrustTeamForks.Value(), pipeline.BitbucketServerSource.DiscoverPRFromForks.Trust)
}

var noTrustForGitlabJobXML = `<?xml version='1.1' encoding='UTF-8'?>
<org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch@2.22">
  <actions/>
  <description></description>
  <properties>
    <org.jenkinsci.plugins.docker.workflow.declarative.FolderConfig plugin="docker-workflow@1.24">
      <dockerLabel></dockerLabel>
      <registry plugin="docker-commons@1.17"/>
    </org.jenkinsci.plugins.docker.workflow.declarative.FolderConfig>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api@2.6.2">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics>
    <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric plugin="cloudbees-folder@6.15">
      <nonRecursive>false</nonRecursive>
    </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api@2.6.2">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
  <orphanedItemStrategy class="com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy" plugin="cloudbees-folder@6.15">
    <pruneDeadBranches>true</pruneDeadBranches>
    <daysToKeep>7</daysToKeep>
    <numToKeep>5</numToKeep>
  </orphanedItemStrategy>
  <triggers/>
  <disabled>false</disabled>
  <sources class="jenkins.branch.MultiBranchProject$BranchSourceList" plugin="branch-api@2.6.2">
    <data>
      <jenkins.branch.BranchSource>
        <source class="io.jenkins.plugins.gitlabbranchsource.GitLabSCMSource" plugin="gitlab-branch-source@1.5.4">
          <id></id>
          <serverName>https://gitlab.com</serverName>
          <projectOwner>linuxsuren1</projectOwner>
          <projectPath>LinuxSuRen1/learn-pipeline-java</projectPath>
          <credentialsId></credentialsId>
          <traits>
            <io.jenkins.plugins.gitlabbranchsource.BranchDiscoveryTrait>
              <strategyId>1</strategyId>
            </io.jenkins.plugins.gitlabbranchsource.BranchDiscoveryTrait>
            <io.jenkins.plugins.gitlabbranchsource.TagDiscoveryTrait/>
            <io.jenkins.plugins.gitlabbranchsource.OriginMergeRequestDiscoveryTrait>
              <strategyId>2</strategyId>
            </io.jenkins.plugins.gitlabbranchsource.OriginMergeRequestDiscoveryTrait>
            <io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait>
              <strategyId>2</strategyId>
            </io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait>
            <jenkins.scm.impl.trait.RegexSCMHeadFilterTrait plugin="scm-api@2.6.4">
              <regex>master|v.*</regex>
            </jenkins.scm.impl.trait.RegexSCMHeadFilterTrait>
          </traits>
          <sshRemote>git@gitlab.com:LinuxSuRen1/learn-pipeline-java.git</sshRemote>
          <httpRemote>https://gitlab.com/LinuxSuRen1/learn-pipeline-java.git</httpRemote>
          <projectId>23723790</projectId>
        </source>
        <strategy class="jenkins.branch.NamedExceptionsBranchPropertyStrategy">
          <defaultProperties class="empty-list"/>
          <namedExceptions class="empty-list"/>
        </strategy>
      </jenkins.branch.BranchSource>
    </data>
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </sources>
  <factory class="org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <scriptPath>Jenkinsfile</scriptPath>
  </factory>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`

var withTrustForGitlabJobXML = `<?xml version='1.1' encoding='UTF-8'?>
<org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch@2.22">
  <actions/>
  <description></description>
  <properties>
    <org.jenkinsci.plugins.docker.workflow.declarative.FolderConfig plugin="docker-workflow@1.24">
      <dockerLabel></dockerLabel>
      <registry plugin="docker-commons@1.17"/>
    </org.jenkinsci.plugins.docker.workflow.declarative.FolderConfig>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api@2.6.2">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics>
    <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric plugin="cloudbees-folder@6.15">
      <nonRecursive>false</nonRecursive>
    </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api@2.6.2">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
  <orphanedItemStrategy class="com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy" plugin="cloudbees-folder@6.15">
    <pruneDeadBranches>true</pruneDeadBranches>
    <daysToKeep>7</daysToKeep>
    <numToKeep>5</numToKeep>
  </orphanedItemStrategy>
  <triggers/>
  <disabled>false</disabled>
  <sources class="jenkins.branch.MultiBranchProject$BranchSourceList" plugin="branch-api@2.6.2">
    <data>
      <jenkins.branch.BranchSource>
        <source class="io.jenkins.plugins.gitlabbranchsource.GitLabSCMSource" plugin="gitlab-branch-source@1.5.4">
          <id></id>
          <serverName>https://gitlab.com</serverName>
          <projectOwner>linuxsuren1</projectOwner>
          <projectPath>LinuxSuRen1/learn-pipeline-java</projectPath>
          <credentialsId></credentialsId>
          <traits>
            <io.jenkins.plugins.gitlabbranchsource.BranchDiscoveryTrait>
              <strategyId>1</strategyId>
            </io.jenkins.plugins.gitlabbranchsource.BranchDiscoveryTrait>
            <io.jenkins.plugins.gitlabbranchsource.TagDiscoveryTrait/>
            <io.jenkins.plugins.gitlabbranchsource.OriginMergeRequestDiscoveryTrait>
              <strategyId>2</strategyId>
            </io.jenkins.plugins.gitlabbranchsource.OriginMergeRequestDiscoveryTrait>
            <io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait>
              <strategyId>2</strategyId>
              <trust class="io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait$TrustEveryone"/>
            </io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait>
            <jenkins.scm.impl.trait.RegexSCMHeadFilterTrait plugin="scm-api@2.6.4">
              <regex>master|v.*</regex>
            </jenkins.scm.impl.trait.RegexSCMHeadFilterTrait>
          </traits>
          <sshRemote>git@gitlab.com:LinuxSuRen1/learn-pipeline-java.git</sshRemote>
          <httpRemote>https://gitlab.com/LinuxSuRen1/learn-pipeline-java.git</httpRemote>
          <projectId>23723790</projectId>
        </source>
        <strategy class="jenkins.branch.NamedExceptionsBranchPropertyStrategy">
          <defaultProperties class="empty-list"/>
          <namedExceptions class="empty-list"/>
        </strategy>
      </jenkins.branch.BranchSource>
    </data>
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </sources>
  <factory class="org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <scriptPath>Jenkinsfile</scriptPath>
  </factory>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`

var noTrustForGitHubJobXML = `<?xml version="1.1" encoding="UTF-8"?><org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch">
  <actions/>
  <properties>
    <org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig plugin="pipeline-model-definition">
      <dockerLabel/>
      <registry plugin="docker-commons"/>
    </org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics>
    <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric plugin="cloudbees-folder">
      <nonRecursive>false</nonRecursive>
    </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
  <description/>
  <orphanedItemStrategy class="com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy" plugin="cloudbees-folder">
    <pruneDeadBranches>true</pruneDeadBranches>
    <daysToKeep>7</daysToKeep>
    <numToKeep>5</numToKeep>
  </orphanedItemStrategy>
  <triggers/>
  <sources class="jenkins.branch.MultiBranchProject$BranchSourceList" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <data>
      <jenkins.branch.BranchSource>
        <strategy class="jenkins.branch.NamedExceptionsBranchPropertyStrategy">
          <defaultProperties class="empty-list"/>
          <namedExceptions class="empty-list"/>
        </strategy>
        <source class="org.jenkinsci.plugins.github_branch_source.GitHubSCMSource" plugin="github-branch-source">
          <id/>
          <credentialsId>github</credentialsId>
          <repoOwner>jenkins-zh</repoOwner>
          <repository>jenkins-client-java</repository>
          <traits>
            <org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait>
              <strategyId>1</strategyId>
            </org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait>
            <org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait>
              <strategyId>2</strategyId>
            </org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait>
            <org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait>
              <strategyId>2</strategyId>
            </org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait>
            <org.jenkinsci.plugins.github__branch__source.TagDiscoveryTrait/>
            <jenkins.plugins.git.traits.CloneOptionTrait>
              <extension class="hudson.plugins.git.extensions.impl.CloneOption">
                <shallow>false</shallow>
                <noTags>false</noTags>
                <honorRefspec>true</honorRefspec>
                <reference/>
                <timeout>20</timeout>
                <depth>1</depth>
              </extension>
            </jenkins.plugins.git.traits.CloneOptionTrait>
          </traits>
        </source>
      </jenkins.branch.BranchSource>
    </data>
  </sources>
  <factory class="org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <scriptPath>Jenkinsfile</scriptPath>
  </factory>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`

var withTrustForGitHubJobXML = `<?xml version="1.1" encoding="UTF-8"?><org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch">
  <actions/>
  <properties>
    <org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig plugin="pipeline-model-definition">
      <dockerLabel/>
      <registry plugin="docker-commons"/>
    </org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics>
    <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric plugin="cloudbees-folder">
      <nonRecursive>false</nonRecursive>
    </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
  <description/>
  <orphanedItemStrategy class="com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy" plugin="cloudbees-folder">
    <pruneDeadBranches>true</pruneDeadBranches>
    <daysToKeep>7</daysToKeep>
    <numToKeep>5</numToKeep>
  </orphanedItemStrategy>
  <triggers/>
  <sources class="jenkins.branch.MultiBranchProject$BranchSourceList" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <data>
      <jenkins.branch.BranchSource>
        <strategy class="jenkins.branch.NamedExceptionsBranchPropertyStrategy">
          <defaultProperties class="empty-list"/>
          <namedExceptions class="empty-list"/>
        </strategy>
        <source class="org.jenkinsci.plugins.github_branch_source.GitHubSCMSource" plugin="github-branch-source">
          <id/>
          <credentialsId>github</credentialsId>
          <repoOwner>jenkins-zh</repoOwner>
          <repository>jenkins-client-java</repository>
          <traits>
            <org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait>
              <strategyId>1</strategyId>
            </org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait>
            <org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait>
              <strategyId>2</strategyId>
            </org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait>
            <org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait>
              <strategyId>2</strategyId>
              <trust class="org.jenkinsci.plugins.github_branch_source.ForkPullRequestDiscoveryTrait$TrustEveryone"/>
            </org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait>
            <org.jenkinsci.plugins.github__branch__source.TagDiscoveryTrait/>
            <jenkins.plugins.git.traits.CloneOptionTrait>
              <extension class="hudson.plugins.git.extensions.impl.CloneOption">
                <shallow>false</shallow>
                <noTags>false</noTags>
                <honorRefspec>true</honorRefspec>
                <reference/>
                <timeout>20</timeout>
                <depth>1</depth>
              </extension>
            </jenkins.plugins.git.traits.CloneOptionTrait>
          </traits>
        </source>
      </jenkins.branch.BranchSource>
    </data>
  </sources>
  <factory class="org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <scriptPath>Jenkinsfile</scriptPath>
  </factory>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`

var noTrustForBitbucketJobXML = `<?xml version='1.1' encoding='UTF-8'?>
<org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch@2.22">
  <actions/>
  <description></description>
  <properties>
    <org.jenkinsci.plugins.docker.workflow.declarative.FolderConfig plugin="docker-workflow@1.24">
      <dockerLabel></dockerLabel>
      <registry plugin="docker-commons@1.17"/>
    </org.jenkinsci.plugins.docker.workflow.declarative.FolderConfig>
    <org.csanchez.jenkins.plugins.kubernetes.KubernetesFolderProperty plugin="kubernetes@1.27.5">
      <permittedClouds/>
    </org.csanchez.jenkins.plugins.kubernetes.KubernetesFolderProperty>
    <org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty plugin="multibranch-action-triggers@1.8.5">
      <createActionJobsToTrigger></createActionJobsToTrigger>
      <deleteActionJobsToTrigger></deleteActionJobsToTrigger>
      <actionJobsToTriggerOnRunDelete></actionJobsToTriggerOnRunDelete>
      <quitePeriod>0</quitePeriod>
      <branchIncludeFilter>*</branchIncludeFilter>
      <branchExcludeFilter></branchExcludeFilter>
      <additionalParameters/>
    </org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api@2.6.2">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics/>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api@2.6.2">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
  <orphanedItemStrategy class="com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy" plugin="cloudbees-folder@6.15">
    <pruneDeadBranches>true</pruneDeadBranches>
    <daysToKeep>-1</daysToKeep>
    <numToKeep>-1</numToKeep>
  </orphanedItemStrategy>
  <triggers/>
  <disabled>false</disabled>
  <sources class="jenkins.branch.MultiBranchProject$BranchSourceList" plugin="branch-api@2.6.2">
    <data>
      <jenkins.branch.BranchSource>
        <source class="com.cloudbees.jenkins.plugins.bitbucket.BitbucketSCMSource" plugin="cloudbees-bitbucket-branch-source@2.9.4">
          <id>edb85ed8-7609-40ac-9bf8-30cdcd762390</id>
          <serverUrl>https://bitbucket.org</serverUrl>
          <repoOwner>linuxsuren</repoOwner>
          <repository>jenkins-cli</repository>
          <traits>
            <com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait>
              <strategyId>1</strategyId>
            </com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait>
            <com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait>
              <strategyId>1</strategyId>
            </com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait>
            <com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait>
              <strategyId>1</strategyId>
            </com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait>
          </traits>
        </source>
        <strategy class="jenkins.branch.DefaultBranchPropertyStrategy">
          <properties class="empty-list"/>
        </strategy>
      </jenkins.branch.BranchSource>
    </data>
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </sources>
  <factory class="org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <scriptPath>Jenkinsfile</scriptPath>
  </factory>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`

var withTrustForBitbucketJobXML = `<?xml version='1.1' encoding='UTF-8'?>
<org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch@2.22">
  <actions/>
  <description></description>
  <properties>
    <org.jenkinsci.plugins.docker.workflow.declarative.FolderConfig plugin="docker-workflow@1.24">
      <dockerLabel></dockerLabel>
      <registry plugin="docker-commons@1.17"/>
    </org.jenkinsci.plugins.docker.workflow.declarative.FolderConfig>
    <org.csanchez.jenkins.plugins.kubernetes.KubernetesFolderProperty plugin="kubernetes@1.27.5">
      <permittedClouds/>
    </org.csanchez.jenkins.plugins.kubernetes.KubernetesFolderProperty>
    <org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty plugin="multibranch-action-triggers@1.8.5">
      <createActionJobsToTrigger></createActionJobsToTrigger>
      <deleteActionJobsToTrigger></deleteActionJobsToTrigger>
      <actionJobsToTriggerOnRunDelete></actionJobsToTriggerOnRunDelete>
      <quitePeriod>0</quitePeriod>
      <branchIncludeFilter>*</branchIncludeFilter>
      <branchExcludeFilter></branchExcludeFilter>
      <additionalParameters/>
    </org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api@2.6.2">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics/>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api@2.6.2">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
  <orphanedItemStrategy class="com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy" plugin="cloudbees-folder@6.15">
    <pruneDeadBranches>true</pruneDeadBranches>
    <daysToKeep>-1</daysToKeep>
    <numToKeep>-1</numToKeep>
  </orphanedItemStrategy>
  <triggers/>
  <disabled>false</disabled>
  <sources class="jenkins.branch.MultiBranchProject$BranchSourceList" plugin="branch-api@2.6.2">
    <data>
      <jenkins.branch.BranchSource>
        <source class="com.cloudbees.jenkins.plugins.bitbucket.BitbucketSCMSource" plugin="cloudbees-bitbucket-branch-source@2.9.4">
          <id>edb85ed8-7609-40ac-9bf8-30cdcd762390</id>
          <serverUrl>https://bitbucket.org</serverUrl>
          <repoOwner>linuxsuren</repoOwner>
          <repository>jenkins-cli</repository>
          <traits>
            <com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait>
              <strategyId>1</strategyId>
            </com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait>
            <com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait>
              <strategyId>1</strategyId>
            </com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait>
            <com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait>
              <strategyId>1</strategyId>
              <trust class="com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait$TrustTeamForks"/>
            </com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait>
          </traits>
        </source>
        <strategy class="jenkins.branch.DefaultBranchPropertyStrategy">
          <properties class="empty-list"/>
        </strategy>
      </jenkins.branch.BranchSource>
    </data>
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </sources>
  <factory class="org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <scriptPath>Jenkinsfile</scriptPath>
  </factory>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`
