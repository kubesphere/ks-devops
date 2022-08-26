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

package gitrepository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/utils/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PullRequestStatusReconciler reconciles a Pipeline build status to the Pull Requests
type PullRequestStatusReconciler struct {
	client.Client
	ExternalAddress string

	log      logr.Logger
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=webhooks,verbs=get;list;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=secrets,verbs=get

// Reconcile is the main entry of this reconciler
func (r *PullRequestStatusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (
	result ctrl.Result, err error) {
	pipelinerun := &v1alpha3.PipelineRun{}
	if err = r.Get(ctx, req.NamespacedName, pipelinerun); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if !pipelinerun.Spec.IsMultiBranchPipeline() || pipelinerun.Spec.SCM == nil {
		return
	}

	r.log.Info(fmt.Sprintf("start to reconcile %s", req.NamespacedName))
	var prNumber int
	if prNumber, err = getPRNumber(pipelinerun.Spec.SCM.RefName); err != nil {
		err = nil
		return
	}

	repoInfo := getRepoInfo(pipelinerun.Spec.PipelineSpec.MultiBranchPipeline)
	if repoInfo.isInvalid() {
		return
	}

	var (
		token    string
		username string
	)
	if username, token, err = r.getTokenFromSecret(&v1.SecretReference{
		Name:      repoInfo.tokenId,
		Namespace: pipelinerun.Namespace,
	}, ""); err != nil {
		err = fmt.Errorf("failed to get token, error %v", err)
		return
	}

	repo := repoInfo.getRepoPath()
	r.log.Info(fmt.Sprintf("start sending status to %s with pr %d", repo, prNumber))

	var target string
	if target, err = r.getExternalPipelineRunAddress(ctx, pipelinerun); err != nil {
		return
	}

	maker := NewStatusMaker(repo, token)
	maker.WithTarget(target).WithPR(prNumber).WithProvider(repoInfo.provider).WithUsername(username)

	err = maker.CreateWithPipelinePhase(ctx, pipelinerun.Status.Phase, "KubeSphere DevOps", string(pipelinerun.Status.Phase))
	if err != nil {
		r.log.Error(err, "failed to send status")
	}
	return
}

type repoInfo struct {
	provider string
	owner    string
	repo     string
	tokenId  string
}

func (r repoInfo) getRepoPath() string {
	return fmt.Sprintf("%s/%s", r.owner, r.repo)
}

func (r repoInfo) isInvalid() bool {
	return r.provider == "" || r.owner == "" || r.repo == "" || r.tokenId == ""
}

func getRepoInfo(repo *v1alpha3.MultiBranchPipeline) (info repoInfo) {
	if repo == nil {
		return
	}
	switch repo.SourceType {
	case v1alpha3.SourceTypeBitbucket:
		if repo.BitbucketServerSource != nil {
			info.provider = "bitbucket"
			info.owner = repo.BitbucketServerSource.Owner
			info.repo = repo.BitbucketServerSource.Repo
			info.tokenId = repo.BitbucketServerSource.CredentialId
			if repo.BitbucketServerSource.ApiUri == "https://bitbucket.org" {
				info.provider = "bitbucketcloud"
			}
		}
	case v1alpha3.SourceTypeGithub:
		if repo.GitHubSource != nil {
			info.provider = "github"
			info.owner = repo.GitHubSource.Owner
			info.repo = repo.GitHubSource.Repo
			info.tokenId = repo.GitHubSource.CredentialId
		}
	case v1alpha3.SourceTypeGitlab:
		if repo.GitlabSource != nil {
			info.provider = "gitlab"
			info.owner = repo.GitlabSource.Owner
			// the repo format of Gitlab is: owner/repo
			info.repo = strings.TrimPrefix(repo.GitlabSource.Repo, repo.GitlabSource.Owner+"/")
			info.tokenId = repo.GitlabSource.CredentialId
		}
	}
	return
}

func (r *PullRequestStatusReconciler) getExternalPipelineRunAddress(ctx context.Context, pipelineRun *v1alpha3.PipelineRun) (target string, err error) {
	var ws string
	if ws, err = r.getWorkspace(ctx, pipelineRun.GetNamespace()); err == nil {
		target = fmt.Sprintf("%s/%s/clusters/default/devops/%s/pipelines/%s/run/%s/task-status",
			net.ParseURL(r.ExternalAddress), ws, pipelineRun.Namespace, pipelineRun.Spec.PipelineRef.Name, pipelineRun.Name)
	}
	return
}

func (r *PullRequestStatusReconciler) getWorkspace(ctx context.Context, ns string) (ws string, err error) {
	project := &v1alpha3.DevOpsProject{}
	if err = r.Get(ctx, types.NamespacedName{
		Name: ns,
	}, project); err == nil {
		ws = project.GetLabels()["kubesphere.io/workspace"]
	}
	return
}

func (r *PullRequestStatusReconciler) getTokenFromSecret(secretRef *v1.SecretReference, defaultNamespace string) (username, token string, err error) {
	var gitSecret *v1.Secret
	if gitSecret, err = r.getSecret(secretRef, defaultNamespace); err == nil {
		switch gitSecret.Type {
		case v1.SecretTypeBasicAuth, v1alpha3.SecretTypeBasicAuth:
			token = string(gitSecret.Data[v1.BasicAuthPasswordKey])
			username = string(gitSecret.Data[v1.BasicAuthUsernameKey])
		case v1.SecretTypeOpaque, v1alpha3.SecretTypeSecretText:
			token = string(gitSecret.Data[v1.ServiceAccountTokenKey])
		}
	}
	return
}

func (r *PullRequestStatusReconciler) getSecret(ref *v1.SecretReference, defaultNamespace string) (secret *v1.Secret, err error) {
	secret = &v1.Secret{}
	ns := ref.Namespace
	if ns == "" {
		ns = defaultNamespace
	}

	if err = r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: ns, Name: ref.Name,
	}, secret); err != nil {
		err = fmt.Errorf("cannot get secret %v, error is: %v", secret, err)
	}
	return
}

