# Copyright 2023 The KubeStellar Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Image repo/tag to use all building/pushing image targets
DOCKER_REGISTRY ?= ghcr.io/kubestellar/kubestellar
CMD_NAME ?= controller-manager
IMAGE_TAG ?= $(shell git rev-parse --short HEAD)
IMAGE_REPO ?= ${DOCKER_REGISTRY}/${CMD_NAME}
IMG ?= ${IMAGE_REPO}:${IMAGE_TAG}

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26.1
# Default Namespace to use for make deploy (mainly for local testing)
DEFAULT_NAMESPACE=default
# Default WDS name to use for make deploy (mainly for local testing)
DEFAULT_WDS_NAME=wds1
# default kind hosting cluster name
KIND_HOSTING_CLUSTER ?= kubeflex

# We need bash for some conditional logic below.
SHELL := /usr/bin/env bash -e

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ARCH := $(shell go env GOARCH)
OS := $(shell go env GOOS)

TOOLS_DIR=hack/tools
TOOLS_GOBIN_DIR := $(abspath $(TOOLS_DIR))
GOBIN_DIR=$(abspath ./bin )
PATH := $(GOBIN_DIR):$(TOOLS_GOBIN_DIR):$(PATH)
TMPDIR := $(shell mktemp -d)
GO_INSTALL = ./hack/go-install.sh

# Detect the path used for the install target
ifeq (,$(shell go env GOBIN))
INSTALL_GOBIN=$(shell go env GOPATH)/bin
else
INSTALL_GOBIN=$(shell go env GOBIN)
endif

GOLANGCI_LINT_VER := v1.50.1
GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(TOOLS_GOBIN_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER)

STATICCHECK_VER := 2022.1
STATICCHECK_BIN := staticcheck
STATICCHECK := $(TOOLS_GOBIN_DIR)/$(STATICCHECK_BIN)-$(STATICCHECK_VER)

OPENSHIFT_GOIMPORTS_VER := 4cd858e694d7dfa32a2e697e0e4bab245c215cf3
OPENSHIFT_GOIMPORTS_BIN := openshift-goimports
OPENSHIFT_GOIMPORTS := $(TOOLS_DIR)/$(OPENSHIFT_GOIMPORTS_BIN)-$(OPENSHIFT_GOIMPORTS_VER)
export OPENSHIFT_GOIMPORTS # so hack scripts can use it

KUBE_CLIENT_MAJOR_VERSION := $(shell go mod edit -json | jq '.Require[] | select(.Path == "k8s.io/client-go") | .Version' --raw-output | sed 's/v\([0-9]*\).*/\1/')
KUBE_CLIENT_MINOR_VERSION := $(shell go mod edit -json | jq '.Require[] | select(.Path == "k8s.io/client-go") | .Version' --raw-output | sed "s/v[0-9]*\.\([0-9]*\).*/\1/")
GIT_COMMIT := $(shell git rev-parse --short HEAD || echo 'local')
GIT_DIRTY := $(shell git diff --quiet && echo 'clean' || echo 'dirty')
GIT_VERSION := $(shell go mod edit -json | jq '.Require[] | select(.Path == "k8s.io/client-go") | .Version' --raw-output)+kflex-$(shell git describe --tags --match='v*' --abbrev=14 "$(GIT_COMMIT)^{commit}" 2>/dev/null || echo v0.0.0-$(GIT_COMMIT))
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
MAIN_VERSION := $(shell git tag -l --sort=-v:refname | head -n1)
LDFLAGS := \
	-X main.Version=${MAIN_VERSION}.${GIT_COMMIT} \
	-X main.BuildDate=${BUILD_DATE} \
	-X k8s.io/client-go/pkg/version.gitCommit=${GIT_COMMIT} \
	-X k8s.io/client-go/pkg/version.gitTreeState=${GIT_DIRTY} \
	-X k8s.io/client-go/pkg/version.gitVersion=${GIT_VERSION} \
	-X k8s.io/client-go/pkg/version.gitMajor=${KUBE_CLIENT_MAJOR_VERSION} \
	-X k8s.io/client-go/pkg/version.gitMinor=${KUBE_CLIENT_MINOR_VERSION} \
	-X k8s.io/client-go/pkg/version.buildDate=${BUILD_DATE} \
	\
	-X k8s.io/component-base/version.gitCommit=${GIT_COMMIT} \
	-X k8s.io/component-base/version.gitTreeState=${GIT_DIRTY} \
	-X k8s.io/component-base/version.gitVersion=${GIT_VERSION} \
	-X k8s.io/component-base/version.gitMajor=${KUBE_CLIENT_MAJOR_VERSION} \
	-X k8s.io/component-base/version.gitMinor=${KUBE_CLIENT_MINOR_VERSION} \
	-X k8s.io/component-base/version.buildDate=${BUILD_DATE} \
	-extldflags '-static'
