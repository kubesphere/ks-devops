package indexers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

// CreatePipelineRunSCMRefNameIndexer creates field indexer which could speed up listing PipelineRun by SCM reference name.
func CreatePipelineRunSCMRefNameIndexer(runtimeCache cache.Cache) error {
	return runtimeCache.IndexField(context.Background(),
		&v1alpha3.PipelineRun{},
		v1alpha3.PipelineRunSCMRefNameField,
		func(o runtime.Object) []string {
			pipelineRun, ok := o.(*v1alpha3.PipelineRun)
			if !ok || pipelineRun == nil {
				return []string{}
			}
			if pipelineRun.Spec.SCM == nil {
				return []string{}
			}
			return []string{pipelineRun.Spec.SCM.RefName}
		})
}
