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

# Usage: $0 $commit

# This will `git checkout $commit`,
# make copy in subdirectory `$PWD/forcount`,
# remove subdirectories that should not be counted,
# then apply `count-tree.sh $PWD/counts $commit $timestamp`.
# The timestamp is found by `git show --no-patch --no-notes --format=%ct`.

if [ $# -ne 1 ]; then
    echo "Usage: $0 \$commit" >&2
    exit 1
fi

commit="$1"
bindir=$(cd $(dirname $0); pwd)

rm -rf forcount
git checkout "$commit"
timestamp=$(git show --no-patch --no-notes --format=%ct)
cp -R . forcount
cd forcount
rm -rf counts .git .vscode docs/venv docs/__pycache__
${bindir}/count-tree.sh ../counts "$commit" "$timestamp"
