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

# Usage: $0 $directory

word=kcp

# This script will output one CSV line with the following columns.
# - Directory
# - Count of "$word"
# - Count of "$word-dev"
# - Count of "$word" in *.go files
# - Count of "$word" in go.* files
# - Count of "$word" in *.sh files
# - Count of "$word" in *.md files
# - Count of "$word" in *.yaml files

place="$1"
grepflags="--binary-files=without-match -r"
total=$(grep $grepflags -i ${word} "$place"  | echo $(wc -l))
dev=$(  grep $grepflags ${word}-dev "$place"  | echo $(wc -l))
dotgo=$(grep $grepflags --include \*.go -i ${word} "$place" | echo $(wc -l))
godot=$(grep $grepflags --include 'go.*[dkm]' ${word} "$place" | echo $(wc -l))
dotsh=$(grep $grepflags --include \*.sh -i ${word} "$place" | echo $(wc -l))
dotmd=$(grep $grepflags --include \*.md -i ${word} "$place" | echo $(wc -l))
dotyl=$(grep $grepflags --include \*.yaml -i ${word} "$place" | echo $(wc -l))
echo "$place,$total,$dev,$dotgo,$godot,$dotsh,$dotmd,$dotyl"
