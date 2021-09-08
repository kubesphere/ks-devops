# Description of testdata directory

* This directory is used to place files needed in the integration test, including:
  * Tekton CRD yaml files
    * These files are downloaded from [tektoncd/pipeline](https://github.com/tektoncd/pipeline/tree/main/config).
    * Since we hold `v0.25.0` version of Tekton in our helm chart, so here we keep consistent with the version in our chart, and download files from [v0.25.0](https://github.com/tektoncd/pipeline/tree/v0.25.0/config).

* Details of `hack/download_tekton_testdata.sh`
  * This script is used to download Tekton CRD yaml files, including `Pipeline`, `PipelineRun` and `Task`.
  * It will download yaml files at the current directory.
  * Usage
    * `bash hack/download_tekton_testdata.sh TektonVersion`
    * e.g. `bash hack/download_tekton_testdata.sh v0.25.0`
