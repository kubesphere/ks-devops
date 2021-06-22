There're several ways to trigger Pipeline in KubeSphere DevOps.

## Manual

TODO.

## Schedule

TODO.

## API

You can trigger a Pipeline via API. For example:

```
curl -X POST 'http://ip:port/kapis/devops.kubesphere.io/v1alpha2/devops/testxnlvz/pipelines/good/runs?token=xxxx' \
  -H 'Cookie: token=ks-token-xxx' \
  -H 'content-type: application/json' \
  -H 'Jenkins-Crumb: a7b67cdd408ba315cd0d1639a30961de0b621ca4c926569abc1ac461a3815170'
```
