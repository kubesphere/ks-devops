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
	"fmt"

	"github.com/kubesphere/ks-devops/pkg/utils/sliceutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddFinalizer adds an finalizer
func AddFinalizer(objectMeta *metav1.ObjectMeta, finalizer string) (added bool) {
	count := len(objectMeta.Finalizers)
	objectMeta.Finalizers = sliceutil.AddToSlice(finalizer, objectMeta.Finalizers)
	added = len(objectMeta.Finalizers) > count
	return
}

// RemoveFinalizer removes an finalizer
func RemoveFinalizer(objectMeta *metav1.ObjectMeta, finalizer string) {
	objectMeta.Finalizers = sliceutil.RemoveString(objectMeta.Finalizers, sliceutil.SameItem(finalizer))
}

func ExtractObjectMeta(obj runtime.Object) (*metav1.ObjectMeta, error) {
	// Ensure the object supports the metav1.Object interface
	accessor, ok := obj.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("object does not implement metav1.Object")
	}

	// Create a copy of the ObjectMeta
	meta := &metav1.ObjectMeta{
		Name:                       accessor.GetName(),
		GenerateName:               accessor.GetGenerateName(),
		Namespace:                  accessor.GetNamespace(),
		SelfLink:                   accessor.GetSelfLink(),
		UID:                        accessor.GetUID(),
		ResourceVersion:            accessor.GetResourceVersion(),
		Generation:                 accessor.GetGeneration(),
		CreationTimestamp:          accessor.GetCreationTimestamp(),
		DeletionTimestamp:          accessor.GetDeletionTimestamp(),
		DeletionGracePeriodSeconds: accessor.GetDeletionGracePeriodSeconds(),
		Labels:                     accessor.GetLabels(),
		Annotations:                accessor.GetAnnotations(),
		OwnerReferences:            accessor.GetOwnerReferences(),
		Finalizers:                 accessor.GetFinalizers(),
		ManagedFields:              accessor.GetManagedFields(),
	}
	return meta, nil
}
