# Image URL to use all building/pushing image targets
CONTROLLER_IMG ?= surenpi/devops-controller:8
APISERVER_IMG ?= surenpi/devops-apiserver:8
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

GV="devops:v1alpha1 devops:v1alpha3"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -a -o bin/controller-manager cmd/controller/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run cmd/controller/main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

lint-chart:
	helm lint charts/ks-devops

install-chart: lint-chart
	helm install ks-ctl charts/ks-devops -n kubesphere-devops-system --set serviceAccount.create=true --create-namespace \
		--set image.pullPolicy=Always

install-jenkins-chart:
	helm install ks-jenkins-test charts/ks-devops/charts/jenkins --set Master.NodePort=

render-jenkins-chart:
	helm template ks-jenkins-test charts/ks-devops/charts/jenkins

uninstall-chart:
	make uninstall-jenkins-chart || true
	helm uninstall ks-ctl -n kubesphere-devops-system

uninstall-jenkins-chart:
	helm uninstall ks-jenkins-test

reinstall-chart:
	make uninstall-chart || true
	make install-chart

reinstall-jenkins-chart:
	make uninstall-jenkins-chart || true
	make install-jenkins-chart

package-chart:
	cd charts && helm package ks-devops

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
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

clientset:
	./hack/generate_client.sh ${GV}

openapi:
	openapi-gen -O openapi_generated -i ./api/v1alpha1 -p kubesphere.io/api/devops/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/violation_exceptions.list

# Build the docker image of controller-manager
docker-build-controller:
	docker build . -f config/dockerfiles/controller-manager/Dockerfile -t ${CONTROLLER_IMG}

# Push the docker image of controller-manager
docker-push-controller:
	docker push ${CONTROLLER_IMG}

# Build and push the docker image
docker-build-push-controller: docker-build-controller docker-push-controller

# Build the docker image of apiserver
docker-build-apiserver:
	docker build . -f config/dockerfiles/apiserver/Dockerfile -t ${APISERVER_IMG}

# Push the docker image of controller-manager
docker-push-apiserver:
	docker push ${APISERVER_IMG}

# Build and push the docker image
docker-build-push-apiserver: docker-build-apiserver docker-push-apiserver

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
