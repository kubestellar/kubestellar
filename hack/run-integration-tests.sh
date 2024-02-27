#!/usr/bin/env bash

# Copyright 2024 The KubeStellar Authors.
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

# Usage: cd $kubestellar_repo_dir; $0

# This will download a copy of etcd and add it to $PATH,
# then run the integration tests.

# Run this ONLY on Linux on amd64 (x64_64).

set -x
set -e

etcd_archive_url="https://github.com/etcd-io/etcd/releases/download/v3.5.9/etcd-v3.5.9-linux-amd64.tar.gz"
wget --no-verbose $etcd_archive_url -O etcd.tar.gz
expected_sha=d59017044eb776597eca480432081c5bb26f318ad292967029af1f62b588b042
if which sha256sum
then got_sha=$(sha256sum     etcd.tar.gz | awk '{ print $1 }')
else got_sha=$(shasum -a 256 etcd.tar.gz | awk '{ print $1 }')
fi
if [ $expected_sha != "$got_sha" ]; then
    echo "Got SHA256 $got_sha instead of expected $expected_sha" >& 2
    exit 1
fi

tar xzf etcd.tar.gz
export PATH=${PATH}:${PWD}/etcd-v3.5.9-linux-amd64
rm etcd.tar.gz

CONTROLLER_TEST_NUM_OBJECTS=12 go test -v ./test/integration/controller-manager -args -v=5
