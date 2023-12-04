# Copyright 2021 The KubeStellar Authors.
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

# We need bash for some conditional logic below.
# test PR

SHELL := /usr/bin/env bash -e

GIT_COMMIT=
GIT_DIRTY=
GIT_VERSION=

KUBE_MAJOR_VERSION := $(shell go mod edit -json | jq '.Require[] | select(.Path == "k8s.io/kubernetes") | .Version' --raw-output | sed 's/v\([0-9]*\).*/\1/')
KUBE_MINOR_VERSION := $(shell go mod edit -json | jq '.Require[] | select(.Path == "k8s.io/kubernetes") | .Version' --raw-output | sed "s/v[0-9]*\.\([0-9]*\).*/\1/")
CORE_PLATFORMS ?= linux/amd64,linux/arm64,linux/ppc64le # todo linux/s390x  ?
CORE_IMAGE_REPO ?= quay.io/kubestellar/space-framework

GIT_INFO_FILE := git-info.txt

BUILD_TIME_TAG := $(shell date -u +b%y-%m-%d-%H-%M-%S)
GIT_TAG = git-${GIT_COMMIT}-${GIT_DIRTY}

GO_INSTALL = ./hack/go-install.sh

TOOLS_DIR=hack/tools
TOOLS_GOBIN_DIR := $(abspath $(TOOLS_DIR))
GOBIN_DIR=$(abspath ./bin )
PATH := $(GOBIN_DIR):$(TOOLS_GOBIN_DIR):$(PATH)
TMPDIR := $(shell mktemp -d)

# Detect the path used for the install target
ifeq (,$(shell go env GOBIN))
INSTALL_GOBIN=$(shell go env GOPATH)/bin
else
INSTALL_GOBIN=$(shell go env GOBIN)
endif

CONTROLLER_GEN_VER := v0.10.0
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER)
export CONTROLLER_GEN # so hack scripts can use it

YAML_PATCH_VER ?= v0.0.11
YAML_PATCH_BIN := yaml-patch
YAML_PATCH := $(TOOLS_DIR)/$(YAML_PATCH_BIN)-$(YAML_PATCH_VER)
export YAML_PATCH # so hack scripts can use it

OPENSHIFT_GOIMPORTS_VER := 4cd858e694d7dfa32a2e697e0e4bab245c215cf3
OPENSHIFT_GOIMPORTS_BIN := openshift-goimports
OPENSHIFT_GOIMPORTS := $(TOOLS_DIR)/$(OPENSHIFT_GOIMPORTS_BIN)-$(OPENSHIFT_GOIMPORTS_VER)
export OPENSHIFT_GOIMPORTS # so hack scripts can use it

GOLANGCI_LINT_VER := v1.50.1
GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(TOOLS_GOBIN_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER)

STATICCHECK_VER := 2022.1
STATICCHECK_BIN := staticcheck
STATICCHECK := $(TOOLS_GOBIN_DIR)/$(STATICCHECK_BIN)-$(STATICCHECK_VER)

GOTESTSUM_VER := v1.8.1
GOTESTSUM_BIN := gotestsum
GOTESTSUM := $(abspath $(TOOLS_DIR))/$(GOTESTSUM_BIN)-$(GOTESTSUM_VER)

LOGCHECK_VER := v0.4.1
LOGCHECK_BIN := logcheck
LOGCHECK := $(TOOLS_GOBIN_DIR)/$(LOGCHECK_BIN)-$(LOGCHECK_VER)
export LOGCHECK # so hack scripts can use it

ARCH := $(shell go env GOARCH)
OS := $(shell go env GOOS)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

ifeq ($(IMPORT_TAGS),yes)
	include ${GIT_INFO_FILE}
else
	GIT_COMMIT=$(shell git rev-parse --short HEAD || echo 'local')
	GIT_DIRTY=$(shell [ $$(git status --porcelain=v2 | wc -l) == 0 ] && echo 'clean' || echo 'dirty') 
	GIT_VERSION=$(shell go mod edit -json | jq '.Require[] | select(.Path == "k8s.io/kubernetes") | .Version' --raw-output)+kcp-$(shell git describe --tags --match='v*' --abbrev=14 "$(GIT_COMMIT)^{commit}" 2>/dev/null || echo v0.0.0-$(GIT_COMMIT)) 
endif

LDFLAGS := \
	-X k8s.io/client-go/pkg/version.gitCommit=${GIT_COMMIT} \
	-X k8s.io/client-go/pkg/version.gitTreeState=${GIT_DIRTY} \
	-X k8s.io/client-go/pkg/version.gitVersion=${GIT_VERSION} \
	-X k8s.io/client-go/pkg/version.gitMajor=${KUBE_MAJOR_VERSION} \
	-X k8s.io/client-go/pkg/version.gitMinor=${KUBE_MINOR_VERSION} \
	-X k8s.io/client-go/pkg/version.buildDate=${BUILD_DATE} \
	\
	-X k8s.io/component-base/version.gitCommit=${GIT_COMMIT} \
	-X k8s.io/component-base/version.gitTreeState=${GIT_DIRTY} \
	-X k8s.io/component-base/version.gitVersion=${GIT_VERSION} \
	-X k8s.io/component-base/version.gitMajor=${KUBE_MAJOR_VERSION} \
	-X k8s.io/component-base/version.gitMinor=${KUBE_MINOR_VERSION} \
	-X k8s.io/component-base/version.buildDate=${BUILD_DATE} \
	-extldflags '-static'

