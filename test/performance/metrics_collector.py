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
import sys
import yaml
import tempfile
import kubernetes
from multiprocessing import Process
from kubernetes import client, config, watch
from kubernetes.client.rest import ApiException
from pprint import pprint
import urllib3
from kshelper import *
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


class DataParser():
    def __init__(self, kubeconfig, ctx_1, ctx_2, ctx_3):
        self.kubeconfig = kubeconfig
        self.current_ctx = ctx_1
        self.c1 = Collector(kubeconfig, ctx_1)
        self.c2 = Collector(kubeconfig, ctx_2)
        self.c3 = Collector(kubeconfig, ctx_3)

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
        os.remove(tmp_config)

    def collect_obj_time(self, c, name, kind, fname=None, obj_ns=None, obj_grp=None, obj_v=None, obj_pl=None, bindingPolicy=None, check=False, single=True):
        output=None
        if kind == "cr":       
           if name != None and obj_ns != None and bindingPolicy != None:
              output= c.get_custom_objects_time(None, obj_grp, obj_v, obj_pl, name, obj_ns, bindingPolicy, check)
           elif name != None and obj_ns != None and bindingPolicy == None:
              output= c.get_custom_objects_time(None, obj_grp, obj_v, obj_pl, name, obj_ns, None, check)
           elif name == None and obj_ns != None and bindingPolicy != None:
              output= c.get_custom_objects_time(fname, obj_grp, obj_v, obj_pl, None, obj_ns, bindingPolicy, check)
           elif name != None and obj_ns == None and bindingPolicy != None:
              output= c.get_custom_objects_time(fname, obj_grp, obj_v, obj_pl, name, None, bindingPolicy, check, False)
           else:
              output= c.get_custom_objects_time(fname, obj_grp, obj_v, obj_pl, name, None, None, check, single)
              
        else:
           if name != None:
              output= c.get_obj_time(fname, kind, obj_ns, name)
           else:
              output= c.get_obj_time(fname, kind, obj_ns, None)
        return output


    def collect_long_exp_metrics(self, path, namespace, watch_interval, numPods, bindingPolicy, wds_ctx=None, wec_ctx=None, label=None):
        try:
            self.load_kubeconfig(self.kubeconfig, self.current_ctx)
            v1 = client.CoreV1Api()
            w = watch.Watch()

            f1 = open(path + "/"  + str(numPods) + "-pod-wds.csv", "w")
            f2 = open(path + "/" + str(numPods) + "-manifestworks.csv", "w")
            f3 = open(path + "/" + str(numPods) + "-workstatus.csv", "w")
            f4 = open(path + "/" + str(numPods) + "-pod-wec.csv", "w")
            f5 = open(path + "/" + str(numPods) + "-appliedmanifestwork.csv", "w")

            c = 0
            i = 0
            for event in w.stream(v1.list_namespaced_pod, namespace=namespace, timeout_seconds=watch_interval):
                if event['type'] == "MODIFIED" or event['type'] == "ADDED":
                    if event['object'].metadata.deletion_grace_period_seconds != None or event['object'].metadata.deletion_timestamp != None:
                        continue

                    if event['object'].status.phase == "Succeeded":
                        i+=1
                        if i == 10:
                           c = c + i
                           i = 0
                           print("Number of pods COMPLETED: ", c)

                        # 1) Collect pod metrics:
                        name = event['object'].metadata.name
                        output=self.collect_obj_time(self.c1, name, 'pod', None, namespace)
                        f1.write(output + "\n")

                        output=self.collect_obj_time(self.c2, bindingPolicy + '-' + wds_ctx, 'cr', None, wec_ctx, "work.open-cluster-management.io", "v1", "manifestworks", bindingPolicy, True)
                        f2.write(output + "\n")

                        output=self.collect_obj_time(self.c2, 'v1-pod-' + namespace + '-' + name, 'cr', None, wec_ctx, "control.kubestellar.io", "v1alpha1", "workstatuses", bindingPolicy, True)
                        f3.write(output + "\n")

                        output=self.collect_obj_time(self.c3, name, 'pod', None, namespace)
                        f4.write(output + "\n")

                        output=self.collect_obj_time(self.c3, bindingPolicy + '-' + wds_ctx, 'cr', None, None, "work.open-cluster-management.io", "v1", "appliedmanifestworks", None, True)
                        f5.write(output + "\n")
                        
                        # 2) Delete pod after collection all metrics associated with it
                        namespace = event['object'].metadata.namespace
                        name = event['object'].metadata.name 
                        try:
                            api_response = v1.delete_namespaced_pod(name, namespace)
                        except ApiException as e:
                            print("Exception when calling CoreV1Api->delete_namespaced_pod: %s\n" % e)

                        if c == numPods:
                           print("Total number of pods COMPLETED: ", c)
                           w.stop()
                           break
            f1.close()
            f2.close()
            f3.close()
            f4.close()
            f5.close()
        except ApiException as e:
            print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)


    def collect_short_exp_metrics(self, path, num_ns, ns_prefix=None, wds_ctx=None, wec_ctx=None):

        obj_kind_list = ["deployments", "secrets", "configmaps", "services"]

        for i in range(0, num_ns):
            ns = ns_prefix + "-" + str(i)
            dir_ns = path + "/" + ns

            if not os.path.exists(dir_ns):
               os.makedirs(dir_ns)

            for kind in obj_kind_list:
                fname = dir_ns + "/" + kind
                self.collect_obj_time(self.c1, None, kind, fname + "-wds1", ns)
                self.collect_obj_time(self.c3, None, kind, fname + "-wec", ns)

            # collect manifestWork, workStatus and appliedManifestWork objects    
            bindingPolicy = ns
            self.collect_obj_time(self.c2, None, 'cr', dir_ns + "/workstatuses", wec_ctx, "control.kubestellar.io", "v1alpha1", "workstatuses", bindingPolicy)
            self.collect_obj_time(self.c2, None, 'cr', dir_ns + "/manifestworks", wec_ctx, "work.open-cluster-management.io", "v1", "manifestworks", bindingPolicy)
            self.collect_obj_time(self.c3, bindingPolicy + '-' + wds_ctx, 'cr', dir_ns + "/appliedmanifestworks", None, "work.open-cluster-management.io", "v1", "appliedmanifestworks", None, True, False)

