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

. demo-magic.sh
clear
TYPE_SPEED=10
DEMO_PROMPT="${GREEN}(cluster aware client)➜ ${CYAN}\W ${COLOR_RESET}"
clear

pei  "kind create cluster --name ks-client1"
echo " "
pe "kubectl --context kind-ks-lc4 create configmap lc1-cm"
echo " "
