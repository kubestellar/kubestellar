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
IMG := kubestellar
NAME := kubestellar
PLATFORMS := linux/amd64,linux/arm64,linux/ppc64le
KUBESTELLAR_VERSION := $(shell curl -sL https://github.com/kubestellar/kubestellar/releases/latest | grep "</h1>" | tail -n 1 | sed -e 's/<[^>]*>//g' | xargs)


# Build a local container image for the host architecture
.PHONY: build
build:
	$(info Building the KubeStellar container using Dockerfile...)
	$(info Arguments: IMG=${IMG})
	$(info Arguments: KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION})
	# docker buildx build --build-arg="KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION}" --tag ${IMG}:${KUBESTELLAR_VERSION} .
	docker build --build-arg="KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION}" --tag ${IMG}:${KUBESTELLAR_VERSION} .


# Build and push a multi-arch container image (docker buildx is required)
.PHONY: buildx
buildx:
	$(info Building the KubeStellar container using Dockerfile...)
	$(info Arguments: IMG=${IMG})
	$(info Arguments: KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION})
	$(info Arguments: PLATFORMS=${PLATFORMS})
	- docker buildx create --name kubestellar-builder --use
	docker buildx build --push --platform ${PLATFORMS} --build-arg="KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION}" --tag ${IMG}:${KUBESTELLAR_VERSION} .
	- docker buildx rm kubestellar-builder


# Run a new container image
.PHONY: run
run:
	$(info Running a new container container...)
	$(info Arguments: IMG=${IMG})
	$(info Arguments: KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION})
	$(info Arguments: NAME=${NAME})
	docker run -d \
		-p 6443:6443 \
		-v ~/.kcp:/.kcp \
		--name ${NAME} ${IMG}:${KUBESTELLAR_VERSION}
	docker cp ${IMG}:/kcp-plugins ~/kcp
	docker cp ${IMG}:/kubestellar ~/kubestellar
	$(info export KUBECONFIG=${HOME}/.kcp/admin.kubeconfig)
	$(info export PATH=$$PATH:${HOME}/kcp:${HOME}/kubestellar/bin)


# Login into the running container
.PHONY: exec
exec:
	$(info Logging into a running container...)
	$(info Arguments: NAME=${NAME})
	docker exec -it ${NAME} /bin/bash


# Show container logs
.PHONY: logs
logs:
	@docker logs ${NAME}


# Kill the container, remove the container image and files
.PHONY: clean
clean:
	$(info Cleaning up...)
	$(info Arguments: IMG=${IMG})
	$(info Arguments: KUBESTELLAR_VERSION=${KUBESTELLAR_VERSION})
	$(info Arguments: NAME=${NAME})
	docker rm --force ${NAME}
	docker image rm ${IMG}:${KUBESTELLAR_VERSION}
	sudo rm -rf ~/kcp ~/kubestellar ~/.kcp


