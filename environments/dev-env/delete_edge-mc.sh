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

pkill -f kubectl-kcp-playground
pkill -f kcp
pkill -f mailbox-controller
pkill -f placement-translator
pkill -f cmd/scheduler/main.go # edge-scheduler
rm -rf $(pwd)/kcp

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