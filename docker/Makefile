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

# Global arguments
IMG ?= kubestellar
NAME ?= kubestellar
PLATFORMS ?= linux/amd64,linux/arm64,linux/ppc64le # kcp does not support linux/s390x
KUBESTELLAR_STABLE := $(shell curl -sL https://raw.githubusercontent.com/kubestellar/kubestellar/main/VERSION | grep stable | sed -n -e 's/^.*=//p')
KUBESTELLAR_LATEST := $(shell curl -sL https://raw.githubusercontent.com/kubestellar/kubestellar/main/VERSION | grep latest | sed -n -e 's/^.*=//p')
KUBESTELLAR_VERSION ?= ${KUBESTELLAR_STABLE}
TAG ?= ${KUBESTELLAR_VERSION}
BASEPATH ?= ${HOME}


# Build a local container image for the host architecture
.PHONY: build
build:
    $(info Building the KubeStellar container using the Dockerfile...)
    $(info Arguments: KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION})
    $(info Arguments: IMG=${IMG})
    $(info Arguments: TAG=${TAG})
    docker build --build-arg="KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION}" --tag ${IMG}:${TAG} .


# Build and push a multi-arch container image (docker buildx is required)
.PHONY: buildx
buildx:
    $(info Building the KubeStellar container using Dockerfile...)
    $(info Arguments: KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION})
    $(info Arguments: IMG=${IMG})
    $(info Arguments: TAG=${TAG})
    $(info Arguments: PLATFORMS=${PLATFORMS})
    - docker buildx rm kubestellar-builder
    docker buildx create --name kubestellar-builder --use
    docker buildx build --push --platform ${PLATFORMS} --build-arg="KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION}" --tag ${IMG}:${TAG} .
    - docker buildx rm kubestellar-builder


# Run a new container image
.PHONY: run
run:
    $(info Running a new container container...)
    $(info Arguments: IMG=${IMG})
    $(info Arguments: TAG=${TAG})
    $(info Arguments: NAME=${NAME})
    $(info Arguments: BASEPATH=${BASEPATH})
    mkdir -p ${BASEPATH}/.kcp ${BASEPATH}/kubestellar-logs
        # -p 6443:6443 \
    docker run -d \
        -v ${BASEPATH}/.kcp:/.kcp \
        -v ${BASEPATH}/kubestellar-logs:/kubestellar-logs \
        --user $(shell id -u) \
        --name ${NAME} ${IMG}:${TAG}
    docker cp ${NAME}:/kcp-plugins ${BASEPATH}/kcp
    docker cp ${NAME}:/kubestellar ${BASEPATH}/kubestellar
    $(info export KUBECONFIG=${BASEPATH}/.kcp/admin.kubeconfig)
    $(info export PATH=$$PATH:${BASEPATH}/kcp:${BASEPATH}/kubestellar/bin)


# Login into the running container
.PHONY: exec
exec:
    $(info Logging into a running container...)
    $(info Arguments: NAME=${NAME})
    docker exec -it ${NAME} /bin/bash


# Kill the container, remove the container image and files
.PHONY: stop
clean:
    $(info Arguments: NAME=${NAME})
    - docker rm --force ${NAME}

# Kill the container, remove the container image and files
.PHONY: clean
clean:
    $(info Cleaning up...)
    $(info Arguments: IMG=${IMG})
    $(info Arguments: TAG=${TAG})
    $(info Arguments: BASEPATH=${BASEPATH})
    - docker image rm ${IMG}:${TAG}
    - rm -rf  ${BASEPATH}/.kcp ${BASEPATH}/kcp ${BASEPATH}/kubestellar ${BASEPATH}/kubestellar-logs
