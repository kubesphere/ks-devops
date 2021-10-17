Jenkins Configuration Controller

## Features

* Reload Jenkins configuration automatically
* Easily switch Jenkins configuration between pre-defined and custom
* Smoothly upgrade Jenkins configuration
* Jenkins configuration management in GitOps way (TODO)

## Background

[Configuration as Code](https://github.com/jenkinsci/configuration-as-code-plugin) is the crucial part of Jenkins configuration. 
Users need to modify the YAML file (in CasC format) if they want to change Jenkins configuration.

Since [CasC v1.51](https://github.com/jenkinsci/configuration-as-code-plugin/releases/tag/configuration-as-code-1.51), it 
only support loading config files from a local file or URL. So, Jenkins cannot loads the config file from Kubernetes (e.g. ConfigMap). 
Our solution is that mount a ConfigMap into the Jenkins Pod.

[kubelet](https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/#kubelet-config-k8s-io-v1beta1-KubeletConfiguration) 
can reload a ConfigMap into a Pod automatically once it detects changes from ConfigMaps. But it's not real time. 
This [controller](https://github.com/kubesphere/ks-devops/tree/master/controllers/jenkinsconfig) will let Jenkins try to 
reload the YAML config file.

It's easy to install Jenkins helm chart for once. But you might meet other issues when you try to upgrade it. Helm will 
overwrite all ConfigMaps when you try to upgrade a chart.

## Design

This is an example of ConfigMap for Jenkins configuration:

```yaml
apiVersion: v1
kind: ConfigMap
data:
  jenkins.yaml: |
    xxx: xxx
```

1. Make sure the new config file exists
   1. Copy `jenkins.yaml` into `ks-jenkins.yaml` if data `ks-jenkins.yaml` not exists
   1. Add annotation `devops.kubesphere.io/ks-jenkins-config: ks-jenkins.yaml`
1. Make sure the annotation `devops.kubesphere.io/jenkins-config-formula: xxx` exists
   1. Set the value of this annotation as `custom` if it's invalid (support values: `low`, `high`, `custom`)
   1. Add annotation `devops.kubesphere.io/jenkins-config-customized: "true"` if it's custom
1. Provide the pre-defined configuration according to the formula name
   1. Skip this process if the formula is custom (or invalid)
   1. Take `jenkins.yaml` as a template to transform the configuration
1. Reload CasC via Jenkins API
   1. Change the config file to `ks-jenkins.yaml`

An example of target ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotation:
    devops.kubesphere.io/ks-jenkins-config: ks-jenkins.yaml
    devops.kubesphere.io/jenkins-config-formula: custom
    devops.kubesphere.io/jenkins-config-customized: "true"
data:
  jenkins.yaml: |
    xxx: xxx
  ks-jenkins.yaml: |
    xxx: xxx
```

## How-to

Users should only modify the configuration from `ks-jenkins.yaml`, and make sure the annotation has an expected value
`devops.kubesphere.io/jenkins-config-formula: custom`.
