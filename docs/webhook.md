There are several ways to trigger Pipeline via [webhook](https://en.wikipedia.org/wiki/Webhook) in KubeSphere DevOps.

* Generic webhook
* SCM Webhook

## SCM Webhook

TODO

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
