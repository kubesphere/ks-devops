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

// ResourceLimit describes the limitation level of jenkins agent pod resource
type ResourceLimit string

const (
	// Resource limit key of map
	resourceLimitKey = "resource.limit"

	jenkinsConfigName     = "jenkins-casc-config"
	jenkinsYamlKey        = "jenkins.yaml"
	jenkinsUserYamlKey    = "jenkins_user.yaml"
	jenkinsCasCConfigName = "jenkins-casc-config"
	workerLimitRangeName  = "worker-limit-range"
	workerResQuotaName    = "worker-resource-quota"
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
