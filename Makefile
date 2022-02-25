# Image URL to use all building/pushing image targets
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := dev-$(shell git describe --tags $(shell git rev-list --tags --max-count=1))

DOCKER_REPO ?= kubespheredev
CONTROLLER_IMG ?= ${DOCKER_REPO}/devops-controller:$(VERSION)-$(COMMIT)
APISERVER_IMG ?= ${DOCKER_REPO}/devops-apiserver:$(VERSION)-$(COMMIT)
TOOLS_IMG ?= ${DOCKER_REPO}/devops-tools:$(VERSION)-$(COMMIT)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

GV="devops.kubesphere.io:v1alpha1 devops.kubesphere.io:v1alpha3 gitops.kubesphere.io:v1alpha1"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: test lint

# Run tests
test: fmt vet # generate manifests
	go test ./... -coverprofile coverage.out

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

install-crd:
	kubectl apply -f config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Install golang-lint via https://golangci-lint.run/usage/install/#local-installation
lint:
	golangci-lint run ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

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

swagger-ui:
	git clone https://github.com/swagger-api/swagger-ui -b v2.2.10 --depth 1 bin/swagger-ui

mock-gen:
	mockgen -source=cmd/tools/jwt/app/configmap_updater.go -destination ./cmd/tools/jwt/app/mock_app/configmap_updater.go
	mockgen -source=cmd/tools/jwt/app/kubernetes.go -destination ./cmd/tools/jwt/app/mock_app/kubernetes.go

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
