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
