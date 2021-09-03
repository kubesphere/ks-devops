## Nightly

KubeSphere provides [nightly build](https://en.wikipedia.org/wiki/Daily_build) for users to try the latest features.

```shell
kk create cluster --with-kubesphere nightly-$(date -d yesterday '+%Y%m%d')
```

## Install via helm

First, please clone this git repository. Then run command: `make install-chart`

> Because the default registry uses `ghcr.io/kubesphere`, if you want to use `docker.io` as the registry for image pull, 
> you can exec `helm install ks-devops chart/ks-devops --set image.registry=kubespheredev`

## KubeSphere CLI

[ks](https://github.com/kubesphere-sigs/ks) is a tool which makes it be easy to work with [KubeSphere](https://github.com/kubesphere/kubesphere).

> This still is a KubeSphere SIG level project. Please consider using it for study purpose.

[ks](https://github.com/kubesphere-sigs/ks) allows you to install KubeSphere on different platforms, such as [k3d](https://github.com/rancher/k3d), 
[kind](https://github.com/kubernetes-sigs/kind), etc.

Why using ks to install KubeSphere? Just because it is easy and simple. You don't need to prepare any dependency requirements. 
More than that, it also can help to enable any KubeSphere component during the installation stage.

### kubekey

```shell
ks install kk --components devops
```

### k3d

[k3d](https://github.com/rancher/k3d) is a little helper to run Rancher Lab's [k3s](https://github.com/k3s-io/k3s) in Docker.

```shell
ks install k3d --components devops
```

### kind

[kind](https://github.com/kubernetes-sigs/kind) is a tool for running local Kubernetes clusters using Docker container `nodes`.

```shell
ks install kind --components devops
```
