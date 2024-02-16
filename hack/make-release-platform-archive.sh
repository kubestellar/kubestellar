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

# Usage: $0 KCPE_VERSION scope target_os target_arch

if [ $# != 4 ]; then
    echo "$0 usage: KCPE_VERSION scope target_os target_arch" >&2
    exit 1
fi

kcpe_version="$1"
shift
scope="$1"
shift
target_os="$1"
shift
target_arch="$1"
shift

case "$scope" in
    (full) scope_part="";;
    (user) scope_part=user;;
    (*) echo "$0: scope must be 'full' or 'user', not '$scope'" >& 2;
	exit 1;;
esac

archname="kubestellar${scope_part}_${kcpe_version}_${target_os}_${target_arch}.tar.gz"

if shasum -a 256 "$0" &> /dev/null
then sumcmd="shasum -a 256"
else sumcmd=sha256sum
fi

set -e

srcdir=$(dirname "$0")
cd "$srcdir/.."

rm -rf bin/*
make ${scope_part}build OS="$target_os" ARCH="$target_arch"

echo $'#!/usr/bin/env bash\necho' ${kcpe_version} > bin/kubestellar-release
chmod a+x bin/kubestellar-release
mkdir -p build/release
tar czf "build/release/$archname" --exclude bin/.gitignore bin config core-helm-chart README.md LICENSE
cd build/release
touch checksums256.txt
grep -vw "$archname" checksums256.txt > /tmp/$$.txt || true
$sumcmd "$archname" >> /tmp/$$.txt
cp /tmp/$$.txt checksums256.txt
rm /tmp/$$.txt
