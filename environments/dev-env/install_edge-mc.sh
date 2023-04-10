#!/usr/bin/env bash

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


# Find os type (supported: linux and darwin)
get_os_type() {
  case "$OSTYPE" in
      darwin*)  echo "darwin" ;;
      linux*)   echo "linux" ;;
      *)        echo "unknown: $OSTYPE" && exit 1 ;;
  esac
}
os_type=$(get_os_type)

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
  if pgrep -x "$SERVICE" >/dev/null
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
    if [ $os_type == "darwin" ]; then
        pkill kubectl-kcp-playground
        pkill kcp
    elif [ $os_type == "linux" ]; then
        kill -9 $(pidof kubectl-kcp-playground)
        kill -9 $(pidof kcp)
    fi 
fi

# Check mailbox-controller is already running
if [ $(process_running mailbox-controller) == "running" ]
then
    echo "An older deployment of mailbox-controller is already running - terminating it ...."
    if [ $os_type == "darwin" ]; then
        pkill mailbox-controller
    elif [ $os_type == "linux" ]; then
        kill -9 $(pidof mailbox-controller)
    fi 
fi

# Check edge-scheduler is already running
if [ $(process_running main) == "running" ]
then
    echo "An older deployment of edge-scheduler is already running - terminating it ...."
    #ps xu | grep scheduler/main.go | grep -v grep | awk '{ print $2 }' | xargs kill -9
    pkill -f  shard-main-shard-admin
fi

# Check placement-translator is already running
if [ $(process_running placement-translator) == "running" ]
then
    echo "An older deployment of placement-translator is already running - terminating it ...."
    if [ $os_type == "darwin" ]; then
        pkill placement-translator
    elif [ $os_type == "linux" ]; then
        kill -9 $(pidof placement-translator)
    fi 
fi



#(1): Clone the kcp-playground tool
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

#(2): Move the kcp-playground yaml files to the target repo
stages_path=$(pwd)/kcp/test/kubectl-kcp-playground/examples/kcp-edge/
if [ ! -d $stages_path ] 
then
    mkdir $stages_path
fi

cp stages/*  $stages_path

#(3): build the binaries for kcp and kcp-playground plugin
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

#(4): Set up the kubeconfig and path variables
kcp_path=$(pwd)/bin
kubeconfig_path=$(pwd)/.kcp-playground/playground.kubeconfig
export PATH=$PATH:$kcp_path
export KUBECONFIG=$kubeconfig_path

echo $kcp_path
#echo $kubeconfig_path


#(5): Start the kcp-playground
echo "****************************************"
echo "Started deploying kcp-playground: complete in ~ 150 sec (maximum waiting time: 300 sec)"
echo "****************************************"
rm -rf .kcp-playground/

if [ $stage == 1 ]; then  
    kubectl kcp playground start -f test/kubectl-kcp-playground/examples/kcp-edge/poc2023q1-stage1.yml >& ../kcp-playground-log.txt &
elif [ $stage == 2 ]; then
    kubectl kcp playground start -f test/kubectl-kcp-playground/examples/kcp-edge/poc2023q1-stage2.yml >& ../kcp-playground-log.txt &
elif [ $stage == 3 ]; then
    kubectl kcp playground start -f test/kubectl-kcp-playground/examples/kcp-edge/poc2023q1-stage2.yml >& ../kcp-playground-log.txt &
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


#(6): Start the edge-mc controller
echo "****************************************"
echo "Started deploying kCP-EDGE controllers ...."
echo "****************************************"
cd ../../..

# Delete default location object in the inventory workspace
kubectl ws imw-1
kubectl delete location default > /dev/null 2>&1

kubectl ws root:espw
go run ./cmd/mailbox-controller --inventory-context=shard-main-root -v=2 >& environments/dev-env/mailbox-controller-log.txt &

run_status=$(wait_for_process mailbox-controller)
if [ $run_status -eq 0 ]; then
    echo " mailbox-controller is running (log file: mailbox-controller-log.txt)"
else
    echo " mailbox-controller failed to start ..... exiting"
    sleep 2
    exit
fi


if [ $stage -gt 1 ]; then 
    # (7): Start the edge-mc scheduler
    sleep 3
    kubectl ws root:espw
    go run cmd/scheduler/main.go -v 2 --root-user shard-main-kcp-admin  --root-cluster shard-main-root  --sysadm-context shard-main-system:admin  --sysadm-user shard-main-shard-admin >& environments/dev-env/edge-scheduler-log.txt &
    message=$(wait_for_process main)
    
    run_status=$(wait_for_process main)
    if [ $run_status -eq 0 ]; then
        echo " scheduler is running (log file: edge-scheduler-log.txt)"
    else
        echo " scheduler failed to start ..... exiting"
        exit
    fi
fi


if [ $stage -gt 2 ]; then 
    # (8): Start the Placement Translator
    sleep 3
    kubectl ws root:espw
    go run ./cmd/placement-translator --allclusters-context  "shard-main-system:admin" >& environments/dev-env/placement-translator-log.txt &

    run_status=$(wait_for_process placement-translator)
    if [ $run_status -eq 0 ]; then
        echo " placement translator is running (log file: placement-translator-log.txt)"
    else
        echo " placement translator failed to start ..... exiting"
        exit
    fi
fi

sleep 5
kubectl ws root

echo "****************************************"
echo "Finished deploying kCP-EDGE controllers ...."
echo "****************************************"

export PATH=$PATH:$(pwd)/kcp/bin
export KUBECONFIG=$(pwd)/kcp/.kcp-playground/playground.kubeconfig

echo "KCP-Edge dev-env successfully started"
echo "To start using the KCP-Edge dev-env: "
echo "   export KUBECONFIG=kcp/.kcp-playground/playground.kubeconfig"