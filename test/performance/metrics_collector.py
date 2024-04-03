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
from multiprocessing import Process
from kubernetes import client, config, watch
from kubernetes.client.rest import ApiException
from pprint import pprint
import urllib3
from kshelper import *
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


class DataParser():
    def __init__(self, num_objs, kubeconfig, ctx_1, ctx_2, ctx_3, f1, f2, f3, f4, f5):
        self.kubeconfig = kubeconfig
        self.context = ctx_1
        self.num_objects = num_objs
        self.c1 = Collector(kubeconfig, ctx_1, f1)
        self.c2 = Collector(kubeconfig, ctx_2, f2)
        self.c3 = Collector(kubeconfig, ctx_2, f3)
        self.c4 = Collector(kubeconfig, ctx_3, f4)
        self.c5 = Collector(kubeconfig, ctx_3, f5)

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

    def collect_metrics(self, c, name, kind, obj_ns=None, obj_grp=None, obj_v=None, obj_pl=None, check=False):
        if kind == "crd":
            c.get_crd_time(obj_grp, obj_v, obj_pl, name, check)

        elif kind == "pod":
            c.get_pod_time(obj_ns, name)

    def check_pod_status(self, namespace, watch_interval, label=None):
        try:
            self.load_kubeconfig(self.kubeconfig, self.context)
            v1 = client.CoreV1Api()
            w = watch.Watch()

            c = 0
            i = 0
            for event in w.stream(v1.list_pod_for_all_namespaces, timeout_seconds=watch_interval):
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
                        self.collect_metrics(self.c1, name, 'pod', namespace)
                        self.collect_metrics(self.c2, 'v1-pod-' + namespace + '-' + name, 'crd', '', "work.open-cluster-management.io", "v1", "manifestworks")
                        self.collect_metrics(self.c3, 'v1-pod-' + namespace + '-' + name, 'crd', '', "control.kubestellar.io", "v1alpha1", "workstatuses", True)
                        self.collect_metrics(self.c4, name, 'pod', namespace)
                        self.collect_metrics(self.c5, 'v1-pod-stress-test-' + name, 'crd', '', "work.open-cluster-management.io", "v1", "appliedmanifestworks", True)
                        
                        # 2) Delete pod after collection all metrics associated with it
                        namespace = event['object'].metadata.namespace
                        name = event['object'].metadata.name 
                        try:
                            api_response = v1.delete_namespaced_pod(name, namespace)
                        except ApiException as e:
                            print("Exception when calling CoreV1Api->delete_namespaced_pod: %s\n" % e)

                        if c == self.num_objects:
                           print("Total number of pods COMPLETED: ", c)
                           w.stop()
                           break
        except ApiException as e:
            print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)

if __name__=="__main__": 
  
    kubeconfig = "kscore-config"
    ns = "stress-test"
    mypath="dat/"
    run_ID="1"
    freq="1"
    numPods=20
    watch_interval = 3600
    method = "exp-RandomizedLoad"

    f1 = mypath + method + "-" + "freq-" + str(numPods) + "-pod-wds-kscore-stage-run" + run_ID + ".txt"
    f2 = mypath + method + "-" + "freq-" + str(numPods) + "-manifestwork-imbs1-run" + run_ID + ".txt"
    f3 = mypath + method + "-" + "freq-" + str(numPods) + "-workstatus-imbs1-run" + run_ID + ".txt"
    f4 = mypath + method + "-" + "freq-" + str(numPods) + "-pod-wec1-run" + run_ID + ".txt"
    f5 = mypath + method + "-" + "freq-" + str(numPods) + "-appliedmanifestwork-wec1-run" + run_ID + ".txt"

    c = DataParser(numPods, kubeconfig, 'wds1', 'imbs1', 'wec1', f1, f2, f3, f4, f5)
    c.check_pod_status(ns, watch_interval)
