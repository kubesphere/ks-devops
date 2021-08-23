# Tekton integration

* [Goal](##Goal)
* [Background](##Background)
* [Quick start](##Quick-start)
* [Design](##Design)
  * [Syntax design](###Syntax-design)
  * [Controllers implementation](###Controllers-implementation)

## Goal

Integrate [Tekton](https://github.com/tektoncd/pipeline) as an alternative CI/CD engine of [KubeSphere Devops](https://github.com/kubesphere/ks-devops).

## Background

Tekton is a cloud-native CI/CD project. Compared to Jenkins, Tekton can take full advantage of the Kubernetes ecosystem. For example, Tekton, as a serverless component, is very easy to scale. However, it could be difficult for new users to operate Tekton for its lack of user-friendly UI. We would like to make full use of the advantages of Tekton, and make it easy to use in KubeSphere.

## Quick start

The design of Tekton integration is currently on the branch [tekton-support](https://github.com/kubesphere/ks-devops/tree/tekton-support) under development.

If you want to try it out, you can follow the below instructions.

1. Change the branch to [tekton-support](https://github.com/kubesphere/ks-devops/tree/tekton-support).
    * `git checkout tekton-support`
1. Setup a clean kubernetes cluster (v1.19.8 is **recommended**) using [kind](https://kind.sigs.k8s.io/), [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) or [kk](https://github.com/kubesphere/kubekey).
    * `kk create cluster --with-kubernetes v1.19.8`
2. Install the tekton pipeline using helm.
    * `helm repo add cdf https://cdfoundation.github.io/tekton-helm-chart`
    * `helm install tekton cdf/tekton-pipeline`
3. Install Pipeline and PipelineRun CRDs to the cluster using `make` command at the root directory of `ks-devops:tekton-support`.
    * `make install`
4. Set up the crd controllers.
    * `go run cmd/controller/main.go --pipeline-backend=Tekton`
5. Create a sample pipeline and pipelinerun using `kubectl`.
    * `kubectl apply -f config/samples/devops_v2alpha1_pipeline.yaml`
        * This pipeline called `hello-world` consists of two tasks, including echoing hello world and date, and printing the current directory.
    * `kubectl apply -f config/samples/devops_v2alpha1_pipelinerun.yaml`
        * This pipelinerun will run the above pipeline `hello-world` with a pipelinerun resource named `run-test-demo`.
6. Check the pipeline and pipelinerun status by tekton cli tool [tkn](https://github.com/tektoncd/cli).
    * Check pipeline
        * `tkn pipeline describe hello-world`
    * Check task definitions
        * `tkn task describe hello-world-echo-hello-world`
        * `tkn task describe hello-world-list-directory`
    * Check pipelinerun status
        * `tkn pipelinerun list`
    * Check details of the sample pipelinerun
        * `tkn pipelinerun logs run-test-demo`
7. Delete the pipeline and pipelinerun by `kubectl`
    * `kubectl delete -f config/samples/devops_v2alpha1_pipeline.yaml`
    * `kubectl apply -f config/samples/devops_v2alpha1_pipelinerun.yaml`
8. Check the deletion of tekton resources
    * `tkn task list`
    * `tkn pipeline list`
    * `tkn pipelinerun list`
    * The above command will show no resources if deletion operation is done.


## Design

* To simplify the usage of [Tekton](https://github.com/tektoncd/pipeline) in [ks-devops](https://github.com/kubesphere/ks-devops), a new syntax for using devops referring to [Github Actions](https://docs.github.com/en/actions) is to be designed.
* New versions `v2alpha1` of `Pipeline` and `PipelineRun` CRDs are provided to support the devops syntax.
* Controllers of above two CRDs are implemented to transform our resources to Tekton resources such as Task, TaskRun, Pipeline, PipelineRun, and so on.


### Syntax design

Here, I will describe the example of the new devops syntax first to give you a blueprint of it.

#### Overview 

* A pipeline is composed of several tasks, and each task consists of plenty of steps.
* A pipeline can specify the executing order of tasks.

```bash
┌─────────────────────────────────┐   ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐
│         Pipeline                │                          
│  ┌─────────────────────────┐    │   │name: "pipeline name"│
│  │      Task               │    │                          
│  │ ┌─────────────────────┐ │    │   │tasks:               │
│  │ │    Step             │ │    │                          
│  │ └─────────────────────┘ │──┐ │   │  - name: "task1"    │
│  │ ┌─────────────────────┐ │  │ │        steps:            
│  │ │    Step             │ │  │ │   │      - step1        │
│  │ └─────────────────────┘ │  │ │          - step2         
│  └─────────────────────────┘  │ │   │                     │
│  ┌─────────────────────────┐  │ │                          
│  │      Task               │  │ │   │                     │
│  │ ┌─────────────────────┐ │  │ │      - name: "task2"     
│  │ │    Step             │ │  │ │   │    need:            │
│  │ └─────────────────────┘ │◀─┘ │          - task1         
│  │ ┌─────────────────────┐ │    │   │    steps:           │
│  │ │    Step             │ │    │          - step1         
│  │ └─────────────────────┘ │    │   │      - step2        │
│  └─────────────────────────┘    │                          
└─────────────────────────────────┘   └ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘                              
```

#### Details of syntax

The detailed syntax can be describe by the following yaml file:
```yaml
name: "pipeline name" # required, description: specify the pipeline name
tasks: # required, description: tasks contains a lot of task
  - name: "task name 1" # required, description: specify the task name
    need: [""] # optional, description: names of tasks which must be done before the current task
    uses: "template name of task" # optional, description: the name of template task
    with: # optional, description: params passed to the template task
      - key: value
    env: # optional, description: environment variable settings
      - env_key: env_value
    steps: # optional, description: for self defined task steps
      - name: "step name 1" # required, description: the name of the step
        image: "gcr.io/image:tag" # optional, default: ubuntu, description: the image url of the container
        workspace: "/path/to/execute/scripts" # optional, description: set the workspace of the step
        command: ["echo"] # optional, description: entry point of the container image
        args: ["hello", "step 1"] # optional, description: arguments of the entry point
        script: | # optional, description: command line scripts
          # command line scripts
          # e.g.
          ls
          echo "hello step 1"
```

#### Use cases

##### hello world
 
The following pipeline will print `hello world`.

```yaml
name: "Hello world"
tasks:
  - name: "echo-hello-world-task"
    steps:
      - name: "echo-hello-world-step"
        image: "ubuntu:16.04"
        command: ["/bin/bash"]
        args: ["-c", "echo hello world"]
```

##### Using git task template

This example is about how to use predefined task template.

```yaml
name: "Git clone"
tasks:
  - name: "git-clone-task"
    uses: "git-cli"
    with:
      - name: "GIT_URL"
        value: "https://github.com/kubesphere/ks-devops"
      - name: "GIT_REVISION"
        value: "master"
      - name: "GIT_SECRET"
        value: "${GIT_TOKEN}"
    env:
      - name: "GIT_TOKEN"
        value: "token"
```
-----

Depending on the above examples, we can briefly describe the usage of an image-based task and a template-based task.
* We can choose only one format of task to work in each task.
  * If we decide to use an image-based task, the keyword `steps` and its content are needed, just like the example `hello world`.
  * If we decide to use a template-based task, the keyword `uses` and `with` are needed. (See the example of `git clone`)
    * `uses` provides the template task name.
    * `with` provides the params needed to pass to the template.

### Controllers implementation

#### Pipeline controller

* `Pipeline` controller is located at `controllers/tekton/pipeline/pipeline_controller.go`.
* The controller is registered at `cmd/controller/app/controllers.go` in `addControllers` function.
* The core logic of reconcile Pipeline crd resources is as follows:
  1. Get the requested Pipeline crd resources.
  2. Process the finalizer of it.
  3. Create Tekton Tasks and Pipelines based on the below rules:
      1. Tekton Task name is composed of Pipeline name and task name, concatenating with a `-` symbol, which can avoid the conflicts of task names in different pipelines.
      2. Match the keyword fields of our Pipeline with Tekton Pipeline and Task.
      3. Create Tekton resources only if the resource does not exist.

#### PipelineRun controller

* `PipelineRun` controller is located at `controllers/tekton/pipelinerun/pipelinerun_controller.go`.
* The controller is registered at `cmd/controller/app/controllers.go` in `addControllers` function.
* The core logic of reconcile PipelineRun crd resources is as follows:
  1. Get the requested Pipeline crd resources.
  2. Process the finalizer of it.
  3. Create Tekton PipelineRun based on the below rules:
      1. Create Tekton PipelineRun only if it does not exist.
