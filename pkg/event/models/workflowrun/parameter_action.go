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
