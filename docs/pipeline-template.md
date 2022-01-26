# Pipeline Template Design

## Background

At present, our Pipeline template includes two pre-defined templates and one user-defined template. The two pre-defined
templates are defined by the console and can hardly be changed. The user-defined template is equivalent to re-editing an
empty Pipeline template, completely losing the function of template - REUSE. Even though we now have the function of
copying Pipeline, but we still can't solve the problem of Pipeline template:

- The Pipeline template should be managed, like adding, modifying and deleting.
- The Pipeline template shall have flexible input parameters, and system can generate the corresponding Pipeline
  execution definition, Jenkinsfile or others, according to the input and template.

### Use cases

- Official pre-defined template

  Users can easily use the official pre-defined templates, such as build of Maven project, to easily build their own
  projects without having to configure and debug from scratch. Of course, we need to provide enough flexible parameters
  to meet customized requirements, such as the URL of source code.

- User-defined template

  Each company has its own unique Pipelines, but the Pipeline configuration of projects within the company may be
  similar. Users can customize their own Pipeline template to improve the efficiency of the team.

### Goals

- Template CRD is provided to allow users to add, modify and delete Pipeline templates by themselves.
- Implement admission webhook to validate Template CRs.
- Improve Pipeline CRD and provide template associated fields.
- Improve Pipeline reconciler and automatically generate Jenkinsfile defined in PipelineSpec.
- Provide rich official pre-defined templates.

### Non-Goals

- Provide Cluster wide Pipeline template. We could implement this feature in the future.
- Provide Template version management. Version management is too complex and difficult to implement currently. Do we
  really need it?
- Use [CEL](https://github.com/google/cel-spec) to validate Template parameters. Due to the need to introduce new
  dependencies, simple validation can be realized.

## Design

We are going to implement Pipeline template by creating a CRD. The kind of CRD is `Template`, and the version of CRD
is `devops.kubesphere.io/v1alpha1`. The following is an example of the corresponding CR:

```yaml
apiVersion: devops.kubesphere.io/v1alpha1
kind: Template
metadata:
  name: my-template
  namespace: my-devops-project
  annotations:
    devops.kubesphere.io/categories: Gradle Project deploy
    devops.kubesphere.io/tags: backend, gradle, java, docker
    devops.kubesphere.io/displayName: Gradle Test, Build and Deploy
spec:
  parameters:
    - name: gitCloneURL
      description: What is your repository URL you want to clone?
      type: string # ignorable
      validation:
        expression: "matches()"
        message: "Please input a correct URL."
    - name: revision
      description: Which revision do you want to clone from?
      default: "main" # Valid JSON value
    - name: buildOnly
      description: Do we really need build stage only?
      default: false # Valid JSON value
      type: bool
    - name: matrix
      description: Matrix versions of gradle
      type: string-array
      default: [ "6.9.1-jdk11", "7.0.0-jdk11", "7.3.3-jdk11" ]

  template: | # Written in Go template
    pipeline {
        agent {
            kubernetes {
            inheritFrom 'gradle'
                containerTemplate {
                    name 'gradle'
                    image 'gradle:7.3.3-jdk11'
                }
            }
        }
        stages {
            stage('Checkout') {
                steps {
                    checkout poll: false, scm: [$class: 'GitSCM', branches: [[name: '*/master']], extensions: [[$class: 'CloneOption', depth: 1, noTags: true, reference: '', shallow: true], [$class: 'SubmoduleOption', depth: 1, disableSubmodules: false, parentCredentials: false, recursiveSubmodules: true, reference: '', shallow: true, trackingSubmodules: false]], userRemoteConfigs: [[url: '{{ .params.gitCloneURL }}']]]
                }
            }
            {{if not .buildOnly}}

            stage('Gradle Check') {
                steps {
                    container('gradle') {
                        sh 'gradle check'
                    }
                }
            }

            stage('Gradle Build') {
                steps {
                    container('gradle') {
                        sh 'gradle build -x test'
                    }
                }
            }

            stage('Archive Assets') {
                steps {
                    archiveArtifacts '**/build/libs/*.jar'
                }
            }
        }
    }
```

### Parameter Definition

| Field       | Type       | Description                                                                                                                     | Default Value |
|-------------|------------|---------------------------------------------------------------------------------------------------------------------------------|---------------|
| name        | string     | Name of parameter. The name needs to conform to the [go template specification](https://pkg.go.dev/text/template#hdr-Arguments) | -             |
| description | string     | Description of the parameter                                                                                                    | ""            |
| default     | json.Value | Default value of the parameter. If the default value is set, this parameter is optional; otherwise, the parameter is required   | nil           |
| type        | string     | Type of the parameter. We look forwad to supporting more types in the future                                                    | string        |
| validation  | Validation | The validation configuration of the parameter includes validation expression and message                                        | nil           |

### Validation Definition

| Field      | Type   | Description                                                                                        | Default Value |
|------------|--------|----------------------------------------------------------------------------------------------------|---------------|
| expression | string | The expression of the validation. Expect to follow [CEL spec](https://github.com/google/cel-spec)。 | -             |
| message    | string | Message given after validation failure.                                                            | -             |

### Pipeline CRD Improvement

```yaml
apiVersion: devops.kubesphere.io/v1alpha4
kind: Pipeline
metadata:
  name: my-pipeline
  namespace: my-devops-project
spec:
  template:
    ref:
      kind: Template # Available kind: Template(ignorable), ClusterTemplate
      name: "my-template"
      # Default namespace is equal to the Pipeline
    parameters:
      - name: gitCloneURL
        value: "https://github.com/halo-dev/halo"
      - name: revision
        value: release
      - name: buildOnly
        value: true
      - name: matrix
        value: [ "7.3.3-jdk11" ]
        # Or value: 7.3.3-jdk11
  pipeline:
    jenkinsfile: |
      pipeline {}
```

## Restrictions

- We cannot edit Pipeline template in the graphical interface directly.

  But we can render the template and show it in the graphical interface. Just like Markdown preview.

- Once a template is associated, it cannot be modified or deleted.

  But a template can be modified or deleted if no Pipeline associated.

- We can only edit Pipeline template using Go template specification.
