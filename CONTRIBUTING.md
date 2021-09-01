Welcome to go through the contribution guide!

## Prepare yourself

Before you get started to contribute. Please the following requirements:

* [Golang](https://golang.org/) 1.13

| Technical area | Level requires | Links |
|---|---|---|
| Golang | Skilled | [Youtube tutorial for beginner](https://www.youtube.com/watch?v=75lJDVT1h0s&list=PLzMcBGfZo4-mtY_SE3HuzQJzuj4VlUG0q) |
| GitHub | Skilled | [Official learning lab](https://lab.github.com/) |
| Kubernetes | Basic development skills | [Official document](https://kubernetes.io/) |
| Helm chart | Basic development skills | [Official document](https://helm.sh/) |
| Docker | Skilled | [Official document](https://docs.docker.com/) |
| Jenkins | Familiar [optional] | [Official website](https://www.jenkins.io/), [Tutorial in Chinese](https://www.bilibili.com/video/BV1fp4y1r7Dd) |

## Run it locally

Technically, [apiserver](cmd/apiserver) and [controller](cmd/controller) are all binary files. So,
it's possible to run them in your local environment. You just need to make sure that the connection
between your environment and a Kubernetes cluster works well. This is a default config file of these
components, please see also [the sample file](config/samples/kubesphere.yaml).

## Development locally

- Run [kind](https://github.com/kubernetes-sigs/kind) in local or remote machine
- Make sure that you can access cluster via kubectl command in local machine
- Execute the following command to install our CRDs:

```shell
make install
```

- Debug code...

- Execute the following command to uninstall our CRDs:

```shell
make uninstall
```

## Experimental support

[octant-ks-devops](https://github.com/LinuxSuRen/octant-ks-devops) is a plugin of [octant](https://github.com/vmware-tanzu/octant/).
It provides a dashboard for Kubernetes and ks-devops.

## APIs

For example, you can access an API like:

```shell script
curl -H "X-Authorization: Bearer xxxx" \
  http://localhost:9090/kapis/devops.kubesphere.io/v1alpha3/devops/testblpsz/pipelines
```

> Please get a token from Kubernetes cluster, and replace `xxxx` with it.

If you want to see ks-devops postman API collection , please visit **[ks-devops postman](https://www.postman.com/ks-devops/workspace/kubesphere-devops)**

## Code contribution

If you're going to update or add CRD go struct, please run the following command once done with that:

`make manifests generate generate-listers`

then, it can generate CRDs and DeepCopy methods.

## Lint your codes

We are using [golangci-lint](https://golangci-lint.run/) as our code linter. Before you make some code changes, please execute following command to check code style:

```shell
golangci-lint run
# Or with specified folder, e.g.
golangci-lint run controllers/jenkinsconfig
```

## References

- [Collection of KubeSphere Devops related projects](docs/projects.md)
- DevOps SIG meeting videos in [Chinese](https://space.bilibili.com/438908638/channel/seriesdetail) and [English](https://www.youtube.com/watch?v=c3V-2RX9yGY&list=PLwDEgvYeF0jL-CAJ9SpCx_QWKMDGLKqgN)
- [Open source best practice](https://github.com/LinuxSuRen/open-source-best-practice) (mostly written in Chinese)
- [Open source guides](https://opensource.guide/) are written by GitHub