if __name__=="__main__": 

    # common parameters
    kubeconfig = str(sys.argv[1]) # path to the kubeconfig file, e.g., "kscore-config"
    wds_ctx = str(sys.argv[2]) # name of the context for the target WDS, e.g., "wds1"
    its_ctx = str(sys.argv[3]) # name of the context for the target ITS, e.g., "its1"
    wec_ctx = str(sys.argv[4]) # name of the context for the target WEC, e.g., "cluster1"
    output_dir = str(sys.argv[6])  # path to the directory for the output files, e.g., $HOME/data/
    #exp_type = "long_duration"
    exp_type = "workload_generation"

    # Short duration experiments parameters
    ns_prefix = "perf-exp"  # prefix of the namespace where the workload are generated 
    num_ns = int(sys.argv[5]) # number of namespaces in created in your experiment, e.g., num_ns=6
   
    # Long duration experiments parameters
    numPods=2 # total number of workload objects set in your experiment
    watch_interval = 60 # (unit: seconds) watch for events that to occur within time interval threshold
    ns = "perf-exp-0" # name of the namespace where the workload are generated 
    bindingPolicy = "perf-exp-0" # name of the bindingPolicy associated with your workload

    c = DataParser(kubeconfig, wds_ctx, its_ctx, wec_ctx)
    if exp_type != "long_duration":
       c.collect_short_exp_metrics(output_dir, num_ns, ns_prefix, wds_ctx, wec_ctx)
    else:
       c.collect_long_exp_metrics(output_dir, ns, watch_interval, numPods, bindingPolicy, wds_ctx, wec_ctx)
