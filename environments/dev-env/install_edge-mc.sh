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

stage=""
clusters="florin guilder"
verbosity=0

while (( $# > 0 )); do
    if [ "$1" == "--stage" ]; then
        stage=$2
        shift
    elif [ "$1" == "--clusters" ]; then
        clusters=$2
        shift
    elif [ "$1" == "-v" ]; then
        verbosity=1
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

# Deleting kind clusters
for c in ${clusters[@]}
do 
  if [ $(kind get clusters | grep $c) > /dev/null 2>&1 ]; then
     echo "kind cluster $c already exists - deleting it ...."
     kind delete cluster --name $c > /dev/null 2>&1
  fi
done

# KCP is an older kcp-edge deployment is already running
process_running() {
  SERVICE="$1"
  if pgrep -f "$SERVICE" >/dev/null
  then
      echo "running"
  else
      echo "stopped" 
  fi
}

# Check kcp-edge is already running
if [ $(process_running kcp) == "running" ]
then
    echo "An older deployment of kcp-playground is already running - terminating it ...."
    pkill -f kubectl-kcp-playground
    pkill -f kcp
fi

# Check mailbox-controller is already running
if [ $(process_running mailbox-controller) == "running" ]
then
    echo "An older deployment of mailbox-controller is already running - terminating it ...."
    pkill -f mailbox-controller
fi

# Check edge-scheduler is already running
if [ $(process_running cmd/scheduler/main.go) == "running" ]
then
    echo "An older deployment of edge-scheduler is already running - terminating it ...."
    #ps xu | grep scheduler/main.go | grep -v grep | awk '{ print $2 }' | xargs kill -9
    pkill -f cmd/scheduler/main.go
fi

# Check placement-translator is already running
if [ $(process_running placement-translator) == "running" ]
then
    echo "An older deployment of placement-translator is already running - terminating it ...."
    pkill -f placement-translator
fi



#(1): Clone the kcp-playground tool
echo "****************************************"
echo "Clonining edge-syncer kcp-plugins repo"
echo "****************************************"
if [[ ! -d $(pwd)/edge-syncer || ! "$(ls -A $(pwd)/edge-syncer)" ]] 
then
    if [ $verbosity == 1 ]; then
        git clone -b emc https://github.com/yana1205/kcp  edge-syncer
    else
        git clone -b emc https://github.com/yana1205/kcp  edge-syncer  > /dev/null 2>&1
        echo "Finished cloning repo ......"
    fi
else 
   echo "   edge-syncer repo already exists ..."
fi


echo "****************************************"
echo "Clonining kcp-playground repo"
echo "****************************************"

if [[ ! -d $(pwd)/kcp || ! "$(ls -A $(pwd)/kcp)" ]] 
then
    if [ $verbosity == 1 ]; then
        git clone -b kcp-playground https://github.com/fabriziopandini/kcp.git
    else
        git clone -b kcp-playground https://github.com/fabriziopandini/kcp.git  > /dev/null 2>&1
        echo "Finished cloning repo ......"
    fi
else 
   echo "  kcp-playground repo already exists ..."
fi

#(2): Move the edge-syncer plugin to the kcp-playground repo
cp -r edge-syncer/pkg/cliplugins/workload/*   kcp/pkg/cliplugins/workload

#(3): Move the kcp-playground yaml files to the target repo
stages_path=$(pwd)/kcp/test/kubectl-kcp-playground/examples/kcp-edge/
if [ ! -d $stages_path ] 
then
    mkdir $stages_path
fi

cp stages/*  $stages_path

#(4): build the binaries for kcp and kcp-playground plugin
echo "****************************************"
echo "Building kcp-playground binaries"
echo "****************************************"
cd $(pwd)/kcp

if [ -f "bin/kubectl-kcp-playground" ] 
then
    echo "  kcp-playground binaries already exists ..."
else 
    if [ $verbosity == 1 ]; then
        make build
    else
        echo "Started building the binaries ..."
        make build  > /dev/null 2>&1
        echo "Finshed building the binaries ..."
    fi
fi

#(5): Set up the kubeconfig and path variables
kcp_path=$(pwd)/bin
kubeconfig_path=$(pwd)/.kcp-playground/playground.kubeconfig
export PATH=$PATH:$kcp_path
export KUBECONFIG=$kubeconfig_path

#(6): Start the kcp-playground
echo "****************************************"
echo "Started deploying kcp-playground: complete in ~ 150 sec (maximum waiting time: 300 sec)"
echo "****************************************"
rm -rf .kcp-playground/

if [ $stage == 0 ]; then
    kubectl kcp playground start -f test/kubectl-kcp-playground/examples/kcp-edge/poc2023q1-BYOW.yaml >& ../kcp-playground-log.txt &
elif [ $stage == 1 ]; then  
    kubectl kcp playground start -f test/kubectl-kcp-playground/examples/kcp-edge/poc2023q1-stage1.yaml >& ../kcp-playground-log.txt &
elif [ $stage == 2 ]; then
    kubectl kcp playground start -f test/kubectl-kcp-playground/examples/kcp-edge/poc2023q1-stage2.yaml >& ../kcp-playground-log.txt &
elif [ $stage == 3 ]; then
    kubectl kcp playground start -f test/kubectl-kcp-playground/examples/kcp-edge/poc2023q1-stage2.yaml >& ../kcp-playground-log.txt &
elif [ $stage == 4 ]; then
    kubectl kcp playground start -f test/kubectl-kcp-playground/examples/kcp-edge/poc2023q1-stage2.yaml >& ../kcp-playground-log.txt &
fi 

#####################################################
MAX_RETRIES=20 # maximum wait: 5 minutes
retries=0
sucess=1

fname=".kcp-playground/playground.kubeconfig"

while [ $retries -le "$MAX_RETRIES" ]; do
    #echo $retries
    retries=$(( retries + 1 ))

    if [ -f $fname ]; then
        sucess=0
        break
    fi

    sleep 15
    sec=$(( retries * 15 ))
    echo "$sec sec"
done

if [ $sucess == 1 ]; then
   echo "kcp-playground kubeconfig not generated - maximum waiting time exceeded: 300 sec"
   exit 
fi 
####################################################
echo "****************************************"
echo "Finished deploying kcp-playground .... (log file: kcp-playground-log.txt)"
echo "****************************************"

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
echo "Started deploying kCP-EDGE controllers ...."
echo "****************************************"
cd ../../..

# Delete default location object in the inventory workspace
if [ $stage -gt 0 ]; then
    kubectl ws imw-1
    kubectl delete location default > /dev/null 2>&1
fi


kubectl ws root:espw
go run ./cmd/mailbox-controller --inventory-context=shard-main-root --mbws-context=shard-main-base -v=2 >& environments/dev-env/mailbox-controller-log.txt &

run_status=$(wait_for_process mailbox-controller)
if [ $run_status -eq 0 ]; then
    echo " mailbox-controller is running (log file: mailbox-controller-log.txt)"
else
    echo " mailbox-controller failed to start ..... exiting"
    sleep 2
    exit
fi


if [ $stage != 1 ]; then 
    # (8): Start the edge-mc scheduler
    sleep 3
    kubectl ws root:espw
    go run cmd/scheduler/main.go -v 2 --root-user shard-main-kcp-admin  --root-cluster shard-main-root  --sysadm-context shard-main-system:admin  --sysadm-user shard-main-shard-admin >& environments/dev-env/edge-scheduler-log.txt &
    message=$(wait_for_process  cmd/scheduler/main.go)
    
    run_status=$(wait_for_process main)
    if [ $run_status -eq 0 ]; then
        echo " scheduler is running (log file: edge-scheduler-log.txt)"
    else
        echo " scheduler failed to start ..... exiting"
        exit
    fi
fi


if [ $stage -eq 0 ] || [ $stage -gt 2 ]; then 
    # (9): Start the Placement Translator
    sleep 3
    kubectl ws root:espw
    go run ./cmd/placement-translator --allclusters-context  "shard-main-system:admin" -v=2 >& environments/dev-env/placement-translator-log.txt &

    run_status=$(wait_for_process placement-translator)
    if [ $run_status -eq 0 ]; then
        echo " placement translator is running (log file: placement-translator-log.txt)"
    else
        echo " placement translator failed to start ..... exiting"
        exit
    fi
fi

sleep 10

echo "****************************************"
echo "Finished deploying kCP-EDGE controllers ...."
echo "****************************************"


if [ $stage -gt 2 ]; then
    # (10): Create syncers manifest and apply it to the edge pcluters
    echo "****************************************"
    echo "Started deploying kCP-EDGE syncer ...."
    echo "****************************************"

    cd environments/dev-env/
    
    if [ $verbosity == 1 ]; then
        ./build-edge-syncer.sh --syncTarget  sync-target-f -v
        ./build-edge-syncer.sh --syncTarget  sync-target-g -v
    else
        ./build-edge-syncer.sh --syncTarget  sync-target-f
        ./build-edge-syncer.sh --syncTarget  sync-target-g     
    fi

    cd kcp/

    if [ $verbosity == 1 ]; then
        # pcluster florin
        kubectl kcp playground use pcluster florin 
        kubectl apply -f ../sync-target-f-syncer.yaml

        # pcluster guilder
        kubectl kcp playground use pcluster guilder
        kubectl apply -f ../sync-target-g-syncer.yaml
    else
        kubectl kcp playground use pcluster florin > /dev/null 2>&1
        kubectl apply -f ../sync-target-f-syncer.yaml > /dev/null 2>&1

        # pcluster guilder
        kubectl kcp playground use pcluster guilder > /dev/null 2>&1
        kubectl apply -f ../sync-target-g-syncer.yaml  > /dev/null 2>&1
    fi


    if [ $verbosity == 1 ]; then
       kubectl kcp playground use shard main # switch to kcp shard
    else
       kubectl kcp playground use shard main > /dev/null 2>&1
    fi
    
    echo "****************************************"
    echo "Finished deploying kCP-EDGE syncer ...."
    echo "****************************************"
fi

kubectl ws root
export PATH=$PATH:$(pwd)/kcp/bin
export KUBECONFIG=$(pwd)/kcp/.kcp-playground/playground.kubeconfig

echo "KCP-Edge dev-env successfully started"
echo "To start using the KCP-Edge dev-env: "
echo '   export KUBECONFIG="$(pwd)/kcp/.kcp-playground/playground.kubeconfig"'
echo '   export PATH="$PATH:$(pwd)/kcp/bin"'
