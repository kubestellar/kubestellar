#!/bin/bash

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

# Check if two parameters are provided
if [ $# -ne 2 ]; then
    echo "Usage: $0 <config file> <provider yaml>"
    exit 1
fi

config="$1"
provider="$2"

# Check if config file exists
if [ ! -f "$config" ]; then
    echo "Error: $config does not exist."
    exit 1
fi

# Check if provider file exists
if [ ! -f "$provider" ]; then
    echo "Error: $provider does not exist."
    exit 1
fi

# Define the specConfig string
specConfig="  Config: |"

# Check if config field exists in the provider yaml file and remove text after it
if grep -q "$specConfig" "$provider"; then
    sed -i "/$specConfig/,\$d" "$provider"
fi

# Add the field string to the provider file
echo "$specConfig" >> "$provider"

# Append the content of the config file to the provider file
sed 's/^/    /' "$config" >> "$provider"

echo "Config has been added to the provider yaml."
