[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/kubesphere/ks-devops)
[![](https://goreportcard.com/badge/kubesphere/ks-devops)](https://goreportcard.com/report/kubesphere/ks-devops)
[![codecov](https://codecov.io/gh/kubesphere/ks-devops/branch/master/graph/badge.svg?token=XS8g2CjdNL)](https://codecov.io/gh/kubesphere/ks-devops)
[![Contributors](https://img.shields.io/github/contributors/kubesphere/ks-devops.svg)](https://github.com/kubesphere/ks-devops/graphs/contributors)

KubeSphere DevOps integrates popular CI/CD tools, provides CI/CD Pipelines based on Jenkins, offers automation toolkits 
including Binary-to-Image (B2I) and Source-to-Image (S2I), and boosts continuous delivery across Kubernetes clusters.

With the container orchestration capability of Kubernetes, KubeSphere DevOps scales Jenkins Agents dynamically, improves 
CI/CD workflow efficiency, and helps organizations accelerate the time to market for products.

## Features

* Out-of-the-Box CI/CD Pipelines
* Built-in Automation Toolkits for DevOps with Kubernetes
* Use Jenkins Pipelines to Implement DevOps on Top of Kubernetes
* Manage Pipelines via [CLI](docs/cli.md)

## Get started

### Quick Start

- Install KubeSphere via [kubekey](https://github.com/kubesphere/kubekey/) (or [other ways](docs/installation.md)).

  ```bash
  kk create cluster --with-kubesphere
  ```

- Enable DevOps application

  ```bash
  kubectl patch -nkubesphere-system cc ks-installer --type=json -p='[{"op": "replace", "path": "/spec/devops/enabled", "value": true}]'
  ```
Want to go into deep? Please check out the [documentation](docs).

## TODO

- A separate front-end project of ks-devops
- Auth support
  - OIDC support as a default provider

Want to be part of us? Please feel free to go through the [contribution guide](CONTRIBUTING.md), 
and pick up a [good-first-issue](https://github.com/kubesphere/ks-devops/contribute).

## Available communication channels:

- [KubeSphere DevOps google group](https://groups.google.com/g/kubesphere-sig-devops/)
- DevOps Slack channel for [English](https://kubesphere.slack.com/archives/C010TH02010) and [Chinese](https://kubesphere.slack.com/archives/C026V4FBWBW)
- [Forum for Chinese speakers](https://kubesphere.com.cn/forum/t/DevOps)
- [KubeSphere DevOps Special Interest Group](https://github.com/kubesphere/community/tree/master/sig-devops)
