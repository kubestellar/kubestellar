#!/usr/bin/env python3

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

import os
import yaml
import tempfile
import kubernetes
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from pprint import pprint
import urllib3
import re
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


class Collector():
    def __init__(self, kubeconfig, k8s_context, fout):
        self.fout = fout
        self.kubeconfig = kubeconfig
        self.context = k8s_context
        self.core_v1_api, self.custom_objects_api = self.load_kubeconfig(kubeconfig, k8s_context)

    def load_kubeconfig(self, kubeconfig, context):
        kube_config_orig = kubeconfig
        tmp_config = tempfile.NamedTemporaryFile().name

        with open(kube_config_orig, "r") as fd:
            kubeconfig = yaml.load(fd, Loader=yaml.FullLoader)
        for cluster in kubeconfig["clusters"]:
            cluster["cluster"]["insecure-skip-tls-verify"] = True
        with open(tmp_config, "w") as fd:
            yaml.dump(kubeconfig, fd, default_flow_style=False)

        if context == "":
            config.load_kube_config(config_file=tmp_config)
        else:
            config.load_kube_config(config_file=tmp_config, context=context)
        
        api_client = kubernetes.client.ApiClient()
        core_v1_api = kubernetes.client.CoreV1Api(api_client)
        custom_objects_api = kubernetes.client.CustomObjectsApi(api_client)

        os.remove(tmp_config)
        return core_v1_api, custom_objects_api

    def get_crd_time(self, grp, v, pl, name=None, check=False):
        try:
            if name != None and check == False:
                api_response = self.custom_objects_api.list_cluster_custom_object(group=grp, version=v, plural=pl, field_selector='metadata.name=' + name)
            else:
                api_response = self.custom_objects_api.list_cluster_custom_object(group=grp, version=v, plural=pl)

            object_list = api_response["items"]
            if check: # temporary fix to collect Workstatus objects - it must be improved later to optimize performance
                for i in range(0, len(object_list)):
                    temp_name = object_list[i]["metadata"]["name"]
                    if name in temp_name:
                        object_list = [object_list[i]]
                        break

            creation_time = ""
            update_time = ""
            status_condition = ""
            op_manager = ""
            
            dict_manager_list = ["controller-manager", "ocm-status-addon", "kubelet", "work"]
            if os.path.exists(self.fout):
                mode = 'a' # append if already exists
            else:
                mode = 'w' # make a new file if not

            fout = open(self.fout, mode)

            if object_list:
                for i in range(0, len(object_list)):
                    name = object_list[i]["metadata"]["name"]
                    creation_time = object_list[i]["metadata"]["creationTimestamp"]

                    if "status" in object_list[i]:
                        status = object_list[i]["status"]
                    else:
                        fout.write(name + "\t" + creation_time + "\t" + update_time + "\t" + status_condition + "\t" + op_manager + "\n")
                        continue

                    if "phase" in status:
                        status_condition = status["phase"]
                   
                    # Get update timestamp for subresource of type "time"
                    managedFields = object_list[i]["metadata"]["managedFields"]
                    for j in range(0, len(managedFields)):
                        operation = managedFields[j]["operation"]
                        manager = managedFields[j]["manager"] 
                        op_type = "subresource"

                        if operation == "Update":
                            if manager in dict_manager_list:
                                if op_type in managedFields[j]:
                                    update_time = managedFields[j]["time"]
                                    op_manager = manager
                                    break
                        else:
                            continue 
                    fout.write(name + "\t" + creation_time + "\t" + update_time + "\t" + status_condition + "\t" + op_manager + "\n")
            fout.close()
        except ApiException as e:
            print("Exception when calling client->CustomObjectsApi(): %s\n" % e)


    def get_pod_time(self, ns, name=None):
        try:
            if name != None:
                api_response = self.core_v1_api.list_namespaced_pod(ns, field_selector='metadata.name=' + name)
            else: 
                api_response = self.core_v1_api.list_namespaced_pod(ns)

            object_list = api_response.items

            creation_time = ""
            update_time = ""
            status_condition = ""
            op_manager = ""

            dict_manager_list = ["controller-manager", "ocm-status-addon", "kubelet", "k3s"]
            if os.path.exists(self.fout):
                mode = 'a' # append if already exists
            else:
                mode = 'w' # make a new file if not

            fout = open(self.fout, mode)

            if object_list:
                for i in range(0, len(object_list)):
                    name = object_list[i].metadata.name
                    creation_time = object_list[i].metadata.creation_timestamp

                    if object_list[i].status:
                        status = object_list[i].status
                    else:
                        fout.write(name + "\t" + creation_time + "\t" + update_time + "\t" + status_condition + "\t" + op_manager + "\n")
                        continue

                    if status.phase:
                        status_state = status.phase
                        status_condition = status_state
                    
                    # Get update timestamp for subresource of type "time"
                    managedFields = object_list[i].metadata.managed_fields
                    for j in range(0, len(managedFields)):

                        operation = managedFields[j].operation
                        manager = managedFields[j].manager 
                        op_type =  managedFields[j].subresource

                        if operation == "Update" and op_type == "status":
                            if manager in dict_manager_list:
                                update_time = managedFields[j].time
                                op_manager = manager
                                break
                        else:
                            continue 
                    fout.write(name + "\t" + str(creation_time) + "\t" + str(update_time) + "\t" + status_condition + "\t" + op_manager + "\n")
            fout.close()
        except ApiException as e:
            print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)
