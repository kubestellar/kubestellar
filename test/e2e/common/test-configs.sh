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
set_scenario_config() {
  case "$1" in
    default)
      config_name="default"
      config_description="Three new kind clusters"
      config_wds_type="new"
      config_its_type="new"
      config_cluster_source="kind"
      config_combined_control_plane="false"
      ;;
    wds-on-host)
      config_name="wds-on-host"
      config_description="WDS on hosting cluster, new ITS"
      config_wds_type="hosting"
      config_its_type="new"
      config_cluster_source="kind"
      config_combined_control_plane="false"
      ;;
    its-on-host)
      config_name="its-on-host"
      config_description="ITS on hosting cluster, new WDS"
      config_wds_type="new"
      config_its_type="hosting"
      config_cluster_source="kind"
      config_combined_control_plane="false"
      ;;
    combined)
      config_name="combined"
      config_description="Single control plane for both WDS and ITS"
      config_wds_type="new"
      config_its_type="new"
      config_cluster_source="kind"
      config_combined_control_plane="true"
      ;;
    *)
      echo "Unknown scenario: $1" >&2
      return 1
      ;;
  esac
}

list_scenarios() {
  echo "Available test scenarios:"
  echo "  default      - Three new kind clusters"
  echo "  wds-on-host  - WDS on hosting cluster, new ITS"
  echo "  its-on-host  - ITS on hosting cluster, new WDS"
  echo "  combined     - Single control plane for both WDS and ITS"
} 