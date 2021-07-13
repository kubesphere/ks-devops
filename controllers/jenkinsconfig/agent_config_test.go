package jenkinsconfig

import (
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	config := getDefaultConfig()
	assert.NotNil(t, config)
	assert.NotNil(t, config["pod.concurrent"])
}

func TestGetHighConfig(t *testing.T) {
	config := getHighConfig()
	assert.NotNil(t, config)
	assert.NotNil(t, config["pod.concurrent"])
}

func TestCreateJenkinsCasCTemplate(t *testing.T) {
	tests := []struct {
		name          string
		template      string
		namespace     string
		initConfigMap *v12.ConfigMap
		wantErr       bool
	}{{
		name:      "valid arguments",
		template:  rand.String(100),
		namespace: rand.String(20),
	}, {
		name:      "empty namespace",
		template:  rand.String(100),
		namespace: "",
	}, {
		name:      "empty template",
		template:  "",
		namespace: rand.String(20),
	}, {
		name:      "existing template",
		template:  "test-template",
		namespace: "test-namespace",
		wantErr:   true,
		initConfigMap: &v12.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      templateConfigMapName,
				Namespace: "test-namespace",
			},
		},
	}}
	for _, tt := range tests {
		var clientset *fake.Clientset
		if tt.initConfigMap != nil {
			clientset = fake.NewSimpleClientset(tt.initConfigMap)
		} else {
			clientset = fake.NewSimpleClientset()
		}
		var configMapGetter v1.ConfigMapsGetter = clientset.CoreV1()
		t.Run(tt.name, func(t *testing.T) {
			tmplConfigMap, err := createJenkinsCascTemplate(configMapGetter, tt.namespace, tt.template)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, true, *tmplConfigMap.Immutable)
			assert.Equal(t, tt.namespace, tmplConfigMap.Namespace)
			assert.Equal(t, templateConfigMapName, tmplConfigMap.Name)
			assert.Equal(t, tt.template, tmplConfigMap.Data[jenkinsYamlKey])
		})
	}
}

func TestGetOrCreateJenkinsCasCTemplate(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		template      string
		wantErr       bool
		initConfigMap *v12.ConfigMap
	}{{
		name:      "valid arguments",
		namespace: rand.String(20),
		template:  rand.String(100),
	}, {
		name:      "empty namespace",
		namespace: "",
		template:  rand.String(100),
	}, {
		name:      "empty template",
		namespace: rand.String(20),
		template:  "",
	}, {
		name:      "existing configmap",
		namespace: "test-namespace",
		template:  rand.String(100),
		initConfigMap: &v12.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      templateConfigMapName,
				Namespace: "test-namespace",
			},
			Data: map[string]string{
				jenkinsYamlKey: "test-template",
			},
		},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset()
			tmpl, err := getOrCreateJenkinsCasCTemplate(clientset.CoreV1(), tt.namespace, tt.template)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.template, tmpl)
		})
	}

}

func TestIsDefaultResourceLimit(t *testing.T) {
	tests := []struct {
		name          string
		resourceLimit string
		isDefault     bool
	}{
		{
			name:          "default resource limit",
			resourceLimit: string(defaultLimit),
			isDefault:     true,
		},
		{
			name:          "Default resource limit",
			resourceLimit: "Default",
			isDefault:     true,
		},
		{
			name:          "empty resource limit",
			resourceLimit: "",
			isDefault:     true,
		},
		{
			name:          "high resource limit",
			resourceLimit: string(highLimit),
			isDefault:     false,
		},
		{
			name:          "custom resource limit",
			resourceLimit: string(customLimit),
			isDefault:     false,
		},
		{
			name:          "random resource limit",
			resourceLimit: rand.String(10),
			isDefault:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isDefault, isDefaultResourceLimit(tt.resourceLimit))
		})
	}
}
