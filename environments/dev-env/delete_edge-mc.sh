#!/usr/bin/env bash

clusters="florin guilder"

while (( $# > 0 )); do
    if [ "$1" == "--clusters" ]; then
        clusters=$2
    fi 
    shift
done

# Deleting kind clusters
for c in ${clusters[@]}
do 
  if [ $(kind get clusters | grep $c) > /dev/null 2>&1 ]; then
     echo "Deleting kind cluster $c ...."
     kind delete cluster --name $c > /dev/null 2>&1
  fi
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

if [ $os_type == "darwin" ]; then
    pkill kubectl-kcp-playground
    pkill kcp
    pkill mailbox-controller
    pkill placement-translator
    pkill main # edge-scheduler
    rm -rf $(pwd)/kcp

elif [ $os_type == "linux" ]; then
    kill -9 $(pidof kubectl-kcp-playground)
    kill -9 $(pidof kcp)
    kill -9 $(pidof mailbox-controller)
    pkill -f  shard-main-shard-admin # edge-scheduler
    kill -9 $(pidof placement-translator)
    rm -rf $(pwd)/kcp
fi 

if [ -f "placement-translator-log.txt" ]; then
      rm placement-translator-log.txt
      echo "Deleted log file: placement-translator-log.txt"
fi 

if [ -f "edge-scheduler-log.txt" ]; then
      rm edge-scheduler-log.txt
      echo "Deleted log file: edge-scheduler-log.txt"
fi

if [ -f "mailbox-controller-log.txt" ]; then
      rm mailbox-controller-log.txt
      echo "Deleted log file: mailbox-controller-log.txt"
fi


if [ -f "kcp-playground-log.txt" ]; then
      rm kcp-playground-log.txt
      echo "Deleted log file: kcp-playground-log.txt"
fi

rm -rf $(pwd)/kcp

echo "Finished deletion ...."