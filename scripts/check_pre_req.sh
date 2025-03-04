#!/usr/bin/env bash
set -e

echo "Creating sysctl-check Job..."
kubectl apply -f sysctl-check.yaml

echo "Waiting for Job to complete..."
kubectl wait --for=condition=complete job/sysctl-check

echo "Fetching Job logs..."
logs=$(kubectl logs job/sysctl-check)

check_sysctl_value() {
  name=$1
  minval=$2
  value=$(echo "$logs" | grep "$name" | awk '{print $3}')
  if [ "$value" -ge "$minval" ]; then
    echo -e "\033[0;32m✔\033[0m $name is $value"
  else
    echo -e "\033[0;31mX\033[0m $name is $value, must be at least $minval"
    exit 1
  fi
}

check_sysctl_value "fs.inotify.max_user_watches" 524288
check_sysctl_value "fs.inotify.max_user_instances" 512

echo "Sysctl settings are correct."
