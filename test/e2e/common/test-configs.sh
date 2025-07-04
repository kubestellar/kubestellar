# shellcheck shell=bash
# test/e2e/common/test-configs.sh
# Copyright 2024 The KubeStellar Authors.
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


get_scenario_config() {
  case "$1" in
    default)
      # name description wds_type its_type cluster_source combined_control_plane
      echo "default 'Three new kind clusters' new new kind false"
      ;;
    wds-on-host)
      echo "wds-on-host 'WDS on hosting cluster, new ITS' hosting new kind false"
      ;;
    its-on-host)
      echo "its-on-host 'ITS on hosting cluster, new WDS' new hosting kind false"
      ;;
    combined)
      echo "combined 'Single control plane for both WDS and ITS' new new kind true"
      ;;
    *)
      echo "Unknown scenario: $1" >&2
      return 1
      ;;
  esac
} 