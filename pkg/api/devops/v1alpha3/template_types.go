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

import (
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ TemplateObject = &Template{}

// TemplateSpec defines the desired state of Template
type TemplateSpec struct {
	// Parameters are used to configure template.
	//+optional
	Parameters []TemplateParameter `json:"parameters,omitempty"`

	// Template is a string with go-template style.
	Template string `json:"template,omitempty"`
}

// TemplateStatus defines the observed state of Template
type TemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// TemplateParameter is definition of how can we configure our parameter.
type TemplateParameter struct {
	// Name is name of the parameter.
	Name string `json:"name"`

	// Description is description of the parameter.
	//+optional
	Description string `json:"description,omitempty"`

	// Default is default value of the parameter.
	//+optional
	Default apiextensionv1.JSON `json:"default,omitempty"`

	// Type is type of the parameter.
	//+optional
	Type string `json:"type,omitempty"`

	// Validation is the validation configuration of the parameter, including validation expression and message.
	//+optional
	Validation *ParameterValidation `json:"validation,omitempty"`
}

// ParameterValidation is definition of how can we validate our parameter.
type ParameterValidation struct {
	// Expression is the expression of the validation.
	Expression string `json:"expression"`

	// Message is given when validation failure.
	Message string `json:"message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Template is the Schema for the templates API
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

// TemplateSpec returns specification of Template.
func (template *Template) TemplateSpec() TemplateSpec {
	return template.Spec
}

//+kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
