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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestGroupedReconcilers_SetupWithManager(t *testing.T) {
	type args struct {
		mgr manager.Manager
	}
	tests := []struct {
		name    string
		g       GroupedReconcilers
		args    args
		wantErr bool
	}{{
		name:    "empty",
		g:       GroupedReconcilers{},
		wantErr: false,
	}, {
		name:    "no errors",
		g:       []GroupedReconciler{&FakeGroupedReconciler{}},
		wantErr: false,
	}, {
		name:    "have errors",
		g:       []GroupedReconciler{&FakeGroupedReconciler{HasError: true}},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.g.SetupWithManager(tt.args.mgr); (err != nil) != tt.wantErr {
				t.Errorf("SetupWithManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGroupedFakeReconciler(t *testing.T) {
	fakeReconciler := &FakeGroupedReconciler{}
	assert.Equal(t, "fake", fakeReconciler.GetName())
	result, err := fakeReconciler.Reconcile(context.Background(), reconcile.Request{})
	assert.True(t, result.IsZero())
	assert.Nil(t, err)

	var group GroupedReconcilers
	group = group.Append(fakeReconciler)
	assert.Equal(t, 1, group.Size())
	assert.Equal(t, "fake", group.GetName())
}
