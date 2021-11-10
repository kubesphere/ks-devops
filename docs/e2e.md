# E2E testing Guide

> SkyWalking Infra E2E is the next generation End-to-End Testing framework that aims to help developers to set up, debug, and verify E2E tests with ease. It’s built based on the lessons learnt from tens of hundreds of test cases in the SkyWalking main repo.

Currently, we implement our E2E testing by using [SkyWalking Infra E2E](https://github.com/apache/skywalking-infra-e2e). If you want to learn more, please refer to it's [documentation website](https://skywalking.apache.org/docs/skywalking-infra-e2e/latest/readme/).

## The Structure of the E2E Cases

```bash
.github
└── workflows
    ├── e2e.install.yaml # E2E testing workflow
    └── e2e.others.yaml  # E2E testing workflow

test
└── e2e
    ├── cases
    │   │── chart-install     # E2E testing instance
    │   │   ├── e2e.yaml      # E2E testing configuration file
    │   │   └── expected.yaml # E2# testing expected template
    │   │── other-cases
    │   │   ├── e2e.yaml
    │   │   └── expected.yaml
    └── common # common folder for E2E testing
        ├── kind-1.19.yaml
        ├── kind-1.20.yaml
        ├── kind-1.21.yaml
        └── kind-1.22.yaml
```

## FAQ

1. Do I need to add E2E testing cases for any new features?

   If you have developed a **BIG** new feature, it would be better to add a new E2E testing case. There are some steps as show below:

   - Create new E2E testing case `e2e.yaml` and expected template `expected.yaml` into `test/e2e/cases/new-feature`
   - Create E2E testing workflow `e2e.new-feature.yaml` into `.github/workflows`

2. How to add a new Kubernetes version for matrix E2E testing?

   At present, Kubernetes version we have support are v1.19, 1.20, 1.21 and 1.22.

   If you want to add a new Kubernetes version for matrix test, there are two steps as shown below:

   - Create new KinD configuration file `kind-1.xy.yaml` into `test/e2e/common` folder
   - Change all `.github/workflows/e2e.*.yaml` workflows which support matrix strategy, e.g.:

     ```yaml
     jobs:
       Install:
       name: Chart Install
       runs-on: ubuntu-20.04
       timeout-minutes: 60
       strategy:
       matrix:
         k8sVersion: ["1.xy", "1.19", "1.20", "1.21", "1.22"]
     ```

3. How to write a E2E configuration?

   Please refer to <https://skywalking.apache.org/docs/skywalking-infra-e2e/latest/en/setup/configuration-file/>.

4. How to debug E2E configuration in local evnironment?

   - Download [latest e2e tool](https://www.apache.org/dyn/closer.cgi/skywalking/infra-e2e) or run commond `go run cmd/main.go` after cloning [source code](https://github.com/apache/skywalking-infra-e2e)
   - Then refer to <https://skywalking.apache.org/docs/skywalking-infra-e2e/latest/en/setup/run-e2e-tests/>.
