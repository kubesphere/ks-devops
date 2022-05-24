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

package webhook

import (
	"context"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/bitbucket"
	"github.com/jenkins-x/go-scm/scm/driver/github"
	"github.com/jenkins-x/go-scm/scm/driver/gitlab"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/jwt/token"
	"kubesphere.io/devops/pkg/kapis/devops/v1alpha3/pipelinerun"
	models "kubesphere.io/devops/pkg/models/pipeline"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

// tokenExpireIn indicates that the temporary token issued by controller will be expired in some time.
const tokenExpireIn time.Duration = 5 * time.Minute

// SCMHandler handles requests from webhooks.
type SCMHandler struct {
	client.Client
	issue   token.Issuer
	jenkins core.JenkinsCore
}

// NewSCMHandler creates a new handler for handling webhooks.
func NewSCMHandler(genericClient client.Client, issue token.Issuer, jenkins core.JenkinsCore) *SCMHandler {
	return &SCMHandler{
		Client:  genericClient,
		issue:   issue,
		jenkins: jenkins,
	}
}

func getSCMClient(request *http.Request) *scm.Client {
	if request.Header.Get("X-Gitlab-Event") != "" {
		return gitlab.NewDefault()
	}

	if request.Header.Get("X-GitHub-Event") != "" {
		return github.NewDefault()
	}

	if strings.HasPrefix(request.Header.Get("User-Agent"), "Bitbucket-Webhooks") {
		return bitbucket.NewDefault()
	}
	return nil
}

func (h *SCMHandler) scmWebhook(request *restful.Request, response *restful.Response) {
	scmClient := getSCMClient(request.Request)
	if scmClient == nil {
		_, _ = response.Write([]byte("unknown SCM type"))
		return
	}

	webhook, err := scmClient.Webhooks.Parse(request.Request, func(webhook scm.Webhook) (string, error) {
		return "", nil
	})
	if err != nil {
		_, _ = response.Write([]byte(err.Error()))
		return
	}

	ctx := context.TODO()
	if webhook.Kind() == scm.WebhookKindPush {
		repo := webhook.Repository()
		link := repo.Link

		pushHook := webhook.(*scm.PushHook)

		pipelineList := &v1alpha3.PipelineList{}
		if err = h.List(ctx, pipelineList); err == nil {
			found := false
			for i := range pipelineList.Items {
				pipeline := pipelineList.Items[i]

				if pipeline.Spec.MultiBranchPipeline != nil {
					gitURL := pipeline.Spec.MultiBranchPipeline.GetGitURL()
					if gitURL != "" {
						if gitRepoMatch(gitURL, link) {
							h.createPipelineRun(pipeline, pushHook)
							found = true
							break
						}
					}
				}
			}

			if !found {
				_, _ = response.Write([]byte("no pipeline matched"))
				return
			}
		}
	}

	_, _ = response.Write([]byte("ok"))
}

func (h *SCMHandler) createPipelineRun(pipeline v1alpha3.Pipeline, hook *scm.PushHook) {
	branch := strings.TrimPrefix(hook.Ref, "refs/heads/")
	if !branchContains(pipeline, branch) {
		if pipeline.IsMultiBranch() {
			_ = scanJenkinsMultiBranchPipeline(pipeline, h.jenkins, h.issue)
		}
		return
	}

	if noChanges(pipeline, branch) {
		return
	}

	scm, err := pipelinerun.CreateScm(&pipeline.Spec, branch)
	fmt.Println(err)

	run := pipelinerun.CreatePipelineRun(&pipeline, &devops.RunPayload{}, scm)
	err = h.Create(context.Background(), run)
	fmt.Println(err)
}

func scanJenkinsMultiBranchPipeline(pipeline v1alpha3.Pipeline, jenkins core.JenkinsCore, issue token.Issuer) (err error) {
	var accessToken string
	accessToken, err = issue.IssueTo(&user.DefaultInfo{Name: "admin"}, token.AccessToken, tokenExpireIn)
	if err != nil {
		err = fmt.Errorf("failed to issue access token for creator webhook, error was %v", err)
		return
	}

	// using a dynamic Jenkins token instead of the static one
	jenkins.Token = accessToken
	jclient := job.Client{
		JenkinsCore: jenkins,
	}

	err = jclient.Build(fmt.Sprintf("%s %s", pipeline.Namespace, pipeline.Name))
	return
}

func branchContains(pipeline v1alpha3.Pipeline, branch string) (ok bool) {
	branchesJSONText := pipeline.Annotations[v1alpha3.PipelineJenkinsBranchesAnnoKey]
	branches, err := models.GetBranchSlice(branchesJSONText)
	if err == nil {
		ok, _ = branches.SearchByName(branch)
	}
	return
}

func noChanges(pipeline v1alpha3.Pipeline, branch string) bool {
	return false
}

// gitRepoMatch if the source matches target
func gitRepoMatch(source, target string) (ok bool) {
	ok = source == target
	return
}