all: build
.PHONY: all

ldflags:
	@echo $(LDFLAGS)

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development Dependencies

CODE_GEN_VER := v0.28.2
CODE_GEN_DIR := $(TOOLS_DIR)/code-generator-clone-$(CODE_GEN_VER)
export CODE_GEN_DIR

$(CODE_GEN_DIR):
	git clone -b $(CODE_GEN_VER) --depth 1 https://github.com/kubernetes/code-generator.git $(CODE_GEN_DIR)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./api/..." output:crd:artifacts:config=config/crd/bases
	cp config/crd/bases/* pkg/crd/files

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate/boilerplate.go.txt" paths="./api/..."

.PHONY: codegenclients
codegenclients: $(CODE_GEN_DIR)
	./hack/update-codegen-clients.sh
	$(MAKE) imports

.PHONY: all-generated
all-generated: manifests generate codegenclients

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet ## Run tests.
	go test ./api/... ./cmd/... ./pkg/... -coverprofile cover.out

##@ Build

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/controller-manager/main.go $(ARGS)

.PHONY: ko-build-local
ko-build-local: ## Build local container image with ko
	$(shell (docker version | { ! grep -qi podman; } ) || echo "DOCKER_HOST=unix://$$HOME/.local/share/containers/podman/machine/qemu/podman.sock ") KO_DOCKER_REPO=ko.local ko build -B ./cmd/${CMD_NAME} -t ${IMAGE_TAG} --platform linux/${ARCH}
	docker tag ko.local/${CMD_NAME}:${IMAGE_TAG} ${IMAGE_REPO}:${IMAGE_TAG}

# this is used for local testing
.PHONY: kind-load-image
kind-load-image:
	kind load --name ${KIND_HOSTING_CLUSTER} docker-image ${IMAGE_REPO}:${IMAGE_TAG}

.PHONY: chart
chart: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(shell echo ${IMG} | sed 's/\(:.*\)v/\1/')
	$(KUSTOMIZE) build config/default > chart/templates/controller-manager.yaml
	scripts/add-helm-code.sh add


.PHONY: local-chart
local-chart: manifests kustomize
ifeq (1,$(shell (git status | grep "config/manager/kustomization.yaml" | wc -l)))
	@echo 'ERROR: config/manager/kustomization.yaml is already checked out!'
	@exit 1
endif
	cp -R chart local-chart
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(shell echo ${IMG} | sed 's/\(:.*\)v/\1/')
	$(KUSTOMIZE) build config/default > local-chart/templates/controller-manager.yaml
	scripts/add-helm-code.sh --dir ${PWD}/local-chart add
	git checkout -- config/manager/kustomization.yaml

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy manager to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | sed -e 's/{{.Release.Namespace}}/${DEFAULT_NAMESPACE}/g' -e 's/{{.Values.ControlPlaneName}}/${DEFAULT_WDS_NAME}/g' | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy manager from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

# installs the chart from ./chart for local dev/testing a WDS using image loaded in kind.
# The Helm chart should be instantiated into the KubeFlex hosting cluster.
# If $(KUBE_CONTEXT) is set then that indicates where to install the chart; otherwise it goes to the current kubeconfig context.
.PHONY: install-local-chart
install-local-chart: local-chart kind-load-image
	helm upgrade $(if $(KUBE_CONTEXT),--kube-context $(KUBE_CONTEXT),) --install kubestellar -n ${DEFAULT_WDS_NAME}-system ./local-chart  --set ControlPlaneName=${DEFAULT_WDS_NAME}

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.13.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@latest

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOTOOLCHAIN=go1.21.8 GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

$(GOLANGCI_LINT):
	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint $(GOLANGCI_LINT_BIN) $(GOLANGCI_LINT_VER)

$(STATICCHECK):
	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) honnef.co/go/tools/cmd/staticcheck $(STATICCHECK_BIN) $(STATICCHECK_VER)

$(LOGCHECK):
	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) sigs.k8s.io/logtools/logcheck $(LOGCHECK_BIN) $(LOGCHECK_VER)

update-contextual-logging: $(LOGCHECK)
	UPDATE=true ./hack/verify-contextual-logging.sh
.PHONY: update-contextual-logging

.PHONY: lint
lint: $(GOLANGCI_LINT) $(STATICCHECK) $(LOGCHECK)
	./hack/verify-contextual-logging.sh

VENVDIR=$(abspath docs/venv)
REQUIREMENTS_TXT=docs/requirements.txt

.PHONY: serve-docs
serve-docs: venv
	. $(VENV)/activate; \
	VENV=$(VENV) REMOTE=$(REMOTE) BRANCH=$(BRANCH) docs/scripts/serve-docs.sh

.PHONY: deploy-docs
deploy-docs: venv
	. $(VENV)/activate; \
	REMOTE=$(REMOTE) BRANCH=$(BRANCH) docs/scripts/deploy-docs.sh

.PHONY: verify-go-versions
verify-go-versions:
	hack/verify-go-versions.sh

$(TOOLS_DIR)/verify_boilerplate.py:
	mkdir -p $(TOOLS_DIR)
	curl --fail --retry 3 -L -o $(TOOLS_DIR)/verify_boilerplate.py https://raw.githubusercontent.com/kubernetes/repo-infra/master/hack/verify_boilerplate.py
	chmod +x $(TOOLS_DIR)/verify_boilerplate.py

.PHONY: verify-boilerplate
verify-boilerplate: $(TOOLS_DIR)/verify_boilerplate.py
	$(TOOLS_DIR)/verify_boilerplate.py --boilerplate-dir=hack/boilerplate --skip docs

# TODO - there is currently no client code generation, need to revisit once that is added back
# this stub is mainly to pass the CI
.PHONY: verify-codegen
verify-codegen:
	./hack/verify-codegen.sh

$(OPENSHIFT_GOIMPORTS):
	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) github.com/openshift-eng/openshift-goimports $(OPENSHIFT_GOIMPORTS_BIN) $(OPENSHIFT_GOIMPORTS_VER)

.PHONY: imports
imports: $(OPENSHIFT_GOIMPORTS) verify-go-versions
	$(OPENSHIFT_GOIMPORTS) -i github.com/kubestellar/kubeflex -m github.com/kubestellar/kubestellar

.PHONY: verify-imports
verify-imports:
	hack/verify-imports.sh

.PHONY: modules
modules: ## Run go mod tidy to ensure modules are up to date
	go mod tidy

.PHONY: verify-modules
verify-modules: modules  ## Verify go modules are up to date
	@if !(git diff --quiet HEAD -- go.sum go.mod); then \
		git diff; \
		echo "go module files are out of date"; exit 1; \
	fi

.PHONY: require-%
require-%:
	@if ! command -v $* 1> /dev/null 2>&1; then echo "$* not found in ${PATH}"; exit 1; fi

.PHONY: build-all
build-all:
	GOOS=$(OS) GOARCH=$(ARCH) $(MAKE) build WHAT='./cmd/...'

.PHONY: build
build: WHAT ?= ./cmd/...
build: bin-dir require-jq require-go require-git verify-go-versions  ## Build the project
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build $(BUILDFLAGS) -ldflags="$(LDFLAGS)" -o bin $(WHAT)

.PHONY: bin-dir
bin-dir:
	mkdir -p bin


include Makefile.venv
