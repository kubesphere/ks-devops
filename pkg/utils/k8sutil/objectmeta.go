package k8sutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/utils/sliceutil"
)

// AddFinalizer adds an finalizer
func AddFinalizer(objectMeta *metav1.ObjectMeta, finalizer string) {
	objectMeta.Finalizers = sliceutil.AddToSlice(finalizer, objectMeta.Finalizers)
}

// RemoveFinalizer removes an finalizer
func RemoveFinalizer(objectMeta *metav1.ObjectMeta, finalizer string) {
	objectMeta.Finalizers = sliceutil.RemoveString(objectMeta.Finalizers, sliceutil.SameItem(finalizer))
}
