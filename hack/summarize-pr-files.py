#!/usr/bin/env python3

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

"""PR-to-files pretty printer

This file takes a PR-to-files map as input and outputs two views of it
in Markdown. A PR-to-files map is a dict where the key is a URL for
the HTML view of a PR, such as
"https://github.com/kubestellar/kubestellar/pull/3047", and the value
is a list of filenames. One view is the reverse map, from file to list
of PRs that touch that file. This view is filtered to only cover files
that are touched by more than one PR. The other view is the forward
view, listing for each PR the files that it touches.

The input is the JSON or Python (they are the same) representation of
the PR-to-files map. For example, one can be produced by the following
`bash` command (which will issue N+1 calls on the GitHub API, where N
is the number of open PRs that are not marked as drafts (works in
progress)).

(
echo '{"": []'
gh api --paginate -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/pulls | jq -r '.[] | select(.draft | not) | .html_url' | while read pullurl; do
    pullnum=${pullurl##*/}
    echo ", \"${pullurl}\": "
    gh api --paginate -H "X-GitHub-Api-Version: 2022-11-28" "/repos/kubestellar/kubestellar/pulls/${pullnum}/files" | jq 'map(.filename)'
done
echo '}'
)

Oh yeah, the entry for the empty PR url is ignored.

"""

import json
import sys

input_json = json.load(sys.stdin)
reversed = dict()
for pullurl, files in input_json.items():
    if len(pullurl) == 0: continue
    for file in files:
        reversed[file] = [pullurl] + reversed.get(file, [])

files = list(reversed.keys())
files.sort()

print('# By file')
for file in files:
    pulls = reversed[file]
    if len(pulls) < 2: continue
    print('## ' + file)
    pulls.sort()
    print(', '.join(pulls))

print('# By PR')
for pullurl, files in input_json.items():
    if len(pullurl) == 0: continue
    parts = pullurl.split('/')
    pullnum = parts[-1]
    print('## [' + pullnum +'](' + pullurl + ')')
    files.sort()
    for file in files:
        print('- ' + file)
