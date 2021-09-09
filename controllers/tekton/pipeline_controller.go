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

package tekton

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

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	TknClientset *tknclient.Clientset
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PipelineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	// First, we get the Pipeline resource by its name in the request namespace.
	pipeline := &devopsv2alpha1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
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
			controllerutil.AddFinalizer(pipeline, pipelineFinalizerName)
			if err := r.Update(ctx, pipeline); err != nil {
				klog.Errorf("unable to add finalizer to pipeline [%s]", pipeline.Spec.Name)
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
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv2alpha1.Pipeline{}).
		Complete(r)
}

// deleteExternalResources deletes any external resources associated with the devopsv2alpha1.Pipeline
func (r *PipelineReconciler) deleteExternalResources(ctx context.Context, pipeline *devopsv2alpha1.Pipeline) error {
	klog.Infof("Pipeline [%s] is under deletion.", pipeline.Name)

	// Firstly, we are to find and delete all the related Tekton Tasks.
	var tknTaskName string
	for _, taskSpec := range pipeline.Spec.Tasks {
		tknTaskName = pipeline.Name + "-" + taskSpec.Name
		if _, err := r.TknClientset.TektonV1beta1().
			Tasks(pipeline.Namespace).
			Get(ctx, tknTaskName, metav1.GetOptions{}); err != nil {
			klog.Infof("unable to find Tekton Task [%s] in namespace %s", tknTaskName, pipeline.Namespace)
			// Tekton Task resource does not exist, which means we should do nothing here.
			continue
		}

		// Since Tekton Task exists, we need to delete it.
		if err := r.TknClientset.TektonV1beta1().
			Tasks(pipeline.Namespace).
			Delete(ctx, tknTaskName, metav1.DeleteOptions{}); err != nil {
			klog.Errorf("unable to delete Tekton Task [%s] using tekton client", tknTaskName)
			return err
		}
	}

	// Secondly, we should find and delete all the related Tekton Pipelines.
	tknPipelineName := pipeline.Spec.Name
	if _, err := r.TknClientset.TektonV1beta1().
		Pipelines(pipeline.Namespace).
		Get(ctx, tknPipelineName, metav1.GetOptions{}); err != nil {
		// Tekton Pipeline resource does not exist and that means we ought to do nothing here.
		klog.V(5).Infof("unable to find Tekton Pipeline [%s] in namespace %s", tknPipelineName, pipeline.Namespace)
		return nil
	}

	// If there exists related Tekton Pipelines, we will delete it.
	if err := r.TknClientset.TektonV1beta1().
		Pipelines(pipeline.Namespace).
		Delete(ctx, tknPipelineName, metav1.DeleteOptions{}); err != nil {
		// When we failed to delete Tekton PipelineRun resource, then we should return an error here.
		klog.Errorf("unable to delete Tekton Pipeline [%s] using tekton client", tknPipelineName)
		return err
	}

	// What's more, there is no need to hold related PipelineRun CRD resources if
	// the Pipeline CRD is deleted.
	// As a result, we need to delete all the devops and Tekton PipelineRun CRD resources,
	// whose pipelineRef is the deleting pipeline, when deleting a Pipeline.

	// 1. Clean devops PipelineRun CRD resources
	devopsPipelineRunList := devopsv2alpha1.PipelineRunList{}
	if err := r.List(ctx, &devopsPipelineRunList); err != nil {
		klog.Error(err, "unable to fetch pipeline crd resources")
		return err
	}
	// delete the relevant devops PipelineRun resources
	for _, devopsPipelineRun := range devopsPipelineRunList.Items {
		if devopsPipelineRun.Spec.PipelineRef != pipeline.Spec.Name {
			continue
		}
		if err := r.Delete(ctx, &devopsPipelineRun); err != nil {
			klog.Errorf("unable to delete devops PipelineRun [%s]", devopsPipelineRun.Name)
			return err
		}
	}

	// 2. Clean Tekton PipelineRun CRD resources
	var err error
	var tknPipelineRunList *tektonv1.PipelineRunList
	if tknPipelineRunList, err = r.TknClientset.TektonV1beta1().PipelineRuns(pipeline.Namespace).List(ctx, metav1.ListOptions{}); err != nil {
		// If we fail to list any Tekton PipelineRun resource, we will return an error here.
		klog.Error("unable to list Tekton PipelineRun resources")
		return err
	}
	for _, tknPipelineRun := range tknPipelineRunList.Items {
		// filter unrelated Tekton PipelineRun CRD resources
		if tknPipelineRun.Spec.PipelineRef.Name != tknPipelineName {
			continue
		}
		// delete target Tekton PipelineRun CRD resources
		if err := r.TknClientset.TektonV1beta1().
			PipelineRuns(pipeline.Namespace).
			Delete(ctx, tknPipelineRun.Name, metav1.DeleteOptions{}); err != nil {
			// When we fail to delete tekton pipelinerun, we will return an error.
			klog.Errorf("unable to delete pipelinerun [%s]", tknPipelineRun.Name)
			return err
		}
	}

	klog.Infof("Pipeline [%s] and its related Tekton resources were deleted successfully.", pipeline.Name)

	return nil
}

// reconcileTektonCrd transforms our Pipeline CRD to Tekton CRDs
func (r *PipelineReconciler) reconcileTektonCrd(ctx context.Context, namespace string, pipeline *devopsv2alpha1.Pipeline) error {
	klog.Infof("Devops pipeline name: %s\ttask num: %d", pipeline.Name, len(pipeline.Spec.Tasks))

	// transform tasks included in the pipeline to Tekton Tasks
	for _, task := range pipeline.Spec.Tasks {
		if err := r.reconcileTektonTask(ctx, namespace, &task, pipeline.Name); err != nil {
			klog.Error(err, "unable to reconcile Tekton task resources")
			return err
		}
	}

	if err := r.reconcileTektonPipeline(ctx, namespace, pipeline); err != nil {
		klog.Error(err, "unable to reconcile Tekton pipeline resources")
		return err
	}

	return nil
}

// reconcileTektonTask transforms tasks in our Pipeline CRD to Tekton Task CRD
func (r *PipelineReconciler) reconcileTektonTask(ctx context.Context, namespace string, taskSpec *devopsv2alpha1.TaskSpec, pipelineName string) error {
	task := &tektonv1.Task{}
	// Notes: in order to differentiate tasks from pipelines, we add pipeline name before the task name and concatenate them with a dash character
	if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: pipelineName + "-" + taskSpec.Name}, task); err != nil {
		// This means that we do not find the Tekton Task in the given namespace,
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

		klog.Infof("Tekton Task %s was created successfully.", tektonTask.ObjectMeta.Name)
	} else {
		klog.Infof("Tekton Task %s already exists.", task.ObjectMeta.Name)
	}

	return nil
}

// reconcileTektonPipeline transforms our Pipeline CRD to Tekton Pipeline CRD
func (r *PipelineReconciler) reconcileTektonPipeline(ctx context.Context, namespace string, pipeline *devopsv2alpha1.Pipeline) error {
	tPipeline := &tektonv1.Pipeline{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: pipeline.Name}, tPipeline); err != nil {
		// This means the current Tekton Pipeline does not exist in the given namespace,
		// i.e. we are supposed to create it.

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
			klog.Error("unable to create tekton pipeline")
			return err
		}

		klog.Infof("Tekton Pipeline resource %s was created successfully.", tektonPipeline.ObjectMeta.Name)
	} else {
		klog.Infof("Tekton Pipeline resource %s already exists.", tPipeline.ObjectMeta.Name)
	}

	return nil
}
