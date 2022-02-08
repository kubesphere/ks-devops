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

package k8sutil

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type demoCR struct {
	metav1.ObjectMeta
}

func TestAddFinalizer(t *testing.T) {
	demo := &demoCR{}

	type args struct {
		objectMeta *metav1.ObjectMeta
		finalizer  string
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, meta *metav1.ObjectMeta)
	}{{
		name: "normal case",
		args: args{
			objectMeta: &demo.ObjectMeta,
			finalizer:  "abc",
		},
		verify: func(t *testing.T, meta *metav1.ObjectMeta) {
			assert.ElementsMatch(t, []string{"abc"}, meta.Finalizers)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddFinalizer(tt.args.objectMeta, tt.args.finalizer)
			tt.verify(t, tt.args.objectMeta)
		})
	}
}

func TestRemoveFinalizer(t *testing.T) {
	demo := &demoCR{
		ObjectMeta: metav1.ObjectMeta{
			Finalizers: []string{"abc", "def"},
		}}

	type args struct {
		objectMeta *metav1.ObjectMeta
		finalizer  string
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, meta *metav1.ObjectMeta)
	}{{
		name: "normal case",
		args: args{
			objectMeta: &demo.ObjectMeta,
			finalizer:  "def",
		},
		verify: func(t *testing.T, meta *metav1.ObjectMeta) {
			assert.ElementsMatch(t, []string{"abc"}, meta.Finalizers)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RemoveFinalizer(tt.args.objectMeta, tt.args.finalizer)
			tt.verify(t, tt.args.objectMeta)
		})
	}
}
