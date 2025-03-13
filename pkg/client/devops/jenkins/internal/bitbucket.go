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
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"k8s.io/klog/v2"

	devopsv1alpha3 "github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
)

func AppendBitbucketServerSourceToEtree(source *etree.Element, gitSource *devopsv1alpha3.BitbucketServerSource) {
	if gitSource == nil {
		klog.Warning("please provide BitbucketServer source when the sourceType is BitbucketServer")
		return
	}
	source.CreateAttr("class", "com.cloudbees.jenkins.plugins.bitbucket.BitbucketSCMSource")
	source.CreateAttr("plugin", "cloudbees-bitbucket-branch-source")
	source.CreateElement("id").SetText(gitSource.ScmId)
	source.CreateElement("credentialsId").SetText(gitSource.CredentialId)
	source.CreateElement("repoOwner").SetText(gitSource.Owner)
	source.CreateElement("repository").SetText(gitSource.Repo)
	source.CreateElement("serverUrl").SetText(gitSource.ApiUri)

	traits := source.CreateElement("traits")
	if gitSource.DiscoverBranches != 0 {
		traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait>").
			CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverBranches))
	}
	if gitSource.DiscoverPRFromOrigin != 0 {
		traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverPRFromOrigin))
	}
	if gitSource.DiscoverPRFromForks != nil {
		forkTrait := traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait")
		forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverPRFromForks.Strategy))
		trustClass := "com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait$"

		if prTrust := PRDiscoverTrust(gitSource.DiscoverPRFromForks.Trust); prTrust.IsValid() {
			trustClass += prTrust.String()
		} else {
			klog.Warningf("invalid Bitbucket discover PR trust value: %d", prTrust.Value())
		}

		forkTrait.CreateElement("trust").CreateAttr("class", trustClass)
	}
	if gitSource.DiscoverTags {
		traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.TagDiscoveryTrait")
	}
	if gitSource.CloneOption != nil {
		cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
		cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
		cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(gitSource.CloneOption.Shallow))
		cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
		cloneExtension.CreateElement("honorRefspec").SetText(strconv.FormatBool(true))
		cloneExtension.CreateElement("reference")
		if gitSource.CloneOption.Timeout >= 0 {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(gitSource.CloneOption.Timeout))
		} else {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
		}

		if gitSource.CloneOption.Depth >= 0 {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(gitSource.CloneOption.Depth))
		} else {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
		}
	}
	if gitSource.RegexFilter != "" {
		regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
		regexTraits.CreateAttr("plugin", "scm-api")
		regexTraits.CreateElement("regex").SetText(gitSource.RegexFilter)
	}
	if !gitSource.AcceptJenkinsNotification {
		skipNotifications := traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.notifications.SkipNotificationsTrait")
		skipNotifications.CreateAttr("plugin", "skip-notifications-trait")
	}
	return
}

func GetBitbucketServerSourceFromEtree(source *etree.Element) *devopsv1alpha3.BitbucketServerSource {
	var s devopsv1alpha3.BitbucketServerSource
	if credential := source.SelectElement("credentialsId"); credential != nil {
		s.CredentialId = credential.Text()
	}
	if repoOwner := source.SelectElement("repoOwner"); repoOwner != nil {
		s.Owner = repoOwner.Text()
	}
	if repository := source.SelectElement("repository"); repository != nil {
		s.Repo = repository.Text()
	}
	if apiUri := source.SelectElement("serverUrl"); apiUri != nil {
		s.ApiUri = apiUri.Text()
	}
	traits := source.SelectElement("traits")
	if branchDiscoverTrait := traits.SelectElement(
		"com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(branchDiscoverTrait.SelectElement("strategyId").Text())
		s.DiscoverBranches = strategyId
	}
	if tagDiscoverTrait := traits.SelectElement(
		"com.cloudbees.jenkins.plugins.bitbucket.TagDiscoveryTrait"); tagDiscoverTrait != nil {
		s.DiscoverTags = true
	}
	if originPRDiscoverTrait := traits.SelectElement(
		"com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait"); originPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(originPRDiscoverTrait.SelectElement("strategyId").Text())
		s.DiscoverPRFromOrigin = strategyId
	}
	if forkPRDiscoverTrait := traits.SelectElement(
		"com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait"); forkPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(forkPRDiscoverTrait.SelectElement("strategyId").Text())
		if trustEle := forkPRDiscoverTrait.SelectElement("trust"); trustEle != nil {
			trustClass := trustEle.SelectAttr("class").Value
			trust := strings.Split(trustClass, "$")

			if prTrust := BitbucketPRDiscoverTrust(1).ParseFromString(trust[1]); prTrust.IsValid() {
				s.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: strategyId,
					Trust:    prTrust.Value(),
				}
			} else {
				klog.Warningf("invalid Bitbucket discover PR trust value: %s", trust[1])
			}
		}

		s.CloneOption = parseFromCloneTrait(traits.SelectElement("jenkins.plugins.git.traits.CloneOptionTrait"))

		if regexTrait := traits.SelectElement(
			"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
			if regex := regexTrait.SelectElement("regex"); regex != nil {
				s.RegexFilter = regex.Text()
			}
		}

		if skipNotificationTrait := traits.SelectElement(
			"com.cloudbees.jenkins.plugins.bitbucket.notifications.SkipNotificationsTrait"); skipNotificationTrait == nil {
			s.AcceptJenkinsNotification = true
		}
	}
	return &s
}
