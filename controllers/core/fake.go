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
	"errors"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// FakeGroupedReconciler is a fake GroupedReconciler which is for the test purpose
type FakeGroupedReconciler struct {
	HasError bool
}

// GetName returns the name
func (f *FakeGroupedReconciler) GetName() string {
	return "fake"
}

// GetGroupName returns the group name
func (f *FakeGroupedReconciler) GetGroupName() string {
	return "fake"
}

// Reconcile is fake reconcile process
func (f *FakeGroupedReconciler) Reconcile(reconcile.Request) (result reconcile.Result, err error) {
	return
}

// SetupWithManager setups the reconciler
func (f *FakeGroupedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if f.HasError {
		return errors.New("fake")
	}
	return nil
}

// NoErrors represents no errors
func NoErrors(t assert.TestingT, err error, i ...interface{}) bool {
	assert.Nil(t, err)
	return true
}
