package jenkinsconfig

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestSetContainersLimit(t *testing.T) {
	goContainer := map[interface{}]interface{}{
		"name":                  "go",
		"resourceRequestCpu":    "100m",
		"resourceRequestMemory": "100Mi",
		"resourceLimitCpu":      "4000m",
		"resourceLimitMemory":   "8192Mi",
	}

	jnlpContainer := map[interface{}]interface{}{
		"name":                  "jnlp",
		"resourceRequestCpu":    "50m",
		"resourceRequestMemory": "400Mi",
		"resourceLimitCpu":      "500m",
		"resourceLimitMemory":   "1536Mi",
	}

	cloneMapString := func(source map[string]string) map[string]string {
		target := make(map[string]string)
		for key, value := range source {
			target[key] = value
		}
		return target
	}
	cloneMapInterface := func(source map[interface{}]interface{}) map[interface{}]interface{} {
		target := make(map[interface{}]interface{})
		for key, value := range source {
			target[key] = value
		}
		return target
	}

	type TestConfig struct {
		name           string
		containers     []interface{}
		providedConfig map[string]string
		assertion      func(config *TestConfig)
	}

	tests := []TestConfig{
		{
			name: "All settings",
			containers: []interface{}{
				cloneMapInterface(goContainer),
				cloneMapInterface(jnlpContainer),
			},
			providedConfig: map[string]string{
				jnlpLimitCPUKey:    "100m",
				jnlpLimitMemoryKey: "2000Mi",
				goLimitCPUKey:      "1000m",
				goLimitMemoryKey:   "4096Mi",
			},
			assertion: func(config *TestConfig) {
				assert.Equal(t, config.providedConfig[jnlpLimitCPUKey], config.containers[1].(map[interface{}]interface{})["resourceLimitCpu"])
				assert.Equal(t, config.providedConfig[jnlpLimitMemoryKey], config.containers[1].(map[interface{}]interface{})["resourceLimitMemory"])
				assert.Equal(t, "50m", config.containers[1].(map[interface{}]interface{})["resourceRequestCpu"])
				assert.Equal(t, "400Mi", config.containers[1].(map[interface{}]interface{})["resourceRequestMemory"])

				assert.Equal(t, config.providedConfig[goLimitCPUKey], config.containers[0].(map[interface{}]interface{})["resourceLimitCpu"])
				assert.Equal(t, config.providedConfig[goLimitMemoryKey], config.containers[0].(map[interface{}]interface{})["resourceLimitMemory"])
				assert.Equal(t, "100m", config.containers[0].(map[interface{}]interface{})["resourceRequestCpu"])
				assert.Equal(t, "100Mi", config.containers[0].(map[interface{}]interface{})["resourceRequestMemory"])
			},
		},
		{
			name: "Partial settings",
			containers: []interface{}{
				cloneMapInterface(goContainer),
				cloneMapInterface(jnlpContainer),
			},
			providedConfig: map[string]string{
				jnlpLimitCPUKey: "100m",
				goLimitCPUKey:   "",
			},
			assertion: func(config *TestConfig) {
				assert.Equal(t, config.providedConfig[jnlpLimitCPUKey], config.containers[1].(map[interface{}]interface{})["resourceLimitCpu"])
				assert.Equal(t, "1536Mi", config.containers[1].(map[interface{}]interface{})["resourceLimitMemory"])
				assert.Equal(t, config.providedConfig[goLimitCPUKey], config.containers[0].(map[interface{}]interface{})["resourceLimitCpu"])
				assert.Equal(t, "8192Mi", config.containers[0].(map[interface{}]interface{})["resourceLimitMemory"])
			},
		},
		{
			name: "Empty provided config",
			containers: []interface{}{
				cloneMapInterface(goContainer),
				cloneMapInterface(jnlpContainer),
			},
			providedConfig: map[string]string{},
			assertion: func(config *TestConfig) {
				assert.True(t, reflect.DeepEqual(config.containers, []interface{}{
					cloneMapInterface(goContainer),
					cloneMapInterface(jnlpContainer),
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setContainersLimit(cloneMapString(tt.providedConfig), tt.containers, "go")
			tt.assertion(&tt)
		})
	}
}
