#!/usr/bin/env bash

# Copyright 2021 The KubeStellar Authors.
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

# This script ensures that the generated client code checked into git is up-to-date
# with the generator. If it is not, re-generate the configuration to update it.

set -o errexit
set -o nounset
set -o pipefail

if [[ "$(git status --porcelain=1 -- api config/crd/bases pkg/crd/files pkg/generated)" != "" ]]; then
    echo "Sorry, some relevant files are not in the current git commit."
    echo "Correct checking can not be done."
    git status -- api config/crd/bases pkg/crd/files pkg/generated
    exit 1
fi

pwd
make all-generated
if [[ "$(git status --porcelain=1 -- api config/crd/bases pkg/crd/files pkg/generated)" != "" ]]; then
	cat << EOF
ERROR: This check enforces that all the derived files have been derived correctly.
ERROR: At least one is not. Run the following command to re-
ERROR: generate the derived files:
ERROR: $ make all-generated
ERROR: The following differences were found:
EOF
	git status -- api config/crd/bases pkg/crd/files pkg/generated
	git diff
	exit 1
fi
