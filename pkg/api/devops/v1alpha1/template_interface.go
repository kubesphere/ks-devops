package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemplateObject interface {
	v1.Object
	runtime.Object
	// TemplateSpec returns TemplateSpec.
	TemplateSpec() TemplateSpec
}
