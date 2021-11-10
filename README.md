[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/kubesphere/ks-devops)
[![](https://goreportcard.com/badge/kubesphere/ks-devops)](https://goreportcard.com/report/kubesphere/ks-devops)
[![codecov](https://codecov.io/gh/kubesphere/ks-devops/branch/master/graph/badge.svg?token=XS8g2CjdNL)](https://codecov.io/gh/kubesphere/ks-devops)
[![Contributors](https://img.shields.io/github/contributors/kubesphere/ks-devops.svg)](https://github.com/kubesphere/ks-devops/graphs/contributors)

KubeSphere DevOps integrates popular CI/CD tools, provides CI/CD Pipelines based on Jenkins, offers automation toolkits 
including Binary-to-Image (B2I) and Source-to-Image (S2I), and boosts continuous delivery across Kubernetes clusters.

With the container orchestration capability of Kubernetes, KubeSphere DevOps scales Jenkins Agents dynamically, improves 
CI/CD workflow efficiency, and helps organizations accelerate the time to market for their products.

## Features

* Out-of-the-Box CI/CD Pipelines
* Built-in Automation Toolkits for DevOps with Kubernetes
* Use Jenkins Pipelines to Implement DevOps on Top of Kubernetes
* Manage Pipelines via [CLI](docs/cli.md)

## Get Started

### Quick Start

- Install KubeSphere via [KubeKey](https://github.com/kubesphere/kubekey/) (or [the methods described here](docs/installation.md)).

  ```bash
  kk create cluster --with-kubesphere
  ```

- Enable DevOps application

  ```bash
  kubectl patch -nkubesphere-system cc ks-installer --type=json -p='[{"op": "replace", "path": "/spec/devops/enabled", "value": true}]'
  ```
For more information, refer to the [documentation](docs).

## Next Steps

- A Separate Front-End Project of KS-DevOps
- Auth Support
  - OIDC support as a default provider


## Communication Channels

- [KubeSphere DevOps Google Group](https://groups.google.com/g/kubesphere-sig-devops/)
- DevOps Slack Channels for [English Speakers](https://kubesphere.slack.com/archives/C010TH02010) and [Chinese Speakers](https://kubesphere.slack.com/archives/C026V4FBWBW)
- [Forum for Chinese Speakers](https://kubesphere.com.cn/forum/t/DevOps)
- [KubeSphere DevOps Special Interest Group](https://github.com/kubesphere/community/tree/master/sig-devops)

## Contribution

Looking forward to becoming a part of us?

Feel free to go through the [Contribution Guide](CONTRIBUTING.md), pick up a [good-first-issue](https://github.com/kubesphere/ks-devops/contribute), and create a pull request.

Thanks to all the people who have already contributed to KS-DevOps!

<a href="https://github.com/kubesphere/ks-devops/graphs/contributors"><img src="https://opencollective.com/ks-devops/contributors.svg?width=890&button=false" /></a>
