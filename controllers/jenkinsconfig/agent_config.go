package jenkinsconfig

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	"strings"
)

// ResourceLimit describes the limitation level of jenkins agent pod resource
type ResourceLimit string

const (
	// default resource limit level
	defaultLimit ResourceLimit = "default"

	// high resource limit level
	highLimit ResourceLimit = "high"

	// custom resource limit level
	customLimit ResourceLimit = "custom"

	// Resource limit key of map
	resourceLimitKey = "resource.limit"

	jenkinsConfigName          = "jenkins-agent-config"
	podResourceLimitConfigName = "agent.pod_resource_limit"
	customizedLabel            = "config.devops.kubesphere.io/customized"
	templateConfigMapName      = "jenkins-casc-config-template"
	jenkinsYamlKey             = "jenkins.yaml"
	jenkinsCasCConfigName      = "jenkins-casc-config"
	workerLimitRangeName       = "worker-limit-range"
	workerResQuotaName         = "worker-resource-quota"
)

const (
	workerLimitCPUKey           = "worker.limit.cpu"
	workerLimitMemoryKey        = "worker.limit.memory"
	workerLRDefaultCPUKey       = "worker.limitrange.container.default.cpu"
	workerLRDefaultMemoryKey    = "worker.limitrange.container.default.memory"
	workerLRDefaultReqCPUKey    = "worker.limitrange.container.defaultrequest.cpu"
	workerLRDefaultReqMemoryKey = "worker.limitrange.container.defaultrequest.memory"
	podConcurrentKey            = "pod.concurrent"
	jnlpLimitCPUKey             = "jnlp.limit.cpu"
	jnlpLimitMemoryKey          = "jnlp.limit.memory"
	baseLimitCPUKey             = "base.limit.cpu"
	baseLimitMemoryKey          = "base.limit.memory"
	nodejsLimitCPUKey           = "nodejs.limit.cpu"
	nodejsLimitMemoryKey        = "nodejs.limit.memory"
	mavenLimitCPUKey            = "maven.limit.cpu"
	mavenLimitMemoryKey         = "maven.limit.memory"
	goLimitCPUKey               = "go.limit.cpu"
	goLimitMemoryKey            = "go.limit.memory"
)

// getDefaultConfig gets default resource limit configuration
func getDefaultConfig() map[string]string {
	return map[string]string{
		workerLimitCPUKey:           "3000m",
		workerLimitMemoryKey:        "3Gi",
		workerLRDefaultCPUKey:       "750m",
		workerLRDefaultMemoryKey:    "1024Mi",
		workerLRDefaultReqCPUKey:    "100m",
		workerLRDefaultReqMemoryKey: "128Mi",

		podConcurrentKey:     "2",
		jnlpLimitCPUKey:      "500m",
		jnlpLimitMemoryKey:   "512Mi",
		baseLimitCPUKey:      "1000m",
		baseLimitMemoryKey:   "1024Mi",
		nodejsLimitCPUKey:    "1000m",
		nodejsLimitMemoryKey: "1024Mi",
		mavenLimitCPUKey:     "1000m",
		mavenLimitMemoryKey:  "1024Mi",
		goLimitCPUKey:        "1000m",
		goLimitMemoryKey:     "1024Mi",
	}
}

// getHighConfig gets high resource limit configuration
func getHighConfig() map[string]string {
	return map[string]string{
		workerLimitCPUKey:           "7000m",
		workerLimitMemoryKey:        "11Gi",
		workerLRDefaultCPUKey:       "2000m",
		workerLRDefaultMemoryKey:    "4Gi",
		workerLRDefaultReqCPUKey:    "200m",
		workerLRDefaultReqMemoryKey: "256Mi",

		podConcurrentKey:     "4",
		jnlpLimitCPUKey:      "500m",
		jnlpLimitMemoryKey:   "1536Mi",
		baseLimitCPUKey:      "3000m",
		baseLimitMemoryKey:   "4096Mi",
		nodejsLimitCPUKey:    "3000m",
		nodejsLimitMemoryKey: "4096Mi",
		mavenLimitCPUKey:     "3000m",
		mavenLimitMemoryKey:  "4096Mi",
		goLimitCPUKey:        "3000m",
		goLimitMemoryKey:     "4096Mi",
	}
}

// getOrCreateJenkinsCasCTemplate fetches or creates Jenkins CasC template. If there has no template before, the provided
// template will be used.
func getOrCreateJenkinsCasCTemplate(configMapClient v1core.ConfigMapsGetter, namespace string, template string) (string, error) {
	// get ConfigMap
	agentConfigTemplate, err := configMapClient.ConfigMaps(namespace).Get(context.Background(), templateConfigMapName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return "", err
		}
		klog.V(5).Infof("%s/%s was not found, and we would create a fresh one", namespace, templateConfigMapName)
		// create ConfigMap
		agentConfigTemplate, err = createJenkinsCascTemplate(configMapClient, namespace, template)
		if err != nil {
			return "", err
		}
	}
	return agentConfigTemplate.Data[jenkinsYamlKey], nil
}

// createJenkinsCascTemplate creates a newly Jenkins CasC template ConfigMap with provided namespace and template.
func createJenkinsCascTemplate(configMapClient v1core.ConfigMapsGetter, namespace string, template string) (*v1.ConfigMap, error) {
	immutable := true
	// construct new template ConfigMap
	templateConfigMap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      templateConfigMapName,
		},
		Data: map[string]string{
			jenkinsYamlKey: template,
		},
		Immutable: &immutable,
	}
	templateConfigMap, err := configMapClient.ConfigMaps(namespace).Create(context.Background(), templateConfigMap, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("failed to create Jenkins CasC template ConfigMap: %s/%s", namespace, templateConfigMapName)
		return nil, err
	}
	return templateConfigMap, nil
}

// Check if provided resource limit is considered as default.
func isDefaultResourceLimit(resourceLimit string) bool {
	return len(resourceLimit) == 0 || strings.Compare(strings.ToLower(resourceLimit), string(defaultLimit)) == 0
}
