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

set -o errexit
set -o nounset
set -o pipefail
# set -o xtrace

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$REPO_ROOT"

FILE_LIST=()
SAVEIFS=$IFS

IFS=',' read -r -a FILE_LIST <<< "${MANIFEST}"
echo ${FILE_LIST[0]}

# regular expression pattern to match backtick fenced shell code blocks
start_pattern="^\`\`\`shell"
start_hidden_pattern="^\`\`\` \{\.bash \.hide-me\}"
stop_pattern="^\`\`\`"
include_pattern="\s*include-markdown\s*"

# array to store the shell code blocks
code_blocks=()
# code_blocks+=("cd $REPO_ROOT/docs/scripts/")
code_blocks+=("cd $REPO_ROOT")

if [ -f "/etc/os-release" ]; then
  if [[ $(grep -i "ubuntu" /etc/os-release) ]]; then
    echo "The operating system is Ubuntu."
  fi
else
  echo "The operating system is not Ubuntu."
fi

repo_url=$(yq -r ".repo_url" $REPO_ROOT/docs/mkdocs.yml)
repo_raw_url=$(yq -r ".repo_raw_url" $REPO_ROOT/docs/mkdocs.yml)
ks_branch=$(yq -r ".ks_branch" $REPO_ROOT/docs/mkdocs.yml)
ks_tag=$(yq -r ".ks_tag" $REPO_ROOT/docs/mkdocs.yml)
ks_kind_port_num=$(yq -r ".ks_kind_port_num" $REPO_ROOT/docs/mkdocs.yml)
ks_current_tag=$(yq -r ".ks_current_tag" $REPO_ROOT/docs/mkdocs.yml)

code_blocks+=('set -o errexit')
code_blocks+=('set -o nounset')
code_blocks+=('set -o pipefail')
code_blocks+=('set -o xtrace')

function parse_file() 
{
  local inside_block=0
  local file_name=$(echo $1 | sed 's/"//g')
  local path_name=$(echo "${file_name%/*}/")

  # echo $path_name
  echo \#\#\# parsing $file_name  
  # read the readme file line by line
  while IFS= read -r line; do
    # check if the line matches the pattern
    if [[ $line =~ $stop_pattern ]]; then
      inside_block=0  # not inside a shell codeblock
    fi
    if [[ $line =~ $start_pattern ]]; then
      inside_block=1  # we are inside a shell codeblock
    fi
    if [[ $line =~ $start_hidden_pattern ]]; then
      inside_block=1  # we are inside a shell codeblock
    fi

    if [[ $line =~ $include_pattern ]]; then
      included_file_name=$(echo $line | sed 's/include-markdown \"//')
      full_included_file_name=$(echo \"$REPO_ROOT/docs/$path_name$included_file_name)
      parse_file "\"$path_name$included_file_name\""
    fi
    
    if [[ $inside_block == 1 ]]; then
      # remove the backticks from the code block
      
      if [[ $line =~ $start_pattern ]]; then
        echo ignore this line: $line
      else
        if [[ $line =~ $start_hidden_pattern ]]; then
          echo ignote this line: $line
        else
          # code_block="${line//\`\`\`shell/}"
          code_block=$line
          code_block="${code_block/\{\{ config.repo_url \}\}/$repo_url}"
          code_block="${code_block/\{\{ config.repo_raw_url \}\}/$repo_raw_url}"
          code_block="${code_block/\{\{ config.ks_branch \}\}/$ks_branch}"
          code_block="${code_block/\{\{ config.ks_current_tag \}\}/$ks_current_tag}"
          code_block="${code_block/\{\{ config.ks_tag \}\}/$ks_tag}"
          code_block="${code_block/\{\{ config.ks_kind_port_num \}\}/$ks_kind_port_num}"
          echo $code_block
          # add the code block to the array
          code_blocks+=("$code_block")
        fi
      fi
    fi
  done < "$REPO_ROOT/$file_name"
}

for FILE_NAME in "${FILE_LIST[@]}"
do
  parse_file "\"$FILE_NAME\""
done

IFS=$SAVEIFS

generated_script_file="$REPO_ROOT/docs/scripts/generated_script.sh"
echo "" > "$generated_script_file"

# echo the code blocks into a script file
for code_block in "${code_blocks[@]}"; do
  echo "$code_block"  >> "$generated_script_file"
done

echo
echo "Generated script file follows"
cat "$generated_script_file"

echo
echo "Execution follows"
set -o xtrace

# make the generated script executable
chmod +x "$generated_script_file"

# run the generated script
"$generated_script_file"

rm $generated_script_file
