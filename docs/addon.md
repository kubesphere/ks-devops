Addon means an extension of ks-devops. For intance: [Jenkins](http://jenkins.io/), [Argo CD](https://github.com/argoproj/argo-cd/),
[SonarQube](https://www.sonarqube.org/), [ks-releaser](https://github.com/kubesphere-sigs/ks-releaser/), etc.

There are two CRDs:

* `Addon`, it represents a component
* `AddonStrategy`, it describes how to install an addon.

## Supported addons

| Name                                                           | Description                                                               |
|----------------------------------------------------------------|---------------------------------------------------------------------------|
| [ks-releaser](https://github.com/kubesphere-sigs/ks-releaser/) | Help to release a project which especially has multiple git repositories. |
| [Argo CD](https://github.com/argoproj/argo-cd/)                | Declarative continuous deployment for Kubernetes.                         |

## How to use?

The `Addon Controller` is optional, please add the flag `--enabled-controllers addon=true` into the controller command line.
For instance:

```yaml
spec:
  containers:
    - args:
      - --enabled-controllers
      - addon=true
      image: ghcr.io/kubesphere/devops-controller
```

then, create the `AddonStrategy`: `simple-operator-argocd` and `ks-releaser-simple-operator`. You can find the YAML files from [here](../config/samples/addon).

install operator, for instance:

```shell
kubectl apply -f https://github.com/kubesphere-sigs/ks-releaser-operator/releases/download/v0.0.2/install.yaml
```

finally, you can install `ks-releaser` by adding the following resource:

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: Addon
metadata:
  name: ks-releaser
spec:
  version: v0.0.14
  strategy:
    name: simple-operator-releasercontroller
```

## Support more?

Want to support more addons? It would be easy if you can find it from the [operator hub](https://operatorhub.io/).

> Restriction:
> * Require install desired operator manually.
> * [Hard code](../controllers/addon/operator_controller.go) about the supported addons
