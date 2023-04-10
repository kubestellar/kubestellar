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


pkill kubectl-kcp-playground
pkill kcp
pkill mailbox-controller
pkill placement-translator
pkill main # edge-scheduler
rm -rf $(pwd)/kcp

rm placement-translator-log.txt
rm edge-scheduler-log.txt
rm mailbox-controller-log.txt
rm kcp-playground-log.txt
rm -rf $(pwd)/kcp

echo "Finished deletion ...."