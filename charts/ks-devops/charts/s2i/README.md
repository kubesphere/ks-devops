# S2i Helm Charts

- Install s2i charts into Kubernetes

```shell
cd charts/ks-devops/charts/s2i
helm install s2ioperator . -n kubesphere-devops-system --create-namespace
```

- Debug s2i charts locally

```shell
cd charts/ks-devops/charts/s2i
helm install s2ioperator . --debug --dry-run -n kubesphere-devops-system --create-namespace
```

## Configuration

The following tables list the configurable parameters of the s2i chart and their default values.

### s2i image

| Parameter      | Description                                                                                                | Default       |
| -------------- | ---------------------------------------------------------------------------------------------------------- | ------------- |
| image.registry | docker image registry related s2i. Available value: docker.io/kubespheredev                                | docker.io/kubespheredev |
