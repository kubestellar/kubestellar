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

if ! which etcd; then
    etcd_version=3.5.16
    platform="$(go env GOOS)-$(go env GOARCH)"

    case "$(go env GOOS)" in
	(darwin|windows) fmt="zip";    extract="unzip"  ;;
	(linux)          fmt="tar.gz"; extract="tar xzf";;
	(*) echo "Unsupported OS $(go env GOOS)" >& 2
	    exit 1;;
    esac

    etcd_archive_url="https://github.com/etcd-io/etcd/releases/download/v${etcd_version}/etcd-v${etcd_version}-${platform}.${fmt}"
    wget --no-verbose $etcd_archive_url -O etcd.${fmt}
    case "$platform" in
	(darwin-amd64)  expected_sha=f9d0c97374655bab27934f7dd19193bc540d692c0a582b3bc686875a0a72754d ;;
	(darwin-arm64)  expected_sha=912c90ce9f79e822a659462a19049bba9e7f0cb42390ed3a7858781d4a7c2eb9;;
	(linux-amd64)   expected_sha=b414b27a5ad05f7cb01395c447c85d3227e3fb1c176e51757a283b817f645ccc;;
	(linux-arm64)   expected_sha=8e68c55e6d72b791a9e98591c755af36f6f55aa9eca63767822cd8a3817fdb23;;
	(windows-amd64) expected_sha=aed7f7c53577e14d12adbbddb1a81faed2f1ff2eacc0ade52c093b35eda7ef38;;
	(*) echo "Unsupported platform $platform" >&2
	    exit 1 ;;
    esac

    if which sha256sum
    then got_sha=$(sha256sum     etcd.${fmt} | awk '{ print $1 }')
    else got_sha=$(shasum -a 256 etcd.${fmt} | awk '{ print $1 }')
    fi
    if [ $expected_sha != "$got_sha" ]; then
	echo "Got SHA256 $got_sha instead of expected $expected_sha for platform $platform" >& 2
	exit 1
    fi

    $extract etcd.${fmt}
    export PATH=${PATH}:${PWD}/etcd-v${etcd_version}-${platform}
    rm etcd.${fmt}
fi

CONTROLLER_TEST_NUM_OBJECTS=360 go test -v ./test/integration/controller-manager -args -v=5
