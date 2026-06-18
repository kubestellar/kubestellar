#!/usr/bin/env bash

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

# Usage: $0 KCPE_VERSION

if [ $# != 1 ]; then
    echo "$0 usage: KCPE_VERSION" >&2
    exit 1
fi

kcpe_version="$1"

set -e

srcdir=$(dirname "$0")
$srcdir/make-release-container.sh $kcpe_version

os_names=( darwin darwin linux linux linux linux )
arch_names=( arm64 amd64 arm64 amd64 ppc64le s390x)
length=${#os_names[@]}
for (( i=0; i<length; i++ ));
do
	echo "${os_names[$i]} ${arch_names[$i]}"
    $srcdir/make-release-platform-archive.sh $kcpe_version user ${os_names[$i]} ${arch_names[$i]}
    $srcdir/make-release-platform-archive.sh $kcpe_version full ${os_names[$i]} ${arch_names[$i]}
done

