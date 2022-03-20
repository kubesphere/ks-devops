package argocd

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// StructToMap converts struct value to map value, which key type is string and value type is interface{}.
// The type of value must be struct. Any other types will lead an error.
func StructToMap(value interface{}) (map[string]interface{}, error) {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	uValue := map[string]interface{}{}
	if err := json.Unmarshal(valueBytes, &uValue); err != nil {
		return nil, err
	}
	return uValue, nil
}

// SetNestedField sets nested field into the object with map type. The type of field value must be struct. Any other
// types will lead to an error.
func SetNestedField(obj map[string]interface{}, value interface{}, fields ...string) error {
	// convert value to unstructured
	mapValue, err := StructToMap(value)
	if err != nil {
		return err
	}
	return unstructured.SetNestedField(obj, mapValue, fields...)
}
