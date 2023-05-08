#!/usr/bin/env bash
# Copyright 2023 The KCP Authors.
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

# Usage: $0 (--clusters <list of kind cluster names> e.g, "florin guilder")

# Purpose: clean up environment by removing kcp & kcp-edge and
# deleting all the generated files (e.g., edge syncer manifests and logs files)

# Assumption: kcp and kcp-edge are running 

# This script requires the kcp folders, logs and syncer manifest files
# are created at your current directory - where this script is executed

clusters="florin guilder"

while (( $# > 0 )); do
    if [ "$1" == "--clusters" ]; then
        clusters=$2
    fi 
    shift
done

# Deleting kind clusters
for c in ${clusters[@]}
do 
  if [ $(kind get clusters | grep $c) > /dev/null 2>&1 ]; then
     echo "Deleting kind cluster $c ...."
     kind delete cluster --name $c > /dev/null 2>&1
  fi
done

pkill -f kubectl-kcp-playground
pkill -f kcp
pkill -f mailbox-controller
pkill -f placement-translator
pkill -f "scheduler -v 2" # edge-scheduler
rm -rf $(pwd)/kcp
rm -rf $(pwd)/.kcp


if [ !$(ls | grep syncer.sh | wc -l) == 0 ]; then
      rm $(pwd)/*syncer.yaml
      echo "Deleted syncer manifest"
fi 

if [ !$(ls | grep log.txt | wc -l) == 0 ]; then
      rm $(pwd)/*log.txt
      echo "Deleted log files"
fi 

echo "Finished deletion ...."