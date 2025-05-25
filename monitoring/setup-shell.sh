# Copyright 2024 The KubeStellar Authors.
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

# Usage: source this file, in a bash shell

function wait-for-object() {

if [ "$#" -ne 4 ]; then
  echo "Usage: $0 <context> <namespace> <object-kind (e.g, deployment or statefulset)> <object-name>"
  exit 1
fi    

CONTEXT=$1
NAMESPACE=$2
OBJECT_KIND=$3
DEPLOYMENT_NAME=$4  
SLEEP_SECONDS=3 # How long to sleep between checks

while :; do
    # Check if the deployment exists
    if kubectl --context ${CONTEXT} get $OBJECT_KIND "$DEPLOYMENT_NAME" -n "$NAMESPACE" > /dev/null 2>&1; then
        echo "Checking deployment $DEPLOYMENT_NAME..."

        # Retrieve the number of ready replicas and desired replicas for the deployment
        READY_REPLICAS=$(kubectl  --context ${CONTEXT} get $OBJECT_KIND "$DEPLOYMENT_NAME" -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}')
        REPLICAS=$(kubectl  --context ${CONTEXT} get $OBJECT_KIND "$DEPLOYMENT_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.replicas}')

        # Check if READY_REPLICAS is unset or empty, setting to zero if it is
        [ -z "$READY_REPLICAS" ] && READY_REPLICAS=0

        # Compare the number of ready replicas with the desired number of replicas
        if [ "$READY_REPLICAS" -eq "$REPLICAS" ]; then
            echo "All replicas are ready."
            break
        else
            echo "Ready replicas ($READY_REPLICAS) do not match the desired count of replicas ($REPLICAS)."
        fi
    else
        echo "$OBJECT_KIND $DEPLOYMENT_NAME does not exist in the namespace $NAMESPACE and context $CONTEXT"
    fi
    
    # Sleep before checking again
    echo "Sleeping for $SLEEP_SECONDS seconds..."
    sleep "$SLEEP_SECONDS"
done
}

export -f wait-for-object


function wait-for-cmd() (
    cmd="$@"
    wait_counter=0
    while ! (eval "$cmd") ; do
        if (($wait_counter > 10)); then
            echo "Failed to ${cmd}."
            exit 1
        fi
        ((wait_counter += 1))
        sleep 5
    done
)

export -f wait-for-cmd
