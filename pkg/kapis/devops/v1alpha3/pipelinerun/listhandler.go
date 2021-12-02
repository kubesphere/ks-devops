package pipelinerun

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	resourcesV1alpha3 "kubesphere.io/devops/pkg/models/resources/v1alpha3"
)

// listHandler is default implementation for PipelineRun.
type listHandler struct {
}

// Make sure backwardListHandler implement ListHandler interface.
var _ resourcesV1alpha3.ListHandler = listHandler{}

func (b listHandler) Comparator() resourcesV1alpha3.CompareFunc {
	return func(left, right runtime.Object, f query.Field) bool {
		leftPipelineRun, ok := left.(*v1alpha3.PipelineRun)
		if !ok {
			return false
		}
		rightPipelineRun, ok := right.(*v1alpha3.PipelineRun)
		if !ok {
			return false
		}
		// Compare start time and creation time(if missing former)
		leftTime := leftPipelineRun.Status.StartTime
		if leftTime.IsZero() {
			leftTime = &leftPipelineRun.CreationTimestamp
		}
		rightTime := rightPipelineRun.Status.StartTime
		if rightTime.IsZero() {
			rightTime = &rightPipelineRun.CreationTimestamp
		}
		if !leftTime.Equal(rightTime) {
			return leftTime.After(rightTime.Time)
		}
		// Why compare name at last?
		// We have to ensure that the order is stable when left time is equal to right time.
		// Exactly, the name is unique globally.
		return strings.Compare(leftPipelineRun.Name, rightPipelineRun.Name) < 0
	}
}

func (b listHandler) Filter() resourcesV1alpha3.FilterFunc {
	return resourcesV1alpha3.DefaultFilter()
}

func (b listHandler) Transformer() resourcesV1alpha3.TransformFunc {
	return resourcesV1alpha3.NoTransformFunc()
}
