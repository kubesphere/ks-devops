We need to pay attention on the permission when there are APIs changed.
There are three types of permissions:

* Anonymous
* Global or cluster level
* Resource specific

Please don't forget to add the corresponding permission setting to [role-templates.yaml](https://github.com/kubesphere/ks-installer/blob/master/roles/ks-core/prepare/files/ks-init/role-templates.yaml) 
when you are trying to change (add, remove) any APIs.

## Anonymous
You could update the `GlobalRole` which is [anonymous](https://github.com/kubesphere/ks-installer/blob/e9e399d74a2fd8dbbb6477a95afb91c40f423b72/roles/ks-core/prepare/files/ks-init/role-templates.yaml#L91) when you are trying to create a new anonymous API.

## Global
As we know, some APIs does not belong to any CR (custom resource). For example: `ci/nodelabels`.
Update [here](https://github.com/kubesphere/ks-installer/blob/e9e399d74a2fd8dbbb6477a95afb91c40f423b72/roles/ks-core/prepare/files/ks-init/role-templates.yaml#L175) when you update a global API.

## Resource specific
You could create (or update) an CR (custom resource) that is `RoleBase` when you are trying to 
create a new resource specific API. Such as, [role-template-manage-pipelines](https://github.com/kubesphere/ks-installer/blob/e9e399d74a2fd8dbbb6477a95afb91c40f423b72/roles/ks-core/prepare/files/ks-init/role-templates.yaml#L3323).

## Examples

### Allow authenticated users checking Pipeline running log

Please feel free to add the following item into a `GlobalRole` of `authenticated`:

You could find it by: `kubectl edit globalrole authenticated`

```yaml
rules:
- apiGroups:
  - devops.kubesphere.io
  resources:
  - devops
  - pipelines
  - pipelines/branches
  - pipelineruns
  - pipelineruns/nodedetails
  verbs:
  - get
```
