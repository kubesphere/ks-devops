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

package workflowrun

import "encoding/json"

// Parameter is an action item storing all parameters passed to WorkflowRun.
type Parameter struct {
	// Kind represents what is the type of this parameter.
	Kind string `json:"_class"`

	// Name indicates that name of the parameter.
	Name string `json:"name" description:"parameter name"`

	// Value indicates that value of the parameter.
	Value interface{} `json:"value" description:"parameter value"`
}

// ParameterAction is an action which carries parameters.
type ParameterAction struct {
	Parameters []Parameter `json:"parameters"`
}

// Kind returns kind of parameter action.
func (*ParameterAction) Kind() string {
	return "hudson.model.ParametersAction"
}

// GetParameters gets a set of parameters passed to WorkflowRun.
func (actions *Actions) GetParameters() ([]Parameter, error) {
	var parameterAction *ParameterAction
	action := actions.GetAction(parameterAction.Kind())
	if action == nil {
		// If not parameter action found
		return nil, nil
	}
	parameterAction = &ParameterAction{}
	if err := json.Unmarshal(action.Raw, parameterAction); err != nil {
		return nil, err
	}
	return parameterAction.Parameters, nil
}
