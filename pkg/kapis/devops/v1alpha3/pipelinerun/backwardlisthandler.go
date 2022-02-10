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

package pipelinerun

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"kubesphere.io/devops/pkg/apiserver/query"
	resourcesV1alpha3 "kubesphere.io/devops/pkg/models/resources/v1alpha3"
)

type backwardListHandler struct {
}

// Make sure backwardListHandler implement ListHandler interface.
var _ resourcesV1alpha3.ListHandler = backwardListHandler{}

func (b backwardListHandler) Comparator() resourcesV1alpha3.CompareFunc {
	return resourcesV1alpha3.DefaultCompare()
}

func (b backwardListHandler) Filter() resourcesV1alpha3.FilterFunc {
	return resourcesV1alpha3.DefaultFilter().And(func(object runtime.Object, filter query.Filter) bool {
		return b.backwardFilter(object)
	})
}

func (b backwardListHandler) Transformer() resourcesV1alpha3.TransformFunc {
	return func(object runtime.Object) interface{} {
		return b.backwardTransformer(object)
	}
}

func checkPipelineRun(object runtime.Object) (*v1alpha3.PipelineRun, bool) {
	pr, ok := object.(*v1alpha3.PipelineRun)
	if !ok || pr == nil {
		return nil, false
	}
	return pr, true
}

func (b backwardListHandler) backwardFilter(object runtime.Object) bool {
	if pr, valid := checkPipelineRun(object); valid {
		return pr.Annotations[v1alpha3.JenkinsPipelineRunStatusAnnoKey] != ""
	}
	return false
}

func (b backwardListHandler) backwardTransformer(object runtime.Object) json.Marshaler {
	pr, valid := checkPipelineRun(object)
	if !valid {
		// should never happen
		return json.RawMessage("{}")
	}
	runStatusJSON := pr.Annotations[v1alpha3.JenkinsPipelineRunStatusAnnoKey]
	rawRunStatus := json.RawMessage(runStatusJSON)
	// check if the run status is a valid JSON
	valid = json.Valid(rawRunStatus)
	if !valid {
		klog.ErrorS(nil, "invalid Jenkins run status",
			"PipelineRun", fmt.Sprintf("%s/%s", pr.GetNamespace(), pr.GetName()), "runStatusJSON", runStatusJSON)
		rawRunStatus = []byte("{}")
	}
	return rawRunStatus
}
