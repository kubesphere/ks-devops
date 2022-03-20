// Copyright 2022 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package template

import (
	"bytes"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/api/devops"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	tmpl "text/template"
)

const parametersKey = "params"

// Parameter is a pair of name and value.
type Parameter struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func render(templateObject v1alpha3.TemplateObject, parameters []Parameter) (v1alpha3.TemplateObject, error) {
	templateObject = templateObject.DeepCopyObject().(v1alpha3.TemplateObject)
	rawTemplate := templateObject.TemplateSpec().Template
	templateName := types.NamespacedName{
		Name:      templateObject.GetName(),
		Namespace: templateObject.GetNamespace(),
	}.String()
	template := tmpl.New(templateName)
	//TODO Make delimiters configurable
	template.Delims("$(", ")")
	if _, err := template.Parse(rawTemplate); err != nil {
		klog.Errorf("failed to parse template: %s, and err = %v", templateName, err)
		return nil, errors.NewBadRequest("Failed to render template, please check the pipeline template for syntax error.")
	}

	// TODO Verify required parameters and default parameters
	// check the required parameters
	parameterMap := map[string]interface{}{}
	for _, parameter := range parameters {
		parameterMap[parameter.Name] = parameter.Value
	}

	parametersData := map[string]map[string]interface{}{}
	parametersData[parametersKey] = parameterMap

	buffer := &bytes.Buffer{}
	if err := template.Execute(buffer, parametersData); err != nil {
		klog.Errorf("failed to apply a parsed template(%s) to the specified data object(%v), and err = %v ", templateName, parametersData, err)
		return nil, errors.NewBadRequest("Failed to render template, please check the parameters you provided.")
	}
	renderResult := buffer.String()

	if templateObject.GetAnnotations() == nil {
		templateObject.SetAnnotations(map[string]string{})
	}
	templateObject.GetAnnotations()[devops.GroupName+devops.RenderResultAnnoKey] = renderResult
	return templateObject, nil
}
