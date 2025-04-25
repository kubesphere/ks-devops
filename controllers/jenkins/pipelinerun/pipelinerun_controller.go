/*
Copyright 2020 The KubeSphere Authors.

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

package pipelinerun

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	devopsClient "github.com/kubesphere/ks-devops/pkg/client/devops"
	"github.com/kubesphere/ks-devops/pkg/client/devops/jenkins"
	cmstore "github.com/kubesphere/ks-devops/pkg/store/configmap"
	storeInter "github.com/kubesphere/ks-devops/pkg/store/store"
	"github.com/kubesphere/ks-devops/pkg/utils/k8sutil"
)

// BuildNotExistMsg indicates the build with pipelinerun-id not exist in jenkins
const BuildNotExistMsg = "not found resources"

// Reconciler reconciles a PipelineRun object
type Reconciler struct {
	client.Client
	req                  ctrl.Request
	ctx                  context.Context
	log                  logr.Logger
	Scheme               *runtime.Scheme
	Options              *jenkins.Options
	DevOpsClient         devopsClient.Interface
	JenkinsCore          core.JenkinsCore
	recorder             record.EventRecorder
	PipelineRunDataStore string
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("PipelineRun", req.NamespacedName)
	r.ctx = ctx
	r.req = req

	// get PipelineRun
	pipelineRun := &v1alpha3.PipelineRun{}
	var err error
	if err = r.Client.Get(ctx, req.NamespacedName, pipelineRun); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	jHandler := &jenkinsHandler{&r.JenkinsCore}

	// don't modify the cache in other places, like informer cache.
	pipelineRunCopied := pipelineRun.DeepCopy()

	// DeletionTimestamp.IsZero() means copyPipeline has not been deleted.
	if !pipelineRunCopied.ObjectMeta.DeletionTimestamp.IsZero() {
		if err = jHandler.deleteJenkinsJobHistory(pipelineRunCopied); err != nil {
			klog.V(4).Infof("failed to delete Jenkins job history from PipelineRun: %s/%s, error: %v",
				pipelineRunCopied.Namespace, pipelineRunCopied.Name, err)
		} else {
			k8sutil.RemoveFinalizer(&pipelineRunCopied.ObjectMeta, v1alpha3.PipelineRunFinalizerName)
			err = r.Update(context.TODO(), pipelineRunCopied)
		}
		return ctrl.Result{}, err
	}

	// the PipelineRun cannot allow building
	if !pipelineRunCopied.Buildable() {
		return ctrl.Result{}, nil
	}

	// check PipelineRef
	if pipelineRunCopied.Spec.PipelineRef == nil || pipelineRunCopied.Spec.PipelineRef.Name == "" {
		// make the PipelineRun as orphan
		return ctrl.Result{}, r.makePipelineRunOrphan(ctx, pipelineRunCopied)
	}

	// get pipeline
	pipeline := &v1alpha3.Pipeline{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: pipelineRunCopied.Namespace, Name: pipelineRunCopied.Spec.PipelineRef.Name}, pipeline); err != nil {
		log.Error(err, "unable to get pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	namespaceName := pipeline.Namespace
	pipelineName := pipeline.GetName()

	// set Pipeline name and SCM ref name into labels
	if pipelineRunCopied.Labels == nil {
		pipelineRunCopied.Labels = make(map[string]string)
	}
	pipelineRunCopied.Labels[v1alpha3.PipelineNameLabelKey] = pipelineName

	log = log.WithValues("namespace", namespaceName, "Pipeline", pipelineName)

	// check PipelineRun status
	if pipelineRunCopied.HasStarted() {
		log.V(5).Info("pipeline has already started, and we are retrieving run data from Jenkins.")
		pipelineBuild, err := jHandler.getPipelineRunResult(namespaceName, pipelineName, pipelineRunCopied)
		if err != nil {
			if err.Error() == BuildNotExistMsg { // retry if get pipelinerun failed by not exist
				runID, _ := pipelineRunCopied.GetPipelineRunID()
				log.Info(fmt.Sprintf("get pipelinerun data(id: %s) error with not exit, retry.", runID))
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}
			log.Error(err, "unable get PipelineRun data.")
			r.recorder.Eventf(pipelineRunCopied, corev1.EventTypeWarning, v1alpha3.RetrieveFailed, "Failed to retrieve running data from Jenkins, and error was %v", err)
			return ctrl.Result{}, err
		}

		nodeDetails, err := jHandler.getPipelineNodeDetails(pipelineName, namespaceName, pipelineRunCopied)
		if err != nil {
			log.Error(err, "unable to get PipelineRun nodes detail")
			r.recorder.Eventf(pipelineRunCopied, corev1.EventTypeWarning, v1alpha3.RetrieveFailed, "Failed to retrieve nodes detail from Jenkins, and error was %v", err)
			return ctrl.Result{}, err
		}
		runResultJSON, err := json.Marshal(pipelineBuild)
		if err != nil {
			log.Error(err, "unable to marshal result data to JSON")
			return ctrl.Result{}, err
		}
		nodeDetailsJSON, err := json.Marshal(nodeDetails)
		if err != nil {
			log.Error(err, "unable to marshal nodes details to JSON")
			return ctrl.Result{}, err
		}

		// store pipelinerun stage to configmap
		if err = r.storePipelineRunData(string(runResultJSON), string(nodeDetailsJSON), pipelineRunCopied); err != nil {
			log.Error(err, "unable to store pipeline stages to configmap.")
			return ctrl.Result{}, err
		}

		// update pipelinerun status with pipelineBuild
		status := pipelineRunCopied.Status.DeepCopy()
		pbApplier := pipelineBuildApplier{pipelineBuild}
		pbApplier.apply(status)
		// Because the status is a subresource of PipelineRun, we have to update status separately.
		// See also: https://book-v1.book.kubebuilder.io/basics/status_subresource.html
		if err := r.updateStatus(ctx, status, req.NamespacedName); err != nil {
			log.Error(err, "unable to update PipelineRun status.")
			return ctrl.Result{}, err
		}

		if err := r.getAgentInfo(ctx, pipelineRunCopied); err != nil {
			log.Error(err, "unable to get agent info")
			r.recorder.Eventf(pipelineRunCopied, corev1.EventTypeWarning, v1alpha3.RetrieveFailed, "Failed to retrieve agent info from Jenkins, and error was %v", err)
			return ctrl.Result{}, err
		}

		r.recorder.Eventf(pipelineRunCopied, corev1.EventTypeNormal, v1alpha3.Updated, "Updated running data for PipelineRun %s", req.NamespacedName)
		// until the status is okay
		// TODO make the RequeueAfter configurable
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}

	// create trigger handler
	triggerHandler := &jenkinsHandler{&r.JenkinsCore}
	// first run
	jobRun, err := triggerHandler.triggerJenkinsJob(namespaceName, pipelineName, &pipelineRunCopied.Spec)
	if err != nil {
		log.Error(err, "unable to run pipeline", "namespace", namespaceName, "pipeline", pipeline.Name)
		r.recorder.Eventf(pipelineRunCopied, corev1.EventTypeWarning, v1alpha3.TriggerFailed, "Failed to trigger PipelineRun %s, and error was %v", req.NamespacedName, err)
		return ctrl.Result{}, err
	}
	// check if there is still a same PipelineRun
	if exists, err := r.hasSamePipelineRun(jobRun, pipeline); err != nil {
		return ctrl.Result{}, err
	} else if exists {
		// if there still exists the same pending PipelineRun, then give up reconciling
		if err := r.Delete(ctx, pipelineRunCopied); err != nil {
			// ignore the not found error here
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		log.Info("Skipped this PipelineRun because there was still a pending Pipeline with the same parameter")
		return ctrl.Result{}, nil
	}

	log.Info("Triggered a PipelineRun", "runID", jobRun.ID)

	// set Jenkins run ID
	if pipelineRunCopied.Annotations == nil {
		pipelineRunCopied.Annotations = make(map[string]string)
	}
	pipelineRunCopied.Annotations[v1alpha3.JenkinsPipelineRunIDAnnoKey] = jobRun.ID

	// the Update method only updates fields except subresource: status
	if err := r.updateLabelsAndAnnotations(ctx, pipelineRunCopied); err != nil {
		log.Error(err, "unable to update PipelineRun labels and annotations.")
		return ctrl.Result{}, err
	}

	pipelineRunCopied.Status.StartTime = &v1.Time{Time: time.Now()}
	pipelineRunCopied.Status.UpdateTime = &v1.Time{Time: time.Now()}
	// due to the status is subresource of PipelineRun, we have to update status separately.
	// see also: https://book-v1.book.kubebuilder.io/basics/status_subresource.html

	if err := r.updateStatus(ctx, &pipelineRunCopied.Status, req.NamespacedName); err != nil {
		log.Error(err, "unable to update PipelineRun status.")
		return ctrl.Result{}, err
	}
	r.recorder.Eventf(pipelineRunCopied, corev1.EventTypeNormal, v1alpha3.Started, "Started PipelineRun %s", req.NamespacedName)
	// requeue after 1 second
	return ctrl.Result{}, nil
}

// match /blue/rest/organizations/jenkins/pipelines/{devops}/{pipeline}/runs/{run}/log/?start=0
// match /blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/log/?
func (r *Reconciler) getAgentInfo(ctx context.Context, pr *v1alpha3.PipelineRun) error {
	if pr.Annotations == nil {
		pr.Annotations = make(map[string]string)
	}

	podName, podNameExist := pr.Annotations[v1alpha3.JenkinsAgentPodNameAnnoKey]
	_, nodeNameExist := pr.Annotations[v1alpha3.JenkinsAgentNodeNameAnnoKey]
	if podNameExist && nodeNameExist || (pr.Status.Phase != v1alpha3.Running && pr.Status.Phase != v1alpha3.Pending) {
		return nil
	}

	if !podNameExist {
		runID, _ := pr.GetPipelineRunID()
		logUrl := jenkins.GetRunLogUrl
		api := fmt.Sprintf(logUrl, pr.Namespace, pr.Spec.PipelineRef.Name, runID)
		if pr.Spec.PipelineSpec.Type == v1alpha3.MultiBranchPipelineType {
			logUrl = jenkins.GetBranchRunLogUrl
			api = fmt.Sprintf(logUrl, pr.Namespace, pr.Spec.PipelineRef.Name, pr.Spec.SCM.RefName, runID)
		}

		_, res, err := r.JenkinsCore.Request(http.MethodGet, api, map[string]string{"Content-Type": "application/json"}, nil)
		if err != nil {
			return err
		}

		lines := strings.Split(string(res), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Agent") && strings.Contains(line, "is provisioned from template") {
				parts := strings.Fields(line)
				if len(parts) > 2 {
					podName = parts[1]
					pr.Annotations[v1alpha3.JenkinsAgentPodNameAnnoKey] = podName
					klog.Infof("get agent pod name: %s", parts[1])
				}
				break
			}
		}
	}

	if podName != "" && !nodeNameExist {
		agentPod := &corev1.Pod{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: r.Options.WorkerNamespace, Name: podName}, agentPod)
		if err != nil {
			return err
		}

		if agentPod.Spec.NodeName != "" {
			pr.Annotations[v1alpha3.JenkinsAgentNodeNameAnnoKey] = agentPod.Spec.NodeName
			klog.Info("get agent pod running node name: ", agentPod.Spec.NodeName)
		}
	}

	return r.updateLabelsAndAnnotations(ctx, pr)
}

func (r *Reconciler) storePipelineRunData(runResultJSON, nodeDetailsJSON string, pipelineRunCopied *v1alpha3.PipelineRun) (err error) {
	if r.PipelineRunDataStore == "" {
		if pipelineRunCopied.Annotations == nil {
			pipelineRunCopied.Annotations = make(map[string]string)
		}
		pipelineRunCopied.Annotations[v1alpha3.JenkinsPipelineRunStatusAnnoKey] = runResultJSON
		pipelineRunCopied.Annotations[v1alpha3.JenkinsPipelineRunStagesStatusAnnoKey] = nodeDetailsJSON

		// update labels and annotations
		if err = r.updateLabelsAndAnnotations(r.ctx, pipelineRunCopied); err != nil {
			r.log.Error(err, "unable to update PipelineRun labels and annotations.")
		}
	} else if r.PipelineRunDataStore == "configmap" {
		var cmStore storeInter.ConfigMapStore
		if cmStore, err = cmstore.NewConfigMapStore(r.ctx, r.req.NamespacedName, r.Client); err == nil {
			cmStore.SetStatus(runResultJSON)
			cmStore.SetStages(nodeDetailsJSON)
			cmStore.SetOwnerReference(v1.OwnerReference{
				APIVersion: pipelineRunCopied.APIVersion,
				Kind:       pipelineRunCopied.Kind,
				Name:       pipelineRunCopied.Name,
				UID:        pipelineRunCopied.UID,
			})
			err = cmStore.Save()
		}
	} else {
		err = fmt.Errorf("unknown pipelineRun data store type: %s", r.PipelineRunDataStore)
	}
	return
}

func (r *Reconciler) hasSamePipelineRun(jobRun *job.PipelineRun, pipeline *v1alpha3.Pipeline) (exists bool, err error) {
	// check if the run ID exists in the PipelineRun
	pipelineRuns := &v1alpha3.PipelineRunList{}
	listOptions := []client.ListOption{
		client.InNamespace(pipeline.Namespace),
		client.MatchingLabels{v1alpha3.PipelineNameLabelKey: pipeline.Name},
	}
	if pipeline.Spec.Type == v1alpha3.MultiBranchPipelineType {
		// add SCM reference name into list options for multi-branch Pipeline
		listOptions = append(listOptions, client.MatchingFields{v1alpha3.PipelineRunSCMRefNameField: jobRun.Pipeline})
	}
	if err = r.Client.List(context.Background(), pipelineRuns, listOptions...); err == nil {
		isMultiBranch := pipeline.Spec.Type == v1alpha3.MultiBranchPipelineType
		finder := newPipelineRunFinder(pipelineRuns.Items)
		_, exists = finder.find(jobRun, isMultiBranch)
	}
	return
}

func getSCMRefName(prSpec *v1alpha3.PipelineRunSpec) (string, error) {
	var branch = ""
	if prSpec.IsMultiBranchPipeline() {
		if prSpec.SCM == nil || prSpec.SCM.RefName == "" {
			return "", fmt.Errorf("failed to obtain SCM reference name for multi-branch Pipeline")
		}
		branch = prSpec.SCM.RefName
	}
	return branch, nil
}

func (r *Reconciler) updateLabelsAndAnnotations(ctx context.Context, pr *v1alpha3.PipelineRun) (err error) {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// get pipeline
		prToUpdate := v1alpha3.PipelineRun{}
		err = r.Get(ctx, client.ObjectKey{Namespace: pr.Namespace, Name: pr.Name}, &prToUpdate)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(pr.Labels, prToUpdate.Labels) && reflect.DeepEqual(pr.Annotations, prToUpdate.Annotations) {
			return nil
		}

		prToUpdate = *prToUpdate.DeepCopy()
		prToUpdate.Labels = pr.Labels
		prToUpdate.Annotations = pr.Annotations
		// make sure all PipelineRuns have the finalizer
		k8sutil.AddFinalizer(&prToUpdate.ObjectMeta, v1alpha3.PipelineRunFinalizerName)
		return r.Update(ctx, &prToUpdate)
	})
}

func (r *Reconciler) updateStatus(ctx context.Context, desiredStatus *v1alpha3.PipelineRunStatus, prKey client.ObjectKey) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		prToUpdate := v1alpha3.PipelineRun{}
		err := r.Get(ctx, prKey, &prToUpdate)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(*desiredStatus, prToUpdate.Status) {
			return nil
		}
		prToUpdate = *prToUpdate.DeepCopy()
		prToUpdate.Status = *desiredStatus
		return r.Status().Update(ctx, &prToUpdate)
	})
}

func (r *Reconciler) makePipelineRunOrphan(ctx context.Context, pr *v1alpha3.PipelineRun) (err error) {
	// make the PipelineRun as orphan
	prToUpdate := pr.DeepCopy()
	prToUpdate.LabelAsAnOrphan()
	now := v1.Now()
	if err = r.updateLabelsAndAnnotations(ctx, prToUpdate); err == nil {
		condition := v1alpha3.Condition{
			Type:               v1alpha3.ConditionSucceeded,
			Status:             v1alpha3.ConditionUnknown,
			Reason:             "SKIPPED",
			Message:            "skipped to reconcile this PipelineRun due to not found Pipeline reference in PipelineRun.",
			LastTransitionTime: now,
			LastProbeTime:      now,
		}
		prToUpdate.Status.AddCondition(&condition)
		prToUpdate.Status.Phase = v1alpha3.Unknown
		err = r.updateStatus(ctx, &pr.Status, client.ObjectKey{Namespace: pr.Namespace, Name: pr.Name})
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// the name should obey Kubernetes naming convention: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
	r.recorder = mgr.GetEventRecorderFor("pipelinerun-controller")
	r.log = ctrl.Log.WithName("pipelinerun-controller")
	return ctrl.NewControllerManagedBy(mgr).
		Named("jenkins_pipelinerun_controller").
		For(&v1alpha3.PipelineRun{}).
		Complete(r)
}
