For many components, they can receive webhook events from other system. For instance, Argo CD can receive webhook events 
from a git repository.

The `GitRepository Webhook Controller` be able to manage the webhooks. You can install it without the whole ks-devops.

## Install

```shell
kustomize build ../config/samples/gitRepository_webhooks | kubectl apply -f -
```

## Get started

First, please prepare a secret of your git repository:

```shell
kubectl create secret generic github --from-literal=token=your-secret
```

then, please create the `Webhook` and `GitRepository`:

```shell
kubectl apply -f ../config/samples/webhook.yaml
kubectl apply -f ../config/samples/gitrepository.yaml
```

now, you can check your git repository. To see if it works well.

## More

Currently, we support GitHub, Gitlab. But thanks to [drone/go-scm](https://github.com/drone/go-scm), 
it's possible to support more git providers.
