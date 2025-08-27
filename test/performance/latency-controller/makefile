# Copyright 2023 The KubeStellar Authors.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ifeq ($(filter docker-build docker-push,$(MAKECMDGOALS)),docker-build docker-push)
ifndef IMAGE
$(error IMAGE is not set. Usage: make <target> IMAGE=<your-image:tag>)
endif
endif

# Build the Go binary (linux/amd64)
build:
	GOOS=linux GOARCH=amd64 go build -o manager main.go

# Build the Docker image (multi-stage). Requires IMAGE arg
docker-build: build
	docker build -t $(IMAGE) -f ./Dockerfile ../../..

# Push the image to the registry (expects credentials already configured)
docker-push:
	docker push $(IMAGE)

# Clean artifacts
clean:
	rm -f manager
	docker rmi $(IMAGE) || true
