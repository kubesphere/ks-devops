package core

import (
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
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
	result, err := fakeReconciler.Reconcile(reconcile.Request{})
	assert.True(t, result.IsZero())
	assert.Nil(t, err)

	var group GroupedReconcilers
	group = group.Append(fakeReconciler)
	assert.Equal(t, 1, group.Size())
	assert.Equal(t, "fake", group.GetName())
}
