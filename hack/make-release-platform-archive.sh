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

# Usage: $0 KCPE_VERSION target_os target_arch

if [ $# != 3 ]; then
    echo "$0 usage: KCPE_VERSION target_os target_arch" >&2
    exit 1
fi

kcpe_version="$1"
target_os="$2"
target_arch="$3"
archname="kubestellar_${kcpe_version}_${target_os}_${target_arch}.tar.gz"

if shasum -a 256 "$0" &> /dev/null
then sumcmd="shasum -a 256"
else sumcmd=sha256sum
fi

set -e

srcdir=$(dirname "$0")
cd "$srcdir/.."

rm -rf bin/*
make build OS="$target_os" ARCH="$target_arch" WHAT="./cmd/kubectl-kubestellar-syncer_gen ./cmd/kubestellar-version ./cmd/kubestellar-where-resolver ./cmd/mailbox-controller ./cmd/placement-translator"
echo $'#!/usr/bin/env bash\necho' ${kcpe_version@Q} > bin/kubestellar-release
chmod a+x bin/kubestellar-release
mkdir -p build/release
tar czf "build/release/$archname" --exclude bin/.gitignore bin config user/helm-chart README.md LICENSE
cd build/release
touch checksums256.txt
grep -vw "$archname" checksums256.txt > /tmp/$$.txt || true
$sumcmd "$archname" >> /tmp/$$.txt
cp /tmp/$$.txt checksums256.txt
rm /tmp/$$.txt
