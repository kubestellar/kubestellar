#!/usr/bin/env bash

# Copyright 2023 The KCP Authors.
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

# Usage: $0

# This script will ensure that https://github.com/yana1205/kcp/tree/emc
# has been `git clone`d to the subdirectory build/syncer-kcp
# and built there.

if [ "$#" != 0 ]; then
    echo "$0 takes no arguments" >&2
    exit 1
fi

if ! [ -d build ]; then
    mkdir build
fi

if ! [ -d build/syncer-kcp ]; then
    git clone https://github.com/yana1205/kcp build/syncer-kcp
fi

set -e

cd build/syncer-kcp
git checkout emc
[ -x bin/kubectl-kcp ] || make build WHAT=./cmd/kubectl-kcp
