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

# shellcheck shell=bash

# Usage: source this file, in a bash shell

# wait-for-cmd concatenates its arguments into one command that is iterated
# until it succeeds, failing if there is not success in 3 minutes.
# Note that the technology used here means that word splitting is done on
# the concatenation, and any quotes used by the caller to surround words
# do not get into here.
function wait-for-cmd() (
    cmd="$@"
    wait_counter=0
    while ! (eval "$cmd") ; do
        if (($wait_counter > 36)); then
            echo "Failed to ${cmd}."
            exit 1
        fi
        ((wait_counter += 1))
        sleep 5
    done
)

export -f wait-for-cmd

# expect-cmd-output takes two arguments:
# - a command to execute to produce some output, and
# - a command to test that output (received on stdin).
# expect-cmd-output executes the first command,
# echoes the output, and then applies the test.
function expect-cmd-output() {
    out=$(eval "$1")
    echo "$out"
    echo "$out" | $(eval "$2")
}

export -f expect-cmd-output

# wait-for-job-complete waits for a Kubernetes Job to complete with robust checks
# and on-timeout diagnostics.
# Usage: wait-for-job-complete <context> <namespace> <job-name> [timeout-seconds]
# - Considers the Job complete if either:
#   * .status.conditions has {type: Complete, status: True}, or
#   * .status.succeeded >= .spec.completions (defaults to 1 when not set), or
#   * any Pod for the Job has phase Succeeded (helps bridge condition propagation delays).
# - On timeout, prints useful diagnostics: describe job, pods, logs, and recent events.
function wait-for-job-complete() {
    local context="$1"
    local namespace="$2"
    local job_name="$3"
    local timeout_sec="${4:-400}"

    if [[ -z "$context" || -z "$namespace" || -z "$job_name" ]]; then
        echo "wait-for-job-complete: missing required arguments: <context> <namespace> <job-name> [timeout-seconds]" >&2
        return 1
    fi

    local interval=5
    local max_iters=$(( timeout_sec / interval ))
    local iter=0

    while true; do
        # Check Job existence and status; tolerate NotFound briefly.
        local job_json
        if ! job_json=$(kubectl --context "$context" -n "$namespace" get job "$job_name" -o json 2>/dev/null); then
            # Job not found yet; keep waiting until timeout
            :
        else
            # Check Complete condition
            local cond
            cond=$(echo "$job_json" | sed -n 's/.*\("type" *: *"Complete"[^]]*\).*/\1/p' | grep -o '"status" *: *"[^"]*"' | tail -n1 | awk -F'"' '{print $4}')
            if [[ "$cond" == "True" ]]; then
                return 0
            fi

            # Check succeeded vs completions
            local completions
            local succeeded
            completions=$(echo "$job_json" | sed -n 's/.*"completions" *: *\([0-9][0-9]*\).*/\1/p')
            succeeded=$(echo "$job_json" | sed -n 's/.*"succeeded" *: *\([0-9][0-9]*\).*/\1/p')
            [[ -z "$completions" ]] && completions=1
            [[ -z "$succeeded" ]] && succeeded=0
            if (( succeeded >= completions )); then
                return 0
            fi
        fi

        # Pod-level success heuristic (covers propagation delays)
        local pod_phases
        pod_phases=$(kubectl --context "$context" -n "$namespace" get pods -l job-name="$job_name" -o jsonpath='{range .items[*]}{.status.phase}{"\n"}{end}' 2>/dev/null || true)
        if echo "$pod_phases" | grep -q "^Succeeded$"; then
            # Brief stabilization wait to allow Job condition to catch up
            sleep 3
            return 0
        fi

        if (( iter >= max_iters )); then
            echo "Timed out waiting for Job $namespace/$job_name to complete in context $context after ${timeout_sec}s" >&2
            echo "--- job.describe ---" >&2
            kubectl --context "$context" -n "$namespace" describe job "$job_name" || true
            echo "--- job.yaml ---" >&2
            kubectl --context "$context" -n "$namespace" get job "$job_name" -o yaml || true
            echo "--- pods (wide) ---" >&2
            kubectl --context "$context" -n "$namespace" get pods -l job-name="$job_name" -o wide || true
            local pods
            pods=$(kubectl --context "$context" -n "$namespace" get pods -l job-name="$job_name" -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null || true)
            for p in $pods; do
                echo "--- logs: $p (all containers) ---" >&2
                kubectl --context "$context" -n "$namespace" logs "$p" --all-containers=true --tail=200 || true
            done
            echo "--- recent events ---" >&2
            kubectl --context "$context" -n "$namespace" get events --sort-by=.lastTimestamp | tail -n 200 || true
            return 1
        fi

        ((iter+=1))
        sleep "$interval"
    done
}

export -f wait-for-job-complete
