# Image URL to use all building/pushing image targets
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := dev-$(shell git describe --tags $(shell git rev-list --tags --max-count=1))

CONTROLLER_IMG ?= surenpi/devops-controller:$(VERSION)-$(COMMIT)
APISERVER_IMG ?= surenpi/devops-apiserver:$(VERSION)-$(COMMIT)
TOOLS_IMG ?= surenpi/devops-tools:$(VERSION)-$(COMMIT)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

GV="devops.kubesphere.io:v1alpha1 devops.kubesphere.io:v1alpha3"

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.21

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: manager

# Run tests
test: manifests generate fmt vet envtest# generate manifests
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile coverage.out

# Build manager binary
manager: generate fmt vet
	go build -a -o bin/controller-manager cmd/controller/main.go

tools-jwt: fmt vet
	go build -a -o bin/jwt cmd/tools/jwt/jwt_cmd.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run cmd/controller/main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${CONTROLLER_IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/api/..."

clientset:
	./hack/generate_client.sh ${GV}

openapi:
	openapi-gen -O openapi_generated -i ./api/v1alpha1 -p kubesphere.io/api/devops/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/violation_exceptions.list

generate-listers:
	lister-gen -v=2 --output-base=. --input-dirs kubesphere.io/devops/pkg/api/devops/v1alpha3,kubesphere.io/devops/pkg/api/devops/v1alpha1  \
 		--output-package pkg/client/listers -h hack/boilerplate.go.txt

# Build the docker image of controller-manager
docker-build-controller:
	docker build --build-arg GOPROXY=https://goproxy.io . -f config/dockerfiles/controller-manager/Dockerfile -t ${CONTROLLER_IMG}

# Push the docker image of controller-manager
docker-push-controller:
	docker push ${CONTROLLER_IMG}

# Build and push the docker image
docker-build-push-controller: docker-build-controller docker-push-controller

# Build the docker image of apiserver
docker-build-apiserver:
	docker build --build-arg GOPROXY=https://goproxy.io . -f config/dockerfiles/apiserver/Dockerfile -t ${APISERVER_IMG}

# Push the docker image of controller-manager
docker-push-apiserver:
	docker push ${APISERVER_IMG}

# Build and push the docker image
docker-build-push-apiserver: docker-build-apiserver docker-push-apiserver

# Build the docker image of apiserver
docker-build-tools:
	docker build . -f config/dockerfiles/tools/Dockerfile -t ${TOOLS_IMG}

# Push the docker image of controller-manager
docker-push-tools:
	docker push ${TOOLS_IMG}

# Build and push the docker image
docker-build-push-tools: docker-build-tools docker-push-tools

docker-build-push: docker-build-push-apiserver docker-build-push-controller

mock-gen:
	mockgen -source=cmd/tools/jwt/app/configmap_updater.go -destination ./cmd/tools/jwt/app/mock_app/configmap_updater.go
	mockgen -source=cmd/tools/jwt/app/kubernetes.go -destination ./cmd/tools/jwt/app/mock_app/kubernetes.go

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

ENVTEST = $(shell pwd)/bin/setup-envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
