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

# Usage: $0 $countdir $id $ts

# This script will write the file ${countdir}/${id}/matrix.csv.
# The matrix has the columns documented in count-directory.sh.

# This script also writes ${countdir}/${id}/sum-over-directories.csv.
# This has one line, with the same columns as the matrix except that
# the directory is replaced by two columns: one holding ${ts} and
# and one holding ${id}.
# Thus, the concatenation of all those sum files makes one CSV table.

if [ $# -ne 3 ]; then
    echo "Usage: $0 \$countdir \$id \$timestamp" >&2
    exit 1
fi

countdir="$1"
id="$2"
ts="$3"
bindir=$(dirname $0)
mkdir -p "${countdir}/${id}"

find . -type d -exec ${bindir}/count-directory.sh \{\} \; > "${countdir}/${id}/matrix.csv"

grep '^\.,' "${countdir}/${id}/matrix.csv" | sed "s/.,/${ts},${id},/" > "${countdir}/${id}/sum-over-directories.csv"
