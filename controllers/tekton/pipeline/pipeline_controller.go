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

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tknclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv2alpha1 "kubesphere.io/devops/pkg/api/devops/v2alpha1"
)

// Reconciler reconciles a Pipeline object
type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	TknClient *tknclient.Clientset
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
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	// for debug
	// klog.Infof("req: %v", req)

	// First, we get the Pipeline resource by its name in the request namespace
	pipeline := &devopsv2alpha1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		klog.Error(err, "unable to fetch pipeline crd resources")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Second, we check whether the pipeline object is being deleted,
	// by examining DeletionTimestamp field in it.
	pipelineFinalizerName := devopsv2alpha1.PipelineFinalizerName
	if pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(pipeline.GetFinalizers(), pipelineFinalizerName) {
			klog.Infof("Add finalizer to %s", pipeline.Name)
			controllerutil.AddFinalizer(pipeline, pipelineFinalizerName)
			if err := r.Update(ctx, pipeline); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(pipeline.GetFinalizers(), pipelineFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, pipeline); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(pipeline, pipelineFinalizerName)
			if err := r.Update(ctx, pipeline); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Below is our reconciling core logic.
	// What we do is transforming our pipeline CRD to tekton CRDs,
	// e.g. Task and Pipeline.
	if err := r.reconcileTektonCrd(ctx, req.Namespace, pipeline); err != nil {
		klog.Error(err, "unable to reconcile tekton crd resources")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv2alpha1.Pipeline{}).
		Complete(r)
}

// deleteExternalResources deletes any external resources associated with the devopsv2alpha1.Pipeline
func (r *Reconciler) deleteExternalResources(ctx context.Context, pipeline *devopsv2alpha1.Pipeline) error {
	klog.Infof("Pipeline [%s] is under deletion.", pipeline.Name)

	// Firstly, we are to find and delete all the related Tekton Tasks
	var tknTaskName string
	for _, taskSpec := range pipeline.Spec.Tasks {
		tknTaskName = pipeline.Name + "-" + taskSpec.Name
		if _, err := r.TknClient.TektonV1beta1().
			Tasks(pipeline.Namespace).
			Get(ctx, tknTaskName, metav1.GetOptions{}); err != nil {
			klog.Infof("Tekton Task [%s] was not found in namespace %s.", tknTaskName, pipeline.Namespace)
			// Tekton Task resource does not exist.
			// We do nothing here.
			continue
		}

		// Tekton Task exists, we need to delete it.
		if err := r.TknClient.TektonV1beta1().
			Tasks(pipeline.Namespace).
			Delete(ctx, tknTaskName, metav1.DeleteOptions{}); err != nil {
			// Failed to delete tekton pipelinerun
			klog.Errorf("Failed to delete Tekton Task [%s] using tekton client.", tknTaskName)
			return err
		}
	}

	// Secondly, we should find and delete all the related Tekton Pipelines
	tknPipelineName := pipeline.Spec.Name
	if _, err := r.TknClient.TektonV1beta1().
		Pipelines(pipeline.Namespace).
		Get(ctx, tknPipelineName, metav1.GetOptions{}); err != nil {
		klog.Infof("Tekton Pipeline [%s] was not found in namespace %s.", tknPipelineName, pipeline.Namespace)
		// Tekton Pipeline resource does not exist.
		// We do nothing here.
	} else {
		// If there exists related Tekton Pipelines,
		// we will delete it.
		if err := r.TknClient.TektonV1beta1().
			Pipelines(pipeline.Namespace).
			Delete(ctx, tknPipelineName, metav1.DeleteOptions{}); err != nil {
			// Failed to delete tekton pipelinerun
			klog.Errorf("Failed to delete Tekton Pipeline [%s] using tekton client.", tknPipelineName)
			return err
		}
	}

	klog.Infof("Pipeline [%s] and its related Tekton resources were deleted successfully.", pipeline.Name)

	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// reconcileTektonCrd transforms our Pipeline CRD to Tekton CRDs
func (r *Reconciler) reconcileTektonCrd(ctx context.Context, namespace string, pipeline *devopsv2alpha1.Pipeline) error {
	// print the pipeline name and the number of its tasks.
	klog.Infof("Pipeline name: %s\tTask num: %d", pipeline.Name, len(pipeline.Spec.Tasks))

	// transform tasks included in the pipeline to Tekton Tasks
	// klog.Info("Going to transform tasks included in the pipeline to Tekton Tasks.")
	for _, task := range pipeline.Spec.Tasks {
		if err := r.reconcileTektonTask(ctx, namespace, &task, pipeline.Name); err != nil {
			klog.Error(err, "Failed to reconcile tekton task resources.")
			return err
		}
	}

	// klog.Info("Going to transform Pipeline CRD to Tekton Pipeline CRD.")
	if err := r.reconcileTektonPipeline(ctx, namespace, pipeline); err != nil {
		klog.Error(err, "Failed to reconcile tekton pipeline resources.")
		return err
	}

	return nil
}

// reconcileTektonTask transforms tasks in our Pipeline CRD to Tekton Task CRD
func (r *Reconciler) reconcileTektonTask(ctx context.Context, namespace string, taskSpec *devopsv2alpha1.TaskSpec, pipelineName string) error {
	// print the task name
	klog.Infof("Transforming task %s to Tekton Task.", taskSpec.Name)

	// check if the task already exists
	// in order to differentiate tasks from pipelines, we add pipeline name before the task name
	task := &tektonv1.Task{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: pipelineName + "-" + taskSpec.Name}, task); err != nil {
		// this means that we do not find the Tekton Task in the given namespace
		// i.e. we need to create it.

		// construct steps first
		steps := make([]tektonv1.Step, len(taskSpec.Steps))
		for i, step := range taskSpec.Steps {
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
			ObjectMeta: metav1.ObjectMeta{Name: pipelineName + "-" + taskSpec.Name, Namespace: namespace},
			Spec: tektonv1.TaskSpec{
				Steps: steps,
			},
		}

		// create Tekton Task
		if err := r.Create(ctx, tektonTask); err != nil {
			return err
		}

		// if create tekton task successfully, log it.
		klog.Infof("Tekton Task %s was created successfully.", tektonTask.ObjectMeta.Name)
	} else {
		klog.Infof("Tekton Task %s already exists.", task.ObjectMeta.Name)
	}

	return nil
}

// reconcileTektonPipeline transforms our Pipeline CRD to Tekton Pipeline CRD
func (r *Reconciler) reconcileTektonPipeline(ctx context.Context, namespace string, pipeline *devopsv2alpha1.Pipeline) error {
	// print the pipeline name
	klog.Infof("Transforming Pipeline %s to Tekton Pipeline.", pipeline.Name)

	// check if the task already exists
	tPipeline := &tektonv1.Pipeline{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: pipeline.Name}, tPipeline); err != nil {
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
			ObjectMeta: metav1.ObjectMeta{Name: pipeline.Name, Namespace: namespace},
			Spec:       tektonv1.PipelineSpec{Tasks: tasks},
		}

		// create tekton pipeline
		if err := r.Create(ctx, tektonPipeline); err != nil {
			return err
		}

		// if create tekton pipeline successfully, log it.
		klog.Infof("Tekton Pipeline resource %s was created successfully.", tektonPipeline.ObjectMeta.Name)
	} else {
		klog.Infof("Tekton Pipeline resource %s already exists.", tPipeline.ObjectMeta.Name)
	}

	return nil
}
