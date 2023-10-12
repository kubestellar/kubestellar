#!/usr/bin/env bash

# Copyright 2023 The KubeStellar Authors.
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


# This is only a place holder once the integration with helm is done this file will be replaced
set -e

function echoerr() {
   echo "ERROR: $1" >&2
}

function run_space_manager() {
    echo "--< Starting space-manager >--"
    if ! space_manager -v=${VERBOSITY} ; then
        echoerr "unable to start space-manager!"
        exit 1
    fi
}

echo "--< Starting SpaceManager container >--"

echo "Environment variables:"
if [ $# -ne 0 ] ; then
    ACTION="$1"
else
    ACTION="sleep"
fi
echo "ACTION=${ACTION}"
if [ "$VERBOSITY" == "" ]; then
    VERBOSITY="2"
fi

echo "VERBOSITY=${VERBOSITY}"

case "${ACTION}" in
(space-manager)
    run_space_manager;;
(sleep)
    echo "Nothing to do... sleeping forever."
    sleep infinity;;
(*)
    echoerr "unknown action '$1'!"
    exit 1;;
esac
