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
    def __init__(self, kubeconfig, k8s_context):
        self.kubeconfig = kubeconfig
        self.context = k8s_context
        self.core_v1_api, self.app_v1_api, self.custom_objects_api = self.load_kubeconfig(kubeconfig, k8s_context)

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
        app_v1_api  = kubernetes.client.AppsV1Api(api_client)
        custom_objects_api = kubernetes.client.CustomObjectsApi(api_client)

        os.remove(tmp_config)
        return core_v1_api, app_v1_api, custom_objects_api

    def get_custom_objects_time(self, fname, grp, v, pl, name=None, ns=None, bindingPolicy=None, check=False, single=True):
        try:
            if ns != None and bindingPolicy != None:
               api_response = self.custom_objects_api.list_namespaced_custom_object(group=grp, version=v, plural=pl, namespace=ns, label_selector='transport.kubestellar.io/originOwnerReferenceBindingKey=' + bindingPolicy)
            elif name != None and ns != None and bindingPolicy == None:
               api_response = self.custom_objects_api.get_namespaced_custom_object(group=grp, version=v, namespace=ns, plural=pl, name=name)
            elif name != None and ns == None and bindingPolicy == None and check==False:
               api_response = self.custom_objects_api.get_cluster_custom_object(group=grp, version=v, plural=pl, name=name)
            else:
               api_response = self.custom_objects_api.list_cluster_custom_object(group=grp, version=v, plural=pl)

            if name == None or single == False:
               # Create the output file
               fout = open(fname + ".csv", 'a')

            if "items" in api_response:
               object_list = api_response["items"]
            else:
               object_list = [api_response]
            
            if check: # temporary fix - it must be improved later to optimize performance
                for i in range(0, len(object_list)):
                    temp_name = object_list[i]["metadata"]["name"]
                    if name in temp_name:
                        object_list = [object_list[i]]
                        break

            creation_timestamp = ""
            update_timestamp = ""
            status_condition = ""
            op_manager = ""
            manager_list = ["controller-manager", "ocm-status-addon", "kubelet", "work", "Go-http-client", "registration-operator", "transport-controller"]

            if object_list:
                for i in range(0, len(object_list)):
                    obj_name = object_list[i]["metadata"]["name"]
                    creation_timestamp = object_list[i]["metadata"]["creationTimestamp"]

                    if "status" in object_list[i]:
                        status = object_list[i]["status"]
                    else:
                        output=obj_name + "\t" + creation_timestamp + "\t" + update_timestamp + "\t" + status_condition + "\t" + op_manager
                        if name != None and single == True:
                           return output
                        else:
                           fout.write(output + "\n")
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
                            if manager in manager_list:
                                if pl == "manifestworks": 
                                   # Handle manifestwork objects differently: extract the timestamp when transport-controller updates a manifestwork object 
                                   if manager == "transport-controller":
                                      update_timestamp = managedFields[j]["time"]
                                      op_manager = manager
                                      break
                                   else:
                                     continue
                                elif op_type in managedFields[j]:
                                     update_timestamp = managedFields[j]["time"]
                                     op_manager = manager
                                     break
                        else:
                            continue

                    output=obj_name + "\t" + creation_timestamp + "\t" + update_timestamp + "\t" + status_condition + "\t" + op_manager
                    if name != None and single == True:
                       return output
                    else:
                       fout.write(output+ "\n")

            if name == None or single == False:
               fout.flush()
               fout.close()
            return None

        except ApiException as e:
            print("Exception when calling client->CustomObjectsApi(): %s\n" % e)


    def get_obj_time(self, fname, kind, ns, name=None):
        try:
            kind = kind.lower()
            if name != None:
                if kind == "pod" or kind == "pods":
                   api_response = self.core_v1_api.read_namespaced_pod(name, ns)
                elif kind == "deployment" or kind == "deployments":
                   api_response = self.app_v1_api.read_namespaced_deployment(name, ns)
                elif kind == "secret" or kind == "secrets":
                   api_response = self.core_v1_api.read_namespaced_secret(name, ns)
                elif kind == "configmap" or kind == "configmaps":
                   api_response = self.core_v1_api.read_namespaced_config_map(name, ns)
                elif kind == "service" or kind == "services":
                   api_response = self.core_v1_api.read_namespaced_service(name, ns)
                else:
                    print("Not supported object kind ...")
                    return
                object_list = [api_response]
            else:
                if kind == "pod" or kind == "pods": 
                   api_response = self.core_v1_api.list_namespaced_pod(ns)
                elif kind == "deployment" or kind == "deployments":
                   api_response = self.app_v1_api.list_namespaced_deployment(ns)
                elif kind == "secret" or kind == "secrets":
                   api_response = self.core_v1_api.list_namespaced_secret(ns)
                elif kind == "configmap" or kind == "configmaps":
                   api_response = self.core_v1_api.list_namespaced_config_map(ns)
                elif kind == "service" or kind == "services":
                   api_response = self.core_v1_api.list_namespaced_service(ns)
                else:
                    print("Not supported object kind ...")
                    return
                object_list = api_response.items
                
                # Create the output file
                fout = open(fname + ".csv", 'a')

            creation_timestamp = ""
            update_timestamp = ""
            status_condition = ""
            op_manager = ""
            manager_list = ["kube-controller-manager", "controller-manager", "ocm-status-addon", "kubelet", "k3s"]

            if object_list:
                for i in range(0, len(object_list)):
                    obj_name = object_list[i].metadata.name
                    creation_timestamp = object_list[i].metadata.creation_timestamp

                    if (kind != "pod" and kind != "pods") and (kind != "deployment" and kind != "deployments") and (kind != "service" and kind != "services"):
                       output=obj_name + "\t" + str(creation_timestamp)
                       if name != None:
                          return output
                       else:
                          fout.write(output + "\n")
                       continue

                    if object_list[i].status:
                       status = object_list[i].status
                    else:
                        output=obj_name + "\t" + str(creation_timestamp) + "\t" + str(update_timestamp) + "\t" + status_condition + "\t" + op_manager
                        if name != None:
                           return output
                        else:
                          fout.write(output+ "\n")
                        continue
                    
                    if kind == "pod" or kind == "pods":
                       if status.phase:
                          status_condition = status.phase
                    if kind == "deployment" or kind == "deployments":
                       replicas = status.replicas
                       readyReplicas = status.ready_replicas

                       if replicas == readyReplicas:
                          status_condition = "Available"
                       else:
                          status_condition = "Unavailable"                   
                       
                    # Get update timestamp for subresource of type "time"
                    managedFields = object_list[i].metadata.managed_fields
                    for j in range(0, len(managedFields)):
                        operation = managedFields[j].operation
                        manager = managedFields[j].manager 
                        op_type =  managedFields[j].subresource

                        if operation == "Update" and op_type == "status":
                            if manager in manager_list:
                               update_timestamp = managedFields[j].time
                               op_manager = manager
                               break
                        else:
                            continue 
                    
                    output = obj_name + "\t" + str(creation_timestamp) + "\t" + str(update_timestamp) + "\t" + status_condition + "\t" + op_manager
                    if name != None:
                       return output
                    else: 
                       fout.write(output + "\n")

            if name == None:
               fout.flush()
               fout.close()
            return None
        except ApiException as e:
            print("Exception when calling CoreV1Api/AppV1Api: %s\n" % e)