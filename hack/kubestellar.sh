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

# Usage: $0 (start or stop | -v)

# Purpose: deploy the kubestellar platform. The following components are created:
#           (a) 1 kcp workspace: edge service provider workspace (espw)
#           (b) 3 kubestellar controllers: edge-scheduler, mailbox-controller and placement-translator

# Assumption: kcp server is running.

# Requirements:
#    KubeStellar bin sists the edge-mc binaries mailbox-controller, scheduler, placement-translator
#    KubeStellar controller binaries are on $PATH.


set -e

install=""
verbosity=0
espw_name="espw"
log_folder="$PWD/kubestellar-logs"

while (( $# > 0 )); do
    case "$1" in
    (start)
        install=1;; 
    (stop) 
        install=0;; 
    (--log-folder)
        if (( $# > 1 ));
        then { log_folder="$2"; shift; }
        else { echo "$0: missing log folder" >&2; exit 1; }
        fi;;
    (--verbose|-V)
        verbosity=1;;
    (-h|--help)
        echo "Usage: $0 [start| stop][--log-folder log_folder] [-V|--verbose]"
        exit 0;;
    (-*)
        echo "$0: unknown flag" >&2 ; exit 1;
        exit 1;;
    (*)
        echo "$0: unknown positional argument" >&2; exit 1;
        exit 1;;
    esac
    shift
done


# Check if docker is running
if ! docker ps > /dev/null
then
  echo "Docker Not running ...."
  exit
fi

# Check go version
go_version=`go version | { read _ _ v _; echo ${v#go}; }`

function ver { printf "%03d%03d%03d%03d" $(echo "$1" | tr '.' ' '); }
if [ $(ver $go_version) -lt $(ver 1.19) ]; then
    echo "Update your go version to at least 1.19"
fi

# Check if a given process name is running
process_running() {
  SERVICE="$1"
  if pgrep -f "$SERVICE" >/dev/null
  then
      echo "running"
  else
      echo "stopped" 
  fi
}

# Check if kcp is running
if [ $(process_running kcp) != "running" ]
then
    echo "kcp is not running - please start it ...."
    exit
fi

# Check mailbox-controller is already running
if [ $(process_running mailbox-controller) == "running" ]
then
    echo "An older deployment of mailbox-controller is already running - terminating it ...."
    pkill -f mailbox-controller
fi

# Check edge-scheduler is already running
if [ $(process_running "scheduler -v 2") == "running" ]
then
    echo "An older deployment of edge-scheduler is already running - terminating it ...."
    pkill -f "scheduler -v 2"
fi

# Check placement-translator is already running
if [ $(process_running placement-translator) == "running" ]
then
    echo "An older deployment of placement-translator is already running - terminating it ...."
    pkill -f placement-translator
fi

kubectl ws root &> /dev/null

if [ $install -eq 0 ]; then 
   if kubectl get Workspace "$espw_name" &> /dev/null; then
      kubectl delete ws "$espw_name"
      echo "Deleted workspace: $espw_name"
   fi
   echo "kcp edge stopped ....."
   exit
fi


wait_for_process(){
  status=$(process_running $1)
  MAX_RETRIES=5
  retries=0
  status_code=0
  while [ $status != "running" ]; do
      if [ $retries -eq $MAX_RETRIES ]; then
           status_code=1
           break
      fi

      retries=$(( retries + 1 ))
      sleep 3
      status=$(process_running $1)
  done
  echo $status_code
}


# Start the edge-mc controllers
echo "****************************************"
echo "Launching KubeStellar ..."
echo "****************************************"


if kubectl get Workspace "$espw_name" &> /dev/null; then
   echo "espw workspace already exists -- using it:"
   kubectl ws "$espw_name"
else 
   if [ $verbosity == 1 ]; then
        kubectl ws create "$espw_name" --enter
        kubectl apply -f  ./config/crds 
        kubectl apply -f  ./config/exports
        echo "Finished populate the espw with kcp edge crds and apiexports"
   else
        kubectl ws create "$espw_name" --enter
        kubectl apply -f  ./config/crds &> /dev/null
        kubectl apply -f  ./config/exports &> /dev/null
        echo "Finished populate the espw with kcp edge crds and apiexports"
   fi
fi

sleep 5

# Create the logs directory
if [[ ! -d $log_folder ]]; then
    mkdir -p "$log_folder"
fi

mailbox-controller -v=2 >& $log_folder/mailbox-controller-log.txt &

run_status=$(wait_for_process mailbox-controller)
if [ $run_status -eq 0 ]; then
    echo " mailbox-controller is running (log file: mailbox-controller-log.txt)"
else
    echo " mailbox-controller failed to start ..... exiting"
    sleep 2
    exit
fi


# Start the edge-mc scheduler
sleep 3
scheduler -v 2 >& $log_folder/edge-scheduler-log.txt &

run_status=$(wait_for_process "scheduler -v 2")
if [ $run_status -eq 0 ]; then
    echo " scheduler is running (log file: edge-scheduler-log.txt)"
else
    echo " scheduler failed to start ..... exiting"
    exit
fi
 
# Start the Placement Translator
sleep 3
placement-translator --allclusters-context  "system:admin" -v=2 >& $log_folder/placement-translator-log.txt &

run_status=$(wait_for_process placement-translator)
if [ $run_status -eq 0 ]; then
    echo " placement translator is running (log file: placement-translator-log.txt)"
else
    echo " placement translator failed to start ..... exiting"
    exit
fi

sleep 10
echo "****************************************"
echo "Finished launching KubeStellar ..."
echo "****************************************"
kubectl ws root

