#!/usr/bin/env bash

get_os_type() {
  case "$OSTYPE" in
      darwin*)  echo "darwin" ;;
      linux*)   echo "linux" ;;
      *)        echo "unknown: $OSTYPE" && exit 1 ;;
  esac
}
os_type=$(get_os_type)
echo $os_type

# process_running() {
#   SERVICE="$1"
#   if ps aux | grep "$SERVICE" >/dev/null
#   then
#       echo "running"
#   else
#       echo "stopped" 
#   fi
# }

# run_status=$(process_running mailbox-controller)
# echo $run_status

# clusters="florin guilder"

# # Deleting kind clusters
# for c in ${clusters[@]}
# do 
#   if [ $(kind get clusters | grep $c) > /dev/null 2>&1 ]; then
#      echo "kind cluster $c already exists - deleting it ...."
#      kind delete cluster --name $c > /dev/null 2>&1
#   fi
# done

# process_running() {
#   SERVICE="$1"
#   if pgrep -x "$SERVICE" >/dev/null
#   then
#       echo "running"
#   else
#       echo "stopped" 
#   fi
# }

# wait_for_process(){
#   status=$(process_running $1)
#   MAX_RETRIES=1
#   retries=0
#   status_code=0
#   while [ $status != "running" ]; do
#       if [ $retries -eq $MAX_RETRIES ]; then
#            status_code=1
#            break
#       fi

#       retries=$(( retries + 1 ))
#       sleep 3
#       status=$(process_running $1)
#   done
#   echo $status_code
# }

# run_status=$(wait_for_process mailbox-controller)
# #echo $run_status
# if [ $run_status -eq 0 ]; then
#     echo " mailbox-controller is running"
# else
#     echo " mailbox-controller failed to start ..... exiting"
#     sleep 2
#     exit
# fi

# message=$(wait_for_process mailbox-controller)
# echo $message

# if [ $message==1 ]; then
#    echo "Process $1 failed to start ..... exiting"
#    exit
# elif [ $message==0 ]; then
#    echo "Running"
# else
#    echo "Unknown code failure"
# fi 

# KCP is an older kcp-edge deployment is already running
# process_running() {
#   SERVICE="$1"
#   if pgrep -x "$SERVICE" >/dev/null
#   then
#       echo "running"
#   else
#       echo "stopped" 
#   fi
# }

# # Check kcp-edge is already running
# if [[ $(process_running kcp) == "running" || $(process_running mailbox-controller) == "running" || $(process_running placement-translator) == "running" || $(process_running main) == "running" ]] 
# then
#     echo "An older deployment of kcp-edge is already running - please stop it by running the deletion script: delete_edge-mc.sh "
#     exit
# fi

#pgrep -x kcp >/dev/null && echo "Process found" || echo "Process not found"

# process_running() {
#   SERVICE="$1"
#   if pgrep -x "$SERVICE" >/dev/null
#   then
#       echo "running"
#   else
#       echo "stopped" 
#   fi
# }

# if [ $(process_running kcp) == "running" ] 
# then
#     echo "KCP is already running - please stop it by running the deletion script: delete_edge-mc.sh "
#     exit
# fi

# test=$(process_running kcp)
# echo $test

# SERVICE="kcp"
# if pgrep -x "$SERVICE" >/dev/null
# then
#     echo "$SERVICE is running"
# else
#     echo "$SERVICE stopped"
#     # uncomment to start nginx if stopped
#     # systemctl start nginx
#     # mail  
# fi

# if [ -f "/Users/brauliodumba/Documents/Projects/edge-computing/edge-computing/challenge4900/kcp-edge/repo/edge-mc/environments/dev-env/kcp/bin/kubectl-kcp-playground" ] 
# then
#     echo "File found ..."
# else 
#     echo "File not found ..."
# fi

#set -e

# Check go version
# version=`go version | { read _ _ v _; echo ${v#go}; }`
# echo $version

# function ver { printf "%03d%03d%03d%03d" $(echo "$1" | tr '.' ' '); }

# if [ $go_version >= "1.19" ]
# then
#   echo "Go version must be 1.19 ...."
#   exit
# fi
# version_a="1.19.5"
# if [ $(ver $version_a) -lt $(ver 1.19) ]; then
#     echo "Update your go version"
# else
#     echo "Version is up-to-date"
# fi

# if ! docker ps > /dev/null
# then
#   echo "Docker Not running ...."
#   exit
# fi

# MAX_RETRIES=5
# retries=0

# fname="kcp/.kcp-playground/playground.kubeconfig"

# while [ $retries -le "$MAX_RETRIES" ]; do
#     echo $retries
#     retries=$(( retries + 1 ))

#     if [ -f $fname ]; then
#         break
#     fi

#     sleep 5
#     echo "Retries in do $retries"
# done
# echo "Finished"

# if ! docker ps
# then
#   echo "Not running"
# fi