.PHONY: all
all: build

.PHONY: bin-dir
bin-dir:
	mkdir -p bin

.PHONY: build
build: WHAT ?= ./cmd/space-manager ./cmd/space-client-test
build: bin-dir 
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build $(BUILDFLAGS) -ldflags="$(LDFLAGS)" -o bin $(WHAT)

ldflags:
	@echo $(LDFLAGS)

.PHONY: require-%
require-%:
	@if ! command -v $* 1> /dev/null 2>&1; then echo "$* not found in \$$PATH"; exit 1; fi

#$(CODE_GENERATOR):
#	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) github.com/kcp-dev/code-generator/v2 $(CODE_GENERATOR_BIN) $(CODE_GENERATOR_VER)

tools: $(GOLANGCI_LINT) $(CONTROLLER_GEN) $(YAML_PATCH) $(GOTESTSUM) $(OPENSHIFT_GOIMPORTS) 
.PHONY: tools

$(CONTROLLER_GEN):
	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) sigs.k8s.io/controller-tools/cmd/controller-gen $(CONTROLLER_GEN_BIN) $(CONTROLLER_GEN_VER)

$(YAML_PATCH):
	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) github.com/pivotal-cf/yaml-patch/cmd/yaml-patch $(YAML_PATCH_BIN) $(YAML_PATCH_VER)

$(GOTESTSUM):
	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) gotest.tools/gotestsum $(GOTESTSUM_BIN) $(GOTESTSUM_VER)

crds: $(CONTROLLER_GEN) $(YAML_PATCH)
	./hack/update-codegen-crds.sh
.PHONY: crds

codegen: crds 
	go mod download
	./hack/update-codegen-clients.sh
	$(MAKE) imports
.PHONY: codegen

codegenclients:  $(CONTROLLER_GEN)
	go mod download
	./hack/update-codegen-clients.sh
	$(MAKE) imports
.PHONY: codegenclients


$(OPENSHIFT_GOIMPORTS):
	GOBIN=$(TOOLS_GOBIN_DIR) $(GO_INSTALL) github.com/openshift-eng/openshift-goimports $(OPENSHIFT_GOIMPORTS_BIN) $(OPENSHIFT_GOIMPORTS_VER)

$(TOOLS_DIR)/verify_boilerplate.py:
	mkdir -p $(TOOLS_DIR)
	curl --fail --retry 3 -L -o $(TOOLS_DIR)/verify_boilerplate.py https://raw.githubusercontent.com/kubernetes/repo-infra/master/hack/verify_boilerplate.py
	chmod +x $(TOOLS_DIR)/verify_boilerplate.py

.PHONY: verify-boilerplate
verify-boilerplate: $(TOOLS_DIR)/verify_boilerplate.py
	$(TOOLS_DIR)/verify_boilerplate.py --boilerplate-dir=hack/boilerplate --skip docs

.PHONY: imports
imports: $(OPENSHIFT_GOIMPORTS) 
	$(OPENSHIFT_GOIMPORTS) -m github.com/kubestellar/kubestellar

.PHONY: modules
modules: ## Run go mod tidy to ensure modules are up to date
	go mod tidy

.PHONY: spacecore-image
spacecore-image: export-git-info
	if ! docker buildx inspect space-framework &> /dev/null; then docker buildx create --name space-framework --platform $(CORE_PLATFORMS); fi
	if [ -n "$(EXTRA_CORE_TAG)" ]; then extra="--tag $(CORE_IMAGE_REPO):$(EXTRA_CORE_TAG)"; else extra=""; fi; eval docker buildx --builder space-framework build --push --platform $(CORE_PLATFORMS) --tag $(CORE_IMAGE_REPO):$(BUILD_TIME_TAG) --tag $(CORE_IMAGE_REPO):$(GIT_TAG) $$extra -f core.Dockerfile .

.PHONY: spacecore-image-local
spacecore-image-local: export-git-info
	if [ -n "$(EXTRA_CORE_TAG)" ]; then extra="--tag $(CORE_IMAGE_REPO):$(EXTRA_CORE_TAG)"; else extra=""; fi; case "$$HOSTTYPE" in (aarch64*|arm64*) extoo="--platform linux/arm64";; esac; eval docker build --tag $(CORE_IMAGE_REPO):$(BUILD_TIME_TAG) --tag $(CORE_IMAGE_REPO):$(GIT_TAG) $$extra -f core.Dockerfile $$extoo .

.PHONY: export-git-info
export-git-info:
	echo  "GIT_COMMIT := " "$(GIT_COMMIT)" > ${GIT_INFO_FILE}
	echo  "GIT_DIRTY := " ${GIT_DIRTY} >> ${GIT_INFO_FILE}
	echo  "GIT_VERSION := " ${GIT_VERSION} >> ${GIT_INFO_FILE}


