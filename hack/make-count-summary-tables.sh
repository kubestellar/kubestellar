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

# Usage: $0

# This creates summary tables in counts based on the per-commit data there.

result=counts/sum-over-directories--table.csv

echo timestamp,commit,descr,total,org,dotgo,gometa,dotsh,dotmd,dotyaml > $result
find counts -name sum-over-directories.csv -exec cat \{\} \; | sort >> $result

find counts -name matrix.csv -exec grep '^\./\(space-framework/\)\?[^/]*,' \{\} \; | cut -f1 -d, | sort | uniq | while read top; do
    topn=${top:2}
    topn=$(sed 's|/|-|' <<<"$topn")
    result=counts/sum-over-directories-${topn}--table.csv
    echo timestamp,commit,descr,total,org,dotgo,gometa,dotsh,dotmd,dotyaml > $result
    find counts -name sum-over-directories-${topn}.csv -exec cat \{\} \; | sort >> $result
done
