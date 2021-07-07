[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/LinuxSuRen/ks-devops)
[![codecov](https://codecov.io/gh/linuxsuren/ks-devops/branch/master/graph/badge.svg?token=XS8g2CjdNL)](https://codecov.io/gh/linuxsuren/ks-devops)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FLinuxSuRen%2Fks-devops.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FLinuxSuRen%2Fks-devops?ref=badge_shield)

## Get started

1. Install KubeSphere via [kk](https://github.com/kubesphere/kubekey/) (or other ways). 
    This is an optional step, basically we need a Kubernetes Cluster and the front-end of DevOps.
1. Install `ks-devops` via [chart](charts/ks-devops)
1. Replace the images of `ks-apiserver` and `ks-controller-manager`. In current phase, we need to use a temporary images of [KubeSphere](https://github.com/kubesphere/kubesphere/) 
which comes from [the branch remove-devops-ctrl](https://github.com/LinuxSuRen/kubesphere/tree/remove-devops-ctrl):

* `kubespheredev/ks-apiserver:remove-devops-ctrl`
* `kubespheredev/ks-controller-manager:remove-devops-ctrl`

Want to go into deep? Please checkout the [documentation](docs).

### Install it as a Helm Chart

First, please clone this git repository. Then run command: `make install-chart`

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

```shell script
curl -H "Authorization: bearer xxxx" \
  http://localhost:9090/kapis/devops.kubesphere.io/v1alpha3/devops/testblpsz/pipelines
```

> Please get a token from Kubernetes cluster, and replace `xxxx` with it.

## Code contribution

If you're going to update or add CRD go struct, please run the following command once done with that:

`make manifests generate generate-listers`

then, it can generate CRDs and DeepCopy methods.

## TODO

* A separate front-end project of ks-devops
* Install `ks-devops` via helm chart in [ks-installer](https://github.com/kubesphere/ks-installer)
* Auth support
    * OIDC support as a default provider

## Experimental support

[octant-ks-devops](https://github.com/LinuxSuRen/octant-ks-devops) is a plugin of [octant](https://github.com/vmware-tanzu/octant/).
It provides a dashboard for Kubernetes and ks-devops.

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FLinuxSuRen%2Fks-devops.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FLinuxSuRen%2Fks-devops?ref=badge_large)
