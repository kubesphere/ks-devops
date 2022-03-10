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

package addon

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=addonStrategies,verbs=get;delete;create;update
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasercontrollers,verbs=get;delete;create;update
//+kubebuilder:rbac:groups=argoproj.io,resources=argocds,verbs=get;delete;create;update

// OperatorCRDReconciler watches those CRDs which belong to Operator
type OperatorCRDReconciler struct {
	client.Client
	log logr.Logger
}

// TODO make this to be configurable
var supportedOperators = []string{"ReleaserController", "ArgoCD"}

// Reconcile manages addonStrategy according to the CRDs of operators
func (r *OperatorCRDReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()

	crd := &apiextensions.CustomResourceDefinition{}
	if err = r.Client.Get(ctx, req.NamespacedName, crd); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	spec := crd.Spec
	err = r.operatorsHandle(spec.Names.Kind, spec.Versions[0].Name)
	return
}

func (r *OperatorCRDReconciler) operatorsHandle(name string, version string) (err error) {
	ctx := context.Background()
	r.log.Info(fmt.Sprintf("start to reconcile: %s-%s", name, version))
	if !operatorSupport(name) {
		r.log.V(8).Info(fmt.Sprintf("not support %s as an operator", name))
		return
	}

	strategyName := getStrategyName(name, string(v1alpha3.AddonInstallStrategySimpleOperator))

	strategy := &v1alpha3.AddonStrategy{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name: strategyName,
	}, strategy); err != nil {
		if apierrors.IsNotFound(err) {
			// create addonStrategy
			strategy.Name = strategyName
			strategy.Spec = v1alpha3.AddStrategySpec{
				Type: v1alpha3.AddonInstallStrategySimpleOperator,
				SimpleOperator: v1.ObjectReference{
					APIVersion: version,
					Kind:       name,
				},
			}

			err = r.Client.Create(ctx, strategy)
			return
		}
	} else {
		strategy.Spec.Available = true
		strategy.Spec.SimpleOperator.APIVersion = version
		err = r.Client.Update(ctx, strategy)
	}
	return
}

func operatorSupport(name string) (support bool) {
	for _, operatorKind := range supportedOperators {
		if name == operatorKind {
			support = true
			break
		}
	}
	return
}

func getStrategyName(operatorName, kind string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s", kind, operatorName))
}

// SetupWithManager set the reconcilers
func (r *OperatorCRDReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = ctrl.Log.WithName("OperatorCRDReconciler")
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiextensions.CustomResourceDefinition{}).
		Complete(r)
}
