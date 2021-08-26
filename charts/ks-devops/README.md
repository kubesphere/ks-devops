# `ks-devops` Helm Chart

Helm chart for [ks-devops](https://github.com/kubesphere/ks-devops).

## Chart details

This chart will do the following:
* Deploy components of ks-devops, e.g. apiserver, controller, etc.
* Install the dependency charts, such as jenkins, s2i, tekton (optional).

### Details of tekton chart package

Currently, we hold the fixed version (**v0.25.0**) of `tekton-pipeline` helm chart package. 

The [tekton-pipeline.tar.gz](charts/tekton-pipeline-0.25.0.tgz) is generated from [tekton-helm-chart](https://github.com/cdfoundation/tekton-helm-chart) repository. 

We apply the command `helm package tekton-pipeline --app-version 0.25.0` to package the tekton chart repository.


## Prerequisite

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

## Usage

### Jenkins installation

Once Helm is set up properly, you can install the chart with the release name `ks-ctl` using the following command in default values:

```bash
helm install ks-ctl . -n kubesphere-devops-system --create-namespace --set charts.jenkins.enabled=true 
```

### Tekton installation

If you want to use tekton as the pipeline backend, you are supposed to use the following command:

```bash
helm install ks-ctl . -n kubesphere-devops-system --create-namespace --set charts.tekton.enabled=true --set image.controller.tag=tekton-support
```
