## Get started

1. Install KubeSphere via [kk](https://github.com/kubesphere/kubekey/) (or other ways). 
    This is an optional step, basically we need a Kubernetes Cluster and the front-end of DevOps.
1. Install `ks-jenkins` via enabling the DevOps component in KubeSphere (Remove this step after we combine Jenkins chart with `ks-devops`)
1. Install `ks-devops` via [chart](charts/ks-devops)

In current phase, we need to use a temporary images of [KubeSphere](https://github.com/kubesphere/kubesphere/) 
which comes from [the branch remove-devops-ctrl](https://github.com/LinuxSuRen/kubesphere/tree/remove-devops-ctrl):

* `kubespheredev/ks-apiserver:remove-devops-ctrl`
* `kubespheredev/ks-controller-manager:remove-devops-ctrl`

KubeSphere is a proxy for `ks-devops`. We need to tell it the address of `ks-devops`. Please change the ConfigMap 
of `kubesphere-config` in namespace `kubesphere-system`:

```
devops:
  enable: false
  devopsServiceAddress: 127.0.0.1:9091
```

### Run it locally

Technically, [apiserver](cmd/apiserver) and [controller](cmd/controller) are all binary files. So, 
it's possible to run them in your local environment. You just need to make sure that the connection 
between your environment and a Kubernetes cluster works well. This is a default config file of these 
components, please see also [the sample file](config/samples/kubesphere.yaml).

## Create Pipeline via CLI

[ks](https://github.com/linuxsuren/ks) is an official client of KubeSphere. You can create a Pipeline by it.

`ks pip create --ws simple --template java --project default --skip-check -b good`

## APIs

For example, you can access an API like:

`curl http://ip:30880/kapis/clusters/host/devops.kubesphere.io/v1alpha3/devops/test847h4/credentials`

## TODO

* A separate front-end project of ks-devops
* Add a Helm chart for [s2i-operator](https://github.com/kubesphere/s2ioperator)
* Migrate Jenkins Helm chart from [ks-installer](https://github.com/kubesphere/ks-installer/tree/master/roles/ks-devops/jenkins/files/jenkins/jenkins)
* Auth support
    * OIDC support as a default provider
