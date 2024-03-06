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
    etcd_version=3.5.12
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
	(darwin-amd64)  expected_sha=96e2e5f3c68744fe0981a3d52b8815aa5d3a3bb3b4e48cee54dd29bd33dfe355;;
	(darwin-arm64)  expected_sha=d2a2d003a237ee3aaed2859492aa63b411d2c5de016a9b44a3be72865cc33933;;
	(linux-amd64)   expected_sha=f2ff0cb43ce119f55a85012255609b61c64263baea83aa7c8e6846c0938adca5;;
	(linux-arm64)   expected_sha=31f30c01918771ece28d6e553e0f33be9483ced989896ecf6bbe1edb07786141;;
	(windows-amd64) expected_sha=ab63500a3eb1cbfd9d863759fc5a039ad724c0b493b26c079bb53b19e7c21330;;
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
