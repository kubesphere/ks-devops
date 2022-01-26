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

package v1alpha3

import apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

// PipelineTemplate contains template referent and parameters to be used to render the execution definition.
type PipelineTemplate struct {

	// Ref is template referent.
	Ref TemplateRef `json:"ref"`

	// Parameters are to be used to render the execution definition, like Jenkinsfile.
	//+optional
	Parameters []TemplateParameter `json:"parameters,omitempty"`
}

// TemplateRef is a referent to a template.
type TemplateRef struct {

	// Name of the template referent.
	Name string `json:"name"`

	// Kind of the template referent. Default is Template
	//+optional
	Kind string `json:"kind,omitempty"`
}

// TemplateParameter is parameter corresponding with parameter definitions in template.
type TemplateParameter struct {

	// Name of template parameter.
	Name string `json:"name"`

	// Value of template parameter.
	// +optional
	Value apiextensionv1.JSON `json:"value,omitempty"`
}
