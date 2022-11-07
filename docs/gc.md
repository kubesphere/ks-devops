We prefer to use the built-in PipelineRun GC instead of Jenkins itself. Please feel free to configure [the `CronJob`](https://github.com/kubesphere-sigs/ks-devops-helm-chart/blob/464a1a9854561ef5666433b8d975b89cced07494/charts/ks-devops/templates/cronjob-gc.yaml) according to your requirement.

Normally, you could find it from namespace `kubesphere-devops-system`. You could specifiy the following options:

* `maxAge` is the maximum time to live for PipelineRuns
* `maxCount` is the max number of the PipelineRuns 

See also the Helm chart [values setting](https://github.com/kubesphere-sigs/ks-devops-helm-chart/blob/464a1a9854561ef5666433b8d975b89cced07494/charts/ks-devops/values.yaml#L17).
