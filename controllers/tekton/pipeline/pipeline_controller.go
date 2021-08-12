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

package pipeline

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv2alpha1 "kubesphere.io/devops/pkg/api/devops/v2alpha1"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PipelineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("pipelinerun", req.NamespacedName)

	// get cr by its name
	pipeline := &devopsv2alpha1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		r.Log.Error(err, "unable to fetch pipeline crd resources")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// translate our pipeline CRD to tekton CRDs, e.g. Task, Pipeline, PipelineResource
	if err := r.reconcileTektonCrd(ctx, pipeline); err != nil {
		r.Log.Error(err, "unable to reconcile tekton crd resources")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv2alpha1.Pipeline{}).
		Complete(r)
}

// reconcileTektonCrd translates our Pipeline CRD to Tekton CRDs
func (r *PipelineReconciler) reconcileTektonCrd(ctx context.Context, pipeline *devopsv2alpha1.Pipeline) error {
	// print the pipeline name and the number of its tasks.
	r.Log.Info(fmt.Sprintf("Pipeline name: %s\tTask num: %d", pipeline.Name, len(pipeline.Spec.Tasks)))

	// translate tasks included in the pipeline to Tekton Tasks
	r.Log.Info("Going to translate tasks included in the pipeline to Tekton Tasks.")
	for _, task := range pipeline.Spec.Tasks {
		if err := r.reconcileTektonTask(ctx, &task, pipeline.Name); err != nil {
			r.Log.Error(err, "Failed to reconcile tekton task resources.")
			return err
		}
	}

	r.Log.Info("Going to translate Pipeline CRD to Tekton Pipeline CRD.")
	if err := r.reconcileTektonPipeline(ctx, pipeline); err != nil {
		r.Log.Error(err, "Failed to reconcile tekton pipeline resources.")
		return err
	}

	return nil
}

// reconcileTektonTask translates tasks in our Pipeline CRD to Tekton Task CRD
func (r *PipelineReconciler) reconcileTektonTask(ctx context.Context, task *devopsv2alpha1.TaskSpec, pipelineName string) error {
	// print the task name
	r.Log.Info(fmt.Sprintf("Translating task %s to Tekton Task.", task.Name))

	// check if the task already exists
	// in order to differentiate tasks from pipelines, we add pipeline name before the task name
	tTask := &tektonv1.Task{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: "default", Name: pipelineName + "-" + task.Name}, tTask); err != nil {
		// this means that we do not find the Tekton Task in the given namespace
		// i.e. we need to create it.

		// construct steps first
		steps := make([]tektonv1.Step, len(task.Steps))
		for i, step := range task.Steps {
			steps[i] = tektonv1.Step{
				Container: corev1.Container{
					Name:    step.Name,
					Image:   step.Image,
					Command: step.Command,
					Args:    step.Args,
				},
				Script: step.Script,
			}
		}

		// add steps in Tekton Task
		tektonTask := &tektonv1.Task{
			TypeMeta:   metav1.TypeMeta{Kind: "Task", APIVersion: "tekton.dev/v1beta1"},
			ObjectMeta: metav1.ObjectMeta{Name: pipelineName + "-" + task.Name, Namespace: "default"},
			Spec: tektonv1.TaskSpec{
				Steps: steps,
			},
		}

		// create Tekton Task
		if err := r.Create(ctx, tektonTask); err != nil {
			return err
		}

		// if create tekton task successfully, log it.
		r.Log.Info(fmt.Sprintf("Tekton Task %s created successfully.", tektonTask.ObjectMeta.Name))
	} else {
		r.Log.Info(fmt.Sprintf("Tekton Task %s already exists.", tTask.ObjectMeta.Name))
	}

	return nil
}

// reconcileTektonPipeline translates our Pipeline CRD to Tekton Pipeline CRD
func (r *PipelineReconciler) reconcileTektonPipeline(ctx context.Context, pipeline *devopsv2alpha1.Pipeline) error {
	// print the pipeline name
	r.Log.Info(fmt.Sprintf("Translating Pipeline %s to Tekton Pipeline.", pipeline.Name))

	// check if the task already exists
	tPipeline := &tektonv1.Pipeline{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: "default", Name: pipeline.Name}, tPipeline); err != nil {
		// this means the current Tekton Pipeline does not exist in the given namespace
		// i.e. we are supposed to create it

		// set up tekton tasks
		tasks := make([]tektonv1.PipelineTask, len(pipeline.Spec.Tasks))
		for i, task := range pipeline.Spec.Tasks {
			tasks[i] = tektonv1.PipelineTask{
				Name: pipeline.Name + "-" + task.Name,
				TaskRef: &tektonv1.TaskRef{
					Name: pipeline.Name + "-" + task.Name,
				},
			}
		}

		// set up tekton pipeline
		tektonPipeline := &tektonv1.Pipeline{
			TypeMeta:   metav1.TypeMeta{Kind: "Pipeline", APIVersion: "tekton.dev/v1beta1"},
			ObjectMeta: metav1.ObjectMeta{Name: pipeline.Name, Namespace: "default"},
			Spec:       tektonv1.PipelineSpec{Tasks: tasks},
		}

		// create tekton pipeline
		if err := r.Create(ctx, tektonPipeline); err != nil {
			return err
		}

		// if create tekton pipeline successfully, log it.
		r.Log.Info(fmt.Sprintf("Tekton Pipeline resource %s created successfully.", tektonPipeline.ObjectMeta.Name))
	} else {
		r.Log.Info(fmt.Sprintf("Tekton Pipeline resource %s already exists.", tPipeline.ObjectMeta.Name))
	}

	return nil
}
