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

set -e

install=""
verbosity=0
espw_name="espw"
user_type="dev"

while (( $# > 0 )); do
    if [ "$1" == "start" ]; then
        install=1
    elif [ "$1" == "stop" ]; then
        install=0
    elif [ "$1" == "-v" ]; then
        verbosity=1
    elif [ "$1" == "--user" ]; then
        user_type=$2
    fi 
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
    echo "Update your go version"
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


#(7): Start the edge-mc controller
echo "****************************************"
echo "Started deploying kCP-EDGE infra ...."
echo "****************************************"


if kubectl get Workspace "$espw_name" &> /dev/null; then
   echo "espw workspace already exists -- using it:"
   kubectl ws "$espw_name"
else 
   if [ $verbosity == 1 ]; then
        kubectl ws create "$espw_name" --enter

        if [ $user_type == "dev" ]; then
            kubectl apply -f  ../../config/crds 
            kubectl apply -f  ../../config/exports
            echo "Finished populate the espw with kcp edge crds and apiexports"

        elif [ $user_type == "kit" ]; then
            # Apply CRDs
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_customizers.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_edgeplacements.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_edgesyncconfigs.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_singleplacementslices.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_syncerconfigs.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/meta.kcp.io_apiresources.yaml

            # Apply Exports
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiexport-edge.kcp.io.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiexport-meta.kcp.io.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-apiresources.meta.kcp.io.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-customizers.edge.kcp.io.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-edgeplacements.edge.kcp.io.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-edgesyncconfigs.edge.kcp.io.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-singleplacementslices.edge.kcp.io.yaml
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-syncerconfigs.edge.kcp.io.yaml
            echo "Finished populate the espw with kcp edge crds and apiexports"
        else
            echo "Unknown user type ..."
            exit 1
        fi 
   else
        kubectl ws create "$espw_name" --enter

        if [ $user_type == "dev" ]; then
            kubectl apply -f  ../../config/crds 
            kubectl apply -f  ../../config/exports
            echo "Finished populate the espw with kcp edge crds and apiexports"

        elif [ $user_type == "kit" ]; then
            # Apply CRDs
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_customizers.yaml  &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_edgeplacements.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_edgesyncconfigs.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_singleplacementslices.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/edge.kcp.io_syncerconfigs.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/crds/meta.kcp.io_apiresources.yaml &> /dev/null

            # Apply Exports
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiexport-edge.kcp.io.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiexport-meta.kcp.io.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-apiresources.meta.kcp.io.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-customizers.edge.kcp.io.yaml  &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-edgeplacements.edge.kcp.io.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-edgesyncconfigs.edge.kcp.io.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-singleplacementslices.edge.kcp.io.yaml &> /dev/null
            kubectl apply -f https://raw.githubusercontent.com/kcp-dev/edge-mc/main/config/exports/apiresourceschema-syncerconfigs.edge.kcp.io.yaml &> /dev/null
            echo "Finished populate the espw with kcp edge crds and apiexports"
        else
            echo "Unknown user type ..."
            exit 1
        fi 
   fi
fi

sleep 5

mailbox-controller --inventory-context=root --mbws-context=base -v=2 >& mailbox-controller-log.txt &

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
scheduler -v 2 --root-user kcp-admin  --root-cluster root  --sysadm-context system:admin  --sysadm-user shard-admin >& edge-scheduler-log.txt &

run_status=$(wait_for_process "scheduler -v 2")
if [ $run_status -eq 0 ]; then
    echo " scheduler is running (log file: edge-scheduler-log.txt)"
else
    echo " scheduler failed to start ..... exiting"
    exit
fi
 
# Start the Placement Translator
sleep 3
placement-translator --allclusters-context  "system:admin" -v=2 >& placement-translator-log.txt &

run_status=$(wait_for_process placement-translator)
if [ $run_status -eq 0 ]; then
    echo " placement translator is running (log file: placement-translator-log.txt)"
else
    echo " placement translator failed to start ..... exiting"
    exit
fi

sleep 10
echo "****************************************"
echo "Finished deploying kCP-EDGE controllers ...."
echo "****************************************"
kubectl ws root