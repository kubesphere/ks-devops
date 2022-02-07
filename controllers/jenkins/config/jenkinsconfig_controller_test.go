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

package config

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/client/devops/fake"
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

func TestReloadJenkinsConfig(t *testing.T) {
	ctrl := Controller{}
	err := ctrl.reloadJenkinsConfig()
	assert.NotNil(t, err)

	ctrl.configOperator = &fake.Devops{}
	err = ctrl.reloadJenkinsConfig()
	assert.Nil(t, err)
}

func TestCheckJenkinsConfigData(t *testing.T) {
	ctrl := Controller{}
	err := ctrl.checkJenkinsConfigData(&v1.ConfigMap{})
	assert.Nil(t, err, "failed when check an empty ConfigMap")

	targetConfigMap := &v1.ConfigMap{
		Data: map[string]string{
			"jenkins.yaml": "fake",
		},
	}
	err = ctrl.checkJenkinsConfigData(targetConfigMap)
	assert.Nil(t, err, "failed when check a ConfigMap with the expected data field")
	assert.Equal(t, "fake", targetConfigMap.Data["jenkins_user.yaml"], "didn't get the expected data field")

	targetConfigMap = &v1.ConfigMap{
		Data: map[string]string{
			"jenkins.yaml":      "jenkins",
			"jenkins_user.yaml": "ks-jenkins",
		},
	}
	err = ctrl.checkJenkinsConfigData(targetConfigMap)
	assert.Nil(t, err, "failed when check a ConfigMap which contains ks-jenkins.yaml")
	assert.Equal(t, "ks-jenkins", targetConfigMap.Data["jenkins_user.yaml"],
		"the existing ks-jenkins.yaml should not be override")
}

func TestCheckJenkinsConfigFormula(t *testing.T) {
	ctrl := Controller{}
	cm := &v1.ConfigMap{}
	err := ctrl.checkJenkinsConfigFormula(cm)
	assert.Nil(t, err, "failed when check the Jenkins config formula from an empty ConfigMap")
	assert.Equal(t, "custom", cm.Annotations["devops.kubesphere.io/jenkins-config-formula"],
		"the formula name should be custom if it's empty")

	cm = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"devops.kubesphere.io/jenkins-config-formula": "invalid",
			},
		},
	}
	err = ctrl.checkJenkinsConfigFormula(cm)
	assert.Nil(t, err, "failed when check the Jenkins config which contains an invalid formula name")
	assert.Equal(t, "custom", cm.Annotations["devops.kubesphere.io/jenkins-config-formula"],
		"the formula name should be custom if it's invalid")

	cm = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"devops.kubesphere.io/jenkins-config-formula": "high",
			},
		},
	}
	err = ctrl.checkJenkinsConfigFormula(cm)
	assert.Nil(t, err, "failed when check the Jenkins config which contains an valid formula name")
	assert.Equal(t, "high", cm.Annotations["devops.kubesphere.io/jenkins-config-formula"],
		"the formula name should not be changed if it's high")
}

func Test_isValidJenkinsConfigFormulaName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "emtpy string",
		args: args{},
		want: false,
	}, {
		name: "invalid name",
		args: args{name: "invalid"},
		want: false,
	}, {
		name: "formula: custom",
		args: args{name: "custom"},
		want: true,
	}, {
		name: "formula: high",
		args: args{name: "high"},
		want: true,
	}, {
		name: "formula: low",
		args: args{name: "low"},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidJenkinsConfigFormulaName(tt.args.name); got != tt.want {
				t.Errorf("isValidJenkinsConfigFormulaName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isJenkinsConfigCustomized(t *testing.T) {
	type args struct {
		annos map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "map is nil",
		want: false,
	}, {
		name: "map is empty",
		want: false,
	}, {
		name: "with valid value: custom",
		args: args{
			annos: map[string]string{
				ANNOJenkinsConfigFormula: "custom",
			},
		},
		want: true,
	}, {
		name: "with valid value: customized",
		args: args{
			annos: map[string]string{
				ANNOJenkinsConfigCustomized: "true",
			},
		},
		want: true,
	}, {
		name: "with invalid value: fake",
		args: args{
			annos: map[string]string{
				ANNOJenkinsConfigFormula: "fake",
			},
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isJenkinsConfigCustomized(tt.args.annos); got != tt.want {
				t.Errorf("isJenkinsConfigCustomized() = %v, want %v", got, tt.want)
			}
		})
	}
}
