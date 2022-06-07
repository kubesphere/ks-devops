## Background
Currently(before v3.3.0), [all Pipeline step UI components](https://github.com/kubesphere/console/tree/master/src/pages/devops/components/Pipeline/StepModals) belong to `.jsx` source file. So, it's not easy to add a new Pipeline step UI component if a contributor is not familiar with the relevant information.

A second big problem with the current pattern is that some steps (or commands) might be difficult for some newcomers. In some cases, some steps need a nested step. For example:

```groovy
container('base'){
  sh 'kubectl get ns'
}

## CRD
Here is a sample of the step template:

```yaml
apiVersion: devops.kubesphere.io/v1alpha3
kind: TaskTemplate
metadata:
  name: docker-login
  annotations:
    devops.kubesphere.io/step-name-zh: "Docker 登录"
    devops.kubesphere.io/step-description-zh: "Docker 登录"
spec:
  name: 'Docker login'
  description: 'Docker login'
  secret:
    type: basic-auth
    wrap: true
    mapping:                           # this is a map, the following key-values are default
      passwordVariable: PASSWORDVARIABLE
      usernameVariable: USERNAMEVARIABLE
      variable: VARIABLE
      sshUserPrivateKey: SSHUSERPRIVATEKEY
      keyFileVariable: KEYFILEVARIABLE
      passphraseVariable: PASSPHRASEVARIABLE
  container: 'base'
  runtime: 'shell|dsl'
  template: |
    docker login -u $username xx -p $password {{.param.registry}}
  parameters:
  - name: registry
    required: true
    display: Image Registry address
    defaultValue: docker.io
```

The result looks like:
```groovy
withCredential[usernamePassword(credentialsId : "$ID" ,passwordVariable : 'PASSWD' ,usernameVariable : 'USER' ,)]) {
  conatiner('base'){
    sh '''
      docker login -u $USER -p $PASSWD docker.io
    '''
  }
}
```

## How to use
API request sample:
```shell
curl http://ip:port/pipeline/steps/render -X POST \
  -d '{"message": "demo message", "secret": "secret-id"}'
```

Response sample:
```groovy
container('base') {
  withCredentials([string(credentialsId : 'feishu' ,variable : 'ID' ,)]) {
    sh '''
      curl -X POST -H "Content-Type: application/json" \
      -d '{"msg_type":"text","content":{"text":"request example"}}' \
      https://open.feishu.cn/open-apis/bot/v2/hook/$ID
    '''
  }
}
```
