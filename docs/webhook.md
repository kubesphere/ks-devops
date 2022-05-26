There are several ways to trigger Pipeline via [webhook](https://en.wikipedia.org/wiki/Webhook) in KubeSphere DevOps.

* Generic webhook
* SCM Webhook

## SCM Webhook

Supported SCM providers:
* GitHub
* Gitlab
* Bitbucket

There are two types of Jenkins based Pipelines: regular or multi-branch Pipeline. When a SCM webhook request received,
the server will search all Pipelines by the Git URL, then trigger the scan action if it's a multi-branch Pipeline,
or create a new PipelineRun if there is an annotation key-value likes the following one:
```
scm.devops.kubesphere.io=https=https://github.com/linuxsuren/tools
```

In case you only want some Pipelines to be triggered when specific branches changed. You can add an annotation:
```
scm.devops.kubesphere.io/ref='["master","fea-.*"]'
```

The webhook address is:
```
http://ip:port/v1alpha3/webhooks/scm
```

### Using webhook locally

It's also possible to use webhook feature locally. You just need to start a proyx with [ngrok](https://ngrok.com/).

## Automatic webhook

TODO

## Generic Webhook

It does not require a particular payload in this kind of webhook. It accepts a standard 
HTTP request with some a generic payload. Users can use it from GitHub or just a curl command line. The backend rely on 
[Jenkins generic-webhook-trigger plugin](https://github.com/jenkinsci/generic-webhook-trigger-plugin).

For example, you can use the following command line to trigger some Pipelines:

`curl -H "Content-Type: application/x-www-form-urlencoded" -X POST "http://ip:port/kapis/devops.kubesphere.io/v1alpha2/webhook/generic-trigger?token=xxxx"`

You can get the output like below if you have a valid token:

```
{
  "jobs": {
    "testxnlvz/test": {
      "regexpFilterExpression": "",
      "triggered": true,
      "resolvedVariables": {},
      "regexpFilterText": "",
      "id": 20,
      "url": "queue/item/20/"
    },
  "message": "Triggered jobs."
}
```

You can get the output like blew if you don't have a valid token:

```
{
  "jobs": null,
  "message": "Did not find any jobs with GenericTrigger configured! If you are using a token, you need to pass it like ...trigger/invoke?token=TOKENHERE. If you are not using a token, you need to authenticate like http://user:passsword@jenkins/generic-webhook... "
}
```
