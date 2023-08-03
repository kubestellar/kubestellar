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
EXTERNAL_HOSTNAME ?= localhost
EXTERNAL_PORT ?= 6443


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
	$(info Arguments: EXTERNAL_HOSTNAME=${EXTERNAL_HOSTNAME})
	$(info Arguments: EXTERNAL_PORT=${EXTERNAL_PORT})
	docker run -d \
		-p ${EXTERNAL_PORT}:6443 \
		-e EXTERNAL_HOSTNAME=${EXTERNAL_HOSTNAME} \
		-e EXTERNAL_PORT=${EXTERNAL_PORT} \
		--user $(shell id -u) \
		--name ${NAME} \
		${IMG}:${TAG}
	$(info Waiting for KubeStellar to be ready... it may take a while!)
	until docker logs kubestellar 2>&1 | grep -Fxq "Ready!" ; do sleep 1 ; done
	docker cp ${NAME}:/home/kubestellar/.kcp ${BASEPATH}/.kcp
	docker cp ${NAME}:/home/kubestellar/kcp-plugins ${BASEPATH}/kcp-plugins
	docker cp ${NAME}:/home/kubestellar/kubestellar ${BASEPATH}/kubestellar
	$(info export KUBECONFIG=${BASEPATH}/.kcp/admin.kubeconfig)
	$(info export PATH=$$PATH:${BASEPATH}/kcp-plugins/bin:${BASEPATH}/kubestellar/bin)


# Login into the running container
.PHONY: exec
exec:
	$(info Logging into a running container...)
	$(info Arguments: NAME=${NAME})
	docker exec -it ${NAME} /bin/bash


# Kill the container, remove the container image and files
.PHONY: stop
stop:
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
	- rm -rf  ${BASEPATH}/.kcp ${BASEPATH}/kcp-plugins ${BASEPATH}/kubestellar
