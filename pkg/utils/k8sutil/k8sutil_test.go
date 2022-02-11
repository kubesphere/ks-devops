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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestIsControlledBy(t *testing.T) {
	type args struct {
		ownerReferences []metav1.OwnerReference
		kind            string
		name            string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "normal case, has the reference",
		args: args{
			ownerReferences: []metav1.OwnerReference{{
				Kind: "kind",
				Name: "name",
			}},
			kind: "kind",
			name: "name",
		},
		want: true,
	}, {
		name: "not have the reference",
		args: args{
			ownerReferences: []metav1.OwnerReference{{
				Kind: "kind",
				Name: "name",
			}},
			kind: "kind",
			name: "fake",
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsControlledBy(tt.args.ownerReferences, tt.args.kind, tt.args.name), "IsControlledBy(%v, %v, %v)", tt.args.ownerReferences, tt.args.kind, tt.args.name)
		})
	}
}

func TestSetOwnerReference(t *testing.T) {
	obj := &unstructured.Unstructured{}
	obj.SetOwnerReferences([]metav1.OwnerReference{{
		Kind: "fake",
	}})

	type args struct {
		object   metav1.Object
		ownerRef metav1.OwnerReference
	}
	tests := []struct {
		name   string
		args   args
		verify func(t *testing.T, object metav1.Object)
	}{{
		name: "without owner references",
		args: args{
			object: &unstructured.Unstructured{},
			ownerRef: metav1.OwnerReference{
				Kind: "kind",
			},
		},
		verify: func(t *testing.T, object metav1.Object) {
			assert.Equal(t, 1, len(object.GetOwnerReferences()))
			assert.Equal(t, "kind", object.GetOwnerReferences()[0].Kind)
		},
	}, {
		name: "no matched owner reference",
		args: args{
			object: obj.DeepCopy(),
			ownerRef: metav1.OwnerReference{
				Kind: "kind",
			},
		},
		verify: func(t *testing.T, object metav1.Object) {
			assert.Equal(t, 2, len(object.GetOwnerReferences()))
		},
	}, {
		name: "have matched owner reference",
		args: args{
			object: obj.DeepCopy(),
			ownerRef: metav1.OwnerReference{
				Kind: "fake",
			},
		},
		verify: func(t *testing.T, object metav1.Object) {
			assert.Equal(t, 1, len(object.GetOwnerReferences()))
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetOwnerReference(tt.args.object, tt.args.ownerRef)
			tt.verify(t, tt.args.object)
		})
	}
}
