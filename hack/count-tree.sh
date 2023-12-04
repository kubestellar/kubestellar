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

# Usage: $0 $countdir $timestamp $id $description

# This script will write the file ${countdir}/${id}/matrix.csv.
# The matrix has the columns documented in count-directory.sh.

# This script also writes ${countdir}/${id}/sum-over-directories.csv.
# This has one line, with the same columns as the matrix except that
# the directory is replaced by three columns: one holding ${ts},
# one holding ${id}, and one holding "${descr}".
# Thus, the concatenation of all those sum files makes one CSV table.

# This script write analogous single-line CVS files for high-level directories.

# This script also writes a tops.csv file that has rows from matrix.csv
# for high-level directories.

if [ $# -ne 4 ]; then
    echo "Usage: $0 \$countdir \$timestamp \$id \$description" >&2
    exit 1
fi

countdir="$1"
ts="$2"
id="$3"
descr="$4"
bindir=$(dirname $0)
mkdir -p "${countdir}/${id}"
echo "${ts},${id},\"${descr}\"" > "${countdir}/${id}/ids.csv"

find . -type d -exec ${bindir}/count-directory.sh \{\} \; | sort > "${countdir}/${id}/matrix.csv"

grep '^\.,' "${countdir}/${id}/matrix.csv" | sed "s/.,/${ts},${id},\"${descr}\",/" > "${countdir}/${id}/sum-over-directories.csv"

grep '^\.\(/\(space-framework/\)\?[^/]*\)\?,' "${countdir}/${id}/matrix.csv" > "${countdir}/${id}/tops.csv"

grep '^\./\(space-framework/\)\?[^/]*,' "${countdir}/${id}/matrix.csv" | cut -f1 -d, | while read top; do
    topn=${top:2}
    topn=$(sed 's|/|-|' <<<"$topn")
    grep "^$top," "${countdir}/${id}/matrix.csv" | sed "s|^${top},|${ts},${id},\"${descr}\",|" > "${countdir}/${id}/sum-over-directories-${topn}.csv"
done
