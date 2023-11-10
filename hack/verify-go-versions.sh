#!/usr/bin/env bash

# Copyright 2022 The KubeStellar Authors.
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

set -e
set -o pipefail

MINIMAL_VERSION=$(grep "go 1." go.mod | sed 's/go //')

# grep "FROM golang:" Dockerfile | { ! grep -v "${MINIMAL_VERSION}"; } || { echo "Wrong go version in Dockerfile, expected ${MINIMAL_VERSION}"; exit 1; }
# grep -w "go-version:" .github/workflows/*.yaml | { ! grep -v "go-version: v${MINIMAL_VERSION}"; } || { echo "Wrong go version in .github/workflows/*.yaml, expected ${MINIMAL_VERSION}"; exit 1; }
# Note CONTRIBUTING.md isn't copied in the Dockerfile
# if [ -e CONTRIBUTING.md ]; then
#   grep "golang.org/doc/install" CONTRIBUTING.md | { ! grep -v "${MINIMAL_VERSION}"; } || { echo "Wrong go version in CONTRIBUTING.md expected ${MINIMAL_VERSION}"; exit 1; }
# fi

if ! [ -x "$(command -v go)" ]; then # validate go is installed
    echo "go is not installed on your machine, exiting."
    exit 1
fi

if [ -z "${IGNORE_GO_VERSION}" ]; then # validate go version is sufficient
  ENV_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
  if [[ "$ENV_VERSION" < "$MINIMAL_VERSION" ]]; then
    echo "Unexpected go version installed. expected minimal go version ${MINIMAL_VERSION} while your environment has version ${ENV_VERSION}. Use IGNORE_GO_VERSION=1 to skip this check."
    exit 1
  fi
fi
