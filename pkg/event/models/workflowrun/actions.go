package workflowrun

import (
	"fmt"
)

// Actions are set of action carried by WorkflowRun.
type Actions []map[string]interface{}

// GetAction gets action from actions in WorkflowRun by full qualified class name.
// Return nil map if not found.
func (actions Actions) GetAction(class string) map[string]interface{} {
	if class == "" {
		return nil
	}
	for _, action := range actions {
		if action["_class"] == class {
			return action
		}
	}
	return nil
}

// Parameter is an action item storing all parameters passed to WorkflowRun.
type Parameter struct {
	// Name indicates that name of the parameter.
	Name string `json:"name" description:"parameter name"`

	// Value indicates that value of the parameter.
	Value string `json:"value" description:"parameter value"`
}

// GetParameters gets a set of parameters passed to WorkflowRun.
func (actions Actions) GetParameters() []Parameter {
	action := actions.GetAction("hudson.model.ParametersAction")

	_parameters := action["parameters"]
	parameters, ok := _parameters.([]interface{})
	if !ok {
		return nil
	}
	var paramsResult []Parameter
	for _, _parameter := range parameters {
		parameter, ok := _parameter.(map[string]interface{})
		if !ok {
			continue
		}
		_name := parameter["name"]
		_value := parameter["value"]
		if _name == nil {
			continue
		}
		name := fmt.Sprint(_name)
		if name == "" {
			// should never happen
			continue
		}

		value := fmt.Sprint(_value)
		if _value == nil {
			value = ""
		}
		paramsResult = append(paramsResult, Parameter{
			Name:  fmt.Sprint(name),
			Value: fmt.Sprint(value),
		})
	}
	return paramsResult
}
