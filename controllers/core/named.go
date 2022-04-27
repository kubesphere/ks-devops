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

package core

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NamedReconciler is an interface with allow a reconciler has a name
type NamedReconciler interface {
	reconcile.Reconciler
	// GetName returns the name of the reconciler
	GetName() string
}

// GroupReconciler is an interface with the group name
// a group means a set of some reconcilers that have a particular feature
type GroupReconciler interface {
	reconcile.Reconciler
	// GetGroupName returns the group name
	GetGroupName() string
}

// GroupedReconciler is an interface for grouping reconciler purpose
type GroupedReconciler interface {
	NamedReconciler
	GroupReconciler

	SetupWithManager(mgr ctrl.Manager) error
}

// GroupedReconcilers is an alias of the slice of GroupedReconciler
type GroupedReconcilers []GroupedReconciler

// Append is a similar function to the original slice append
func (g GroupedReconcilers) Append(reconciler GroupedReconciler) GroupedReconcilers {
	g = append(g, reconciler)
	return g
}

// Size returns the size of the slice
func (g GroupedReconcilers) Size() int {
	return len(g)
}

// GetName returns the name of the group
func (g GroupedReconcilers) GetName() (name string) {
	if len(g) > 0 {
		name = g[0].GetGroupName()
	}
	return
}

// SetupWithManager setups with a group of reconcilers
func (g GroupedReconcilers) SetupWithManager(mgr manager.Manager) (err error) {
	for i := range g {
		item := g[i]
		if err = item.SetupWithManager(mgr); err != nil {
			break
		}
	}
	return
}
