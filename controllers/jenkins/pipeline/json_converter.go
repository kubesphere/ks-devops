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

package pipeline

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/tools/record"
	v1alpha3 "kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/jwt/token"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// tokenExpireIn indicates that the temporary token issued by controller will be expired in some time.
const tokenExpireIn time.Duration = 5 * time.Minute

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines,verbs=get;list;update;patch;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// JenkinsfileReconciler will convert between JSON and Jenkinsfile (as groovy) formats
type JenkinsfileReconciler struct {
	log      logr.Logger
	recorder record.EventRecorder

	client.Client
	JenkinsCore core.JenkinsCore
	TokenIssuer token.Issuer
}

// Reconcile is the main entrypoint of this controller
func (r *JenkinsfileReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()

	pip := &v1alpha3.Pipeline{}
	if err = r.Get(ctx, req.NamespacedName, pip); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if pip.Spec.Type != v1alpha3.NoScmPipelineType || pip.Spec.Pipeline == nil {
		return
	}

	// set up the Jenkins client
	var c *core.JenkinsCore
	if c, err = r.getOrCreateJenkinsCore(map[string]string{
		v1alpha3.PipelineRunCreatorAnnoKey: "admin",
	}); err != nil {
		err = fmt.Errorf("failed to create Jenkins client, error: %v", err)
		return
	}
	c.RoundTripper = r.JenkinsCore.RoundTripper
	coreClient := core.Client{JenkinsCore: *c}

	editMode := pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey]
	switch editMode {
	case v1alpha3.PipelineJenkinsfileEditModeRaw:
		result, err = r.reconcileJenkinsfileEditMode(pip, coreClient)
	case v1alpha3.PipelineJenkinsfileEditModeJSON:
		result, err = r.reconcileJSONEditMode(pip, coreClient)
	case "":
	default:
		r.log.Info(fmt.Sprintf("invalid edit mode: %s", editMode))
		return
	}
	return
}

func (r *JenkinsfileReconciler) reconcileJenkinsfileEditMode(pip *v1alpha3.Pipeline, coreClient core.Client) (
	result ctrl.Result, err error) {
	jenkinsfile := pip.Spec.Pipeline.Jenkinsfile

	var toJSONResult core.GenericResult
	if toJSONResult, err = coreClient.ToJSON(jenkinsfile); err != nil || toJSONResult.GetStatus() != "success" {
		err = fmt.Errorf("failed to convert Jenkinsfile to JSON format, error: %v", err)
		return
	}

	if pip.Annotations == nil {
		pip.Annotations = map[string]string{}
	}
	pip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey] = toJSONResult.GetResult()
	pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = ""
	err = r.Update(context.Background(), pip)
	return
}

func (r *JenkinsfileReconciler) reconcileJSONEditMode(pip *v1alpha3.Pipeline, coreClient core.Client) (
	result ctrl.Result, err error) {
	var jsonData string
	if jsonData = pip.Annotations[v1alpha3.PipelineJenkinsfileValueAnnoKey]; jsonData != "" {
		var toResult core.GenericResult
		if toResult, err = coreClient.ToJenkinsfile(jsonData); err != nil || toResult.GetStatus() != "success" {
			err = fmt.Errorf("failed to convert JSON format to Jenkinsfile, error: %v", err)
			return
		}

		pip.Annotations[v1alpha3.PipelineJenkinsfileEditModeAnnoKey] = ""
		pip.Spec.Pipeline.Jenkinsfile = toResult.GetResult()
		err = r.Update(context.Background(), pip)
	}
	return
}

// GetName returns the name of this controller
func (r *JenkinsfileReconciler) GetName() string {
	return "JenkinsfileController"
}

// GetGroupName returns the group name of this controller
func (r *JenkinsfileReconciler) GetGroupName() string {
	return ControllerGroupName
}

func (r *JenkinsfileReconciler) getOrCreateJenkinsCore(annotations map[string]string) (*core.JenkinsCore, error) {
	creator, ok := annotations[v1alpha3.PipelineRunCreatorAnnoKey]
	if !ok || creator == "" {
		return &r.JenkinsCore, nil
	}
	// create a new JenkinsCore for current creator
	accessToken, err := r.TokenIssuer.IssueTo(&user.DefaultInfo{Name: creator}, token.AccessToken, tokenExpireIn)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token for creator %s, error was %v", creator, err)
	}
	jenkinsCore := &core.JenkinsCore{
		URL:      r.JenkinsCore.URL,
		UserName: creator,
		Token:    accessToken,
	}
	return jenkinsCore, nil
}

// SetupWithManager setups the log and recorder
func (r *JenkinsfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName(r.GetName())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha3.Pipeline{}).
		Complete(r)
}
