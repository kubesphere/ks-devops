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

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv2alpha1 "kubesphere.io/devops/pkg/api/devops/v2alpha1"
)

// PipelineRunReconciler reconciles a PipelineRun object
type PipelineRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelineruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PipelineRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PipelineRunReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	// get the pipelinerun crd
	pipelineRun := &devopsv2alpha1.PipelineRun{}
	if err := r.Get(ctx, req.NamespacedName, pipelineRun); err != nil {
		klog.Error(err, "unable to fetch pipeline crd")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.reconcileTektonCrd(ctx, req.Namespace, pipelineRun); err != nil {
		klog.Error(err, "Failed to reconcile Tekton PipelineRun resources.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv2alpha1.PipelineRun{}).
		Complete(r)
}

// reconcileTektonCrd translates our crd to Tekton crd
func (r *PipelineRunReconciler) reconcileTektonCrd(ctx context.Context, namespace string, pipelineRun *devopsv2alpha1.PipelineRun) error {
	if err := r.reconcileTektonPipelineRun(ctx, namespace, &pipelineRun.Spec); err != nil {
		return err
	}
	return nil
}

// reconcileTektonPipelineRun translates our PipelineRun to Tekton PipelineRun
func (r *PipelineRunReconciler) reconcileTektonPipelineRun(ctx context.Context, namespace string, pipelineRun *devopsv2alpha1.PipelineRunSpec) error {
	// print the pipelinerun name
	klog.Infof("Going to create Tekton PipelineRun resource called %s", pipelineRun.Name)

	// translate PipelineRun to Tekton PipelineRun
	tPipelineRun := &tektonv1.PipelineRun{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: pipelineRun.Name}, tPipelineRun); err != nil {
		// this means the current Tekton PipelineRun does not exist in the given namespace
		// i.e. we can safely create it.

		// set up struct of tekton pipelinerun
		tektonPipelineRun := &tektonv1.PipelineRun{
			TypeMeta:   metav1.TypeMeta{Kind: "PipelineRun", APIVersion: "tekton.dev/v1beta1"},
			ObjectMeta: metav1.ObjectMeta{Name: pipelineRun.Name, Namespace: namespace},
			Spec: tektonv1.PipelineRunSpec{
				PipelineRef: &tektonv1.PipelineRef{Name: pipelineRun.PipelineRef},
			},
		}

		// create tekton pipelinerun resource
		if err := r.Create(ctx, tektonPipelineRun); err != nil {
			return err
		}

		// log if create successfully
		klog.Infof("Tekton PipelineRun resource %s created successfully", pipelineRun.Name)
	} else {
		// This means that a Tekton PipelineRun resource has already exists in the given namespace,
		// which can be a problem.
		klog.Infof("Tekton PipelineRun resource %s already exists!", pipelineRun.Name)
	}

	return nil
}
