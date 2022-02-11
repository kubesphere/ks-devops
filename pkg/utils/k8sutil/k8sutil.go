/*
Copyright 2019 The KubeSphere Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsControlledBy returns whether the ownerReferences contains the specified resource kind
func IsControlledBy(ownerReferences []metav1.OwnerReference, kind string, name string) bool {
	for _, owner := range ownerReferences {
		if owner.Kind == kind && (name == "" || owner.Name == name) {
			return true
		}
	}
	return false
}

// SetOwnerReference adds or updates an owner reference to an object
func SetOwnerReference(object metav1.Object, ownerRef metav1.OwnerReference) {
	allRefs := object.GetOwnerReferences()
	if len(allRefs) == 0 {
		object.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	} else {
		for i, ref := range allRefs {
			if ref.Name == ownerRef.Name && ref.Kind == ownerRef.Kind &&
				ref.APIVersion == ownerRef.APIVersion {
				allRefs[i] = ownerRef
				return
			}
		}

		object.SetOwnerReferences(append(object.GetOwnerReferences(), ownerRef))
	}
}
