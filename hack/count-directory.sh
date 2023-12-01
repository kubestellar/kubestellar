#!/usr/bin/env bash

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
dotgo=$(grep $grepflags ${word} "$place" --include \*.go | echo $(wc -l))
godot=$(grep $grepflags ${word} "$place" --include 'go.*[dkm]' | echo $(wc -l))
dotsh=$(grep $grepflags ${word} "$place" --include \*.sh | echo $(wc -l))
dotmd=$(grep $grepflags ${word} "$place" --include \*.md | echo $(wc -l))
dotyl=$(grep $grepflags ${word} "$place" --include \*.yaml | echo $(wc -l))
echo "$place,$total,$dev,$dotgo,$godot,$dotsh,$dotmd,$dotyl"
