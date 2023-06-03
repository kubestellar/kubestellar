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
cd "$REPO_ROOT/docs"

FILE_LIST="$1"

# regular expression pattern to match backtick fenced shell code blocks
start_pattern="^\`\`\`shell"
stop_pattern="^\`\`\`"

# array to store the shell code blocks
code_blocks=()
code_blocks+=("cd $REPO_ROOT/docs/scripts/")

if [ -f "/etc/os-release" ]; then
  if [[ $(grep -i "ubuntu" /etc/os-release) ]]; then
    echo "The operating system is Ubuntu."
    # code_blocks+=('sudo apt-get install -y curl wget gnupg')
    # code_blocks+=('source /etc/os-release ; sudo sh -c "echo 'deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_${VERSION_ID}/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list"2')
    # code_blocks+=('source /etc/os-release ; wget -nv https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable/xUbuntu_${VERSION_ID}/Release.key -O- | sudo apt-key add -')
    # code_blocks+=('sudo apt update && sudo apt install -y podman && podman machine init && podman machine start')
    # code_blocks+=('sudo snap install go --classic')
    # code_blocks+=('go install sigs.k8s.io/kind@v0.17.0')
    # code_blocks+=('curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/')
    # code_blocks+=('sudo apt update && sudo apt install -y jq')
    # code_blocks+=('sudo apt update && sudo apt install -y kubectl')
    # code_blocks+=('alias docker=podman')
  fi
else
  echo "The operating system is not Ubuntu."
  code_blocks+=("brew install podman && podman machine init && podman machine start")
  code_blocks+=('alias docker=podman')
  code_blocks+=('KIND_EXPERIMENTAL_PROVIDER=podman')

fi

inside_block=0
repo_raw_url='https://raw.githubusercontent.com/kcp-dev/edge-mc'
ks_branch='main'
ks_tag='latest'

# read the readme file line by line
while IFS= read -r line; do
  # check if the line matches the pattern
  if [[ $line =~ $stop_pattern ]]; then
    inside_block=0
  fi
  if [[ $line =~ $start_pattern ]]; then
    inside_block=1
  fi
  
  if [[ $inside_block == 1 ]]; then
    # remove the backticks from the code block
    code_block="${line//\`\`\`shell/}"
    code_block="${code_block/\{\{ config.repo_raw_url \}\}/$repo_raw_url}"
    code_block="${code_block/\{\{ config.ks_branch \}\}/$ks_branch}"
    code_block="${code_block/\{\{ config.ks_tag \}\}/$ks_tag}"

    # add the code block to the array
    code_blocks+=("$code_block")
  fi
done < "$FILE_LIST"

generated_script_file="$REPO_ROOT/docs/scripts/generated_script.sh"
echo "" > "$generated_script_file"

# echo the code blocks into a script file
for code_block in "${code_blocks[@]}"; do
  echo "$code_block"  >> "$generated_script_file"
done

# make the generated script executable
chmod +x "$generated_script_file"

# run the generated script
"$generated_script_file"

rm $generated_script_file