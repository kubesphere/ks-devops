## Background

Technically, GitOps does not have a strong relationship with CI (Continous Integration). 
It even does not care about where those images come from. The core concept of GitOps is 
having a Git repository that always contains declarative descriptions of the infrastructure 
currently desired in the target environment and an automated process to make the 
environment match the described state in the repository.

But people usually want to update the images in some environments if there are some specific images changed.
The biggest challenge might be the git repository part. A Kubernetes-based Application might be 
installed from different formats, such as Helm charts, Kustomization or others.

## Solution
This document is going to explain all these details and try to provide a solution and design for it.

[Argo CD][https://github.com/argoproj] and [Flux CD](https://github.com/fluxcd) are two very popular GitOps
solutions. They all have their way to update the images automatically. See also the following projects:

* [Argo CD Image Updater](https://github.com/argoproj-labs/argocd-image-updater)
  * Not ready for the production environment yet
* [Image Reflector Controller](https://github.com/fluxcd/image-reflector-controller) and [Image Automation Controller](https://github.com/fluxcd/image-automation-controller)

Both of them adopt `Apache 2.0` as the open-source License. So, we can integrate them into this project.

For the Flux CD, its solution will watch the changes of container images, 
then update the images according to some kind of marker expressions. For example,
add something like this `image: nginx # {"$imagepolicy": "flux-system:nginx"}` to a Deployment or StatefulSet. See also [more details](https://fluxcd.io/docs/guides/image-update/).

But Argo CD will focus on the Application level. It will watch all the changes of the container 
images that are annotated in the Applications.

Since `v3.3.0`, KubeSphere DevOps already integrated Argo CD as the GitOps Engine. And we have 
a plan [to integrate Flux CD](https://github.com/kubesphere/community/blob/master/sig-advocacy-and-outreach/ospp-2022/ks-devops-fluxcd-integrations_zh-CN.md) as well. So, we need to figure out a way to provide a 
consistent way to help users to use it easily.

Providing a CRD (Custom Resource Definition) as the abstraction of the Image update process is a good idea. We could have different implementations of the CRD, but the end-users don't need 
to know the implementation details.

## Design
Below is a sample of the CRD.

```yaml
apiVersion: gitops.kubesphere.io/v1alpha1
kind: ImageUpdater
metadata:
  name: demo
spec:
  ## 
  images:
  - nginx:^0.1
  - myalias=some/image
  kind: argocd | fluxcd
  argocd:
    app:
      name: demo
      namespace: demo
    write: built-in | git
    update-strategy:
      nginx: semver
      myalias: latest
    allow-tags:
      nginx: any
    ignore-tags:
      tomcat: "*-dev"
    platforms:
      nginx: linux/amd64
    secret:
      nginx: docker-hub-secret
  fluxcd:
    interval: 1m
    secret: docker-hub-secret
    policy:
      semver:
        range: 5.0.x
    sourceRef:
      kind: GitRepository
      name: podinfo
    git:
      checkout:
        ref:
          branch: main
      commit:
        author:
          email: fluxcdbot@users.noreply.github.com
          name: fluxcdbot
        messageTemplate: '{{range .Updated.Images}}{{println .}}{{end}}'
      push:
        branch: main
    update:
      path: ./clusters/my-cluster
      strategy: Setters
```

Provide one or more controllers to update the images accoding to `spec.engine`. Basically, the controllers will be a 
bridge between the CRD and the actual implementation.

APIs are required for the image automation update feature.

| API | Description |
|---|---|
| POST `/namespaces/{namespace}/imageupdaters` | Create an image update configuration |
| GET `/namespaces/{namespace}/imageupdaters/{imageupdater}` | Returns the details of a specific imageUpdater |
| GET `/namespaces/{namespace}/imageupdaters` | Return the list of imageUpdaters |
| PUT `/namespaces/{namespace}/imageupdaters/{imageupdater}` | Update a specific imageUpdater |
| DELETE `/namespaces/{namespace}/imageupdaters/{imageupdater}` | Delete a specific imageUpater |
