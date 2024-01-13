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

set -o errexit
set -o nounset
set -o pipefail

if [ $# != 2 ]; then
    echo "Usage: $0 \$script_working_directory \$HTML_file_pathname" >&2
    exit 1
fi

wdir="$(cd "$1" && pwd)"
shift
html_file="$1"
generated_script_file="./generated_script.sh"
generated_script_suffix_file="generated_script_suffix"

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

"${SCRIPT_DIR}/extract-bash.py" "$html_file" "$generated_script_suffix_file"
cat > "$generated_script_file" <<EOF
cd '${wdir}'
set -o errexit
set -o nounset
set -o pipefail
set -o xtrace
EOF
cat "$generated_script_suffix_file" >> "$generated_script_file"
rm "$generated_script_suffix_file"

echo
echo "Generated script file follows"
cat "$generated_script_file"

echo
echo "Execution follows"

# make the generated script executable
chmod +x "$generated_script_file"

# run the generated script
"$generated_script_file"

rm $generated_script_file