func getPRNumber(pr string) (int, error) {
	pr = strings.ToLower(pr)
	pr = strings.TrimPrefix(pr, "pr-")
	pr = strings.TrimPrefix(pr, "mr-")
	return strconv.Atoi(pr)
}

// GetName returns the name of this reconciler
func (r *PullRequestStatusReconciler) GetName() string {
	return "pull-request-status-controller"
}

// GetGroupName returns the gorup name of the set of reconcilers
func (r *PullRequestStatusReconciler) GetGroupName() string {
	return groupName
}

// SetupWithManager sets up the controller with the Manager.
func (r *PullRequestStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	r.log = ctrl.Log.WithName(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha3.PipelineRun{}).
		Complete(r)
}

// StatusMaker responsible for Pull Requests status creating
type StatusMaker struct {
	provider string
	server   string
	repo     string
	pr       int
	token    string
	username string
	target   string
}

// NewStatusMaker creates an instance of statusMaker
func NewStatusMaker(repo, token string) *StatusMaker {
	return &StatusMaker{
		repo:  repo,
		token: token,
	}
}

// WithUsername sets the username
func (s *StatusMaker) WithUsername(username string) *StatusMaker {
	s.username = username
	return s
}

// WithProvider sets the provider
func (s *StatusMaker) WithProvider(provider string) *StatusMaker {
	s.provider = provider
	return s
}

// WithProvider sets the server
func (s *StatusMaker) WithServer(server string) *StatusMaker {
	s.server = server
	return s
}

// WithTarget sets the target URL
func (s *StatusMaker) WithTarget(target string) *StatusMaker {
	s.target = target
	return s
}

// WithPR sets the pr number
func (s *StatusMaker) WithPR(pr int) *StatusMaker {
	s.pr = pr
	return s
}

// Create creates a generic status
func (s *StatusMaker) Create(ctx context.Context, status scm.State, label, desc string) (err error) {
	var client *scm.Client
	client, err = factory.NewClient(s.provider, s.server, s.token, func(c *scm.Client) {
		c.Username = s.username
	})
	if err != nil {
		return
	}

	var pullRequest *scm.PullRequest
	if pullRequest, _, err = client.PullRequests.Find(ctx, s.repo, s.pr); err == nil {
		_, _, err = client.Repositories.CreateStatus(ctx, s.repo, pullRequest.Sha, &scm.StatusInput{
			Desc:   desc,
			Label:  label,
			State:  status,
			Target: s.target,
		})
	}
	return
}

// CreateWithPipelinePhase creates a generic status with the PipelineRun phase
func (s *StatusMaker) CreateWithPipelinePhase(ctx context.Context, phase v1alpha3.RunPhase, label, desc string) (err error) {
	return s.Create(ctx, convertPipelineRunPhaseToSCMStatus(phase), label, desc)
}

func convertPipelineRunPhaseToSCMStatus(phase v1alpha3.RunPhase) (status scm.State) {
	switch phase {
	case v1alpha3.Pending:
		status = scm.StatePending
	case v1alpha3.Failed:
		status = scm.StateFailure
	case v1alpha3.Running:
		status = scm.StateRunning
	case v1alpha3.Succeeded:
		status = scm.StateSuccess
	case v1alpha3.Cancelled:
		status = scm.StateCanceled
	default:
		status = scm.StateUnknown
	}
	return status
}
