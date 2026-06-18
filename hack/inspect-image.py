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

from __future__ import annotations
import argparse
import json
import re
import sys
import typing
import urllib.parse
import urllib.request

"""
This script is given a container image repository reference and looks up all
the tags in that repository and fetches and prints the associated manifest
for each tag. The output is suitable as input to `jq`.

Different image registries have different requirements for authentication and
authorization.

ghcr.io is particlarly puzzling to me. _sometimes_ the following bashery works,
evem though the ghcr doc says to use a GitHub PAT,
and sometimes it does not. It seems to be a trimmed version of what the
www-authenticate header in the error response suggests.
Code has been revised to try what is suggested automatically.

TOKEN="$(
  curl "https://ghcr.io/token?scope=repository:${ORG}/${PKG}:pull" |
  awk -F'"' '$0=$4'
)"
"""

link_pat = '<([^>]*)> *;( *([^ =]*) *= *(.*) *)'
link_re = re.compile(link_pat)

def rfc5988_parse_next(link: str, url):
    match = link_re.fullmatch(link)
    if not match:
        print(f'GET {url} returned malformed Link header value {link}!', file=sys.stderr)
        return None
    groups = match.groups()
    next_ref = groups[0]
    for idx in range(2, len(groups), 3):
        parm_name = groups[idx]
        parm_value = groups[idx+1]
        if parm_name != 'rel':
            continue
        if parm_value != 'next' and parm_value != '"next"':
            continue
        next_url_str = url.scheme + '://' + url.netloc + next_ref
        next_url = urllib.parse.urlparse(next_url_str)
        return next_url
    return None

def request_bearer_token(authclue: str) -> typing.Union[str,None]:
    if not authclue.startswith('Bearer '):
        return None
    authclue = authclue.removeprefix('Bearer ')
    parms_list = authclue.split(',')
    parms_dict = dict()
    for parm in parms_list:
        parts = parm.partition('=')
        parms_dict[parts[0]] = parts[2].strip('"')
    if 'realm' not in parms_dict:
        return None
    req_url_str = parms_dict['realm']
    sep = '?'
    for key, val in parms_dict.items():
        if key == 'realm':
            continue
        req_url_str += sep + key + '=' + val
        sep = '&'
    try:
        with urllib.request.urlopen(req_url_str) as resp:
            if resp.status == 200:
                body_bytes = resp.read()
                body_str = body_bytes.decode('utf-8')
                body_parsed = json.loads(body_str)
                if 'token' in body_parsed:
                    token = body_parsed['token']
                    print(f'Subsidiary request to {req_url_str} returned bearer token {token}', file=sys.stderr)
                    return token
                print(f'Subsidiary request to {req_url_str} returned junk: {body_str}', file=sys.stderr)
                return None
            print(f'Subsidiary request to {req_url_str} returned status {resp.status}!', file=sys.stderr)
            return None
    except Exception as err:
        print(f'Subsidiary request {req_url_str} raised {type(err)} {err}!', file=sys.stderr)
        return None

def read_with_next(url, headers:dict[str,str], accept:typing.Optional[str]=None) -> list:
    url_str = urllib.parse.urlunparse(url)
    for retry in ('Authorization' not in headers, False):
        all_headers = dict(**headers)
        if accept:
            all_headers['Accept'] = accept
        req = urllib.request.Request(url_str, headers=all_headers)
        try:
            with urllib.request.urlopen(req) as resp:
                if resp.status != 200:
                    print(f'GET {url_str} returned status {resp.status}!', file=sys.stderr)
                    sys.exit(10)
                    return
                body_bytes = resp.read()
                body = body_bytes.decode('utf-8')
                ans = [(resp.headers, body)]
                break
        except urllib.error.HTTPError as err:
            print(f'GET {url_str} with headers {all_headers} raised HTTPError: code={err.code}, reason={err.reason}',
                file=sys.stderr)
            if retry and err.code == 401 and 'www-authenticate' in err.headers:
                token = request_bearer_token(err.headers['www-authenticate'])
                if token is not None:
                    headers['Authorization'] = 'Bearer ' + token
                    continue
            return []
        except Exception as err:
            print(f'GET {url_str} raised {type(err)} {err}!', file=sys.stderr)
            return []
    link = ans[0][0]['Link']
    if not link:
        return ans
    next_url = rfc5988_parse_next(link, url)
    if not next_url:
        return ans
    rest = read_with_next(next_url, headers)
    return ans + rest

def main() -> None:
    parser = argparse.ArgumentParser(description='Inspect image reference to produce a series of JSON objects.')
    parser.add_argument('ref', help='the image reference')
    parser.add_argument('--bearer-token', help='for Authorization: header')
    args = parser.parse_args()
    headers = dict[str,str]()
    if args.bearer_token:
        headers['Authorization'] = 'Bearer ' + args.bearer_token
    ref = args.ref
    url = urllib.parse.urlparse('https://' + ref)
    image_path = url.path
    url = url._replace(path = '/v2' + image_path)
    tags_url = url._replace(path = url.path + '/tags/list')
    tags_replies = read_with_next(tags_url, headers)
    tags = []
    for tags_reply in tags_replies:
        tags_parsed = json.loads(tags_reply[1])
        tags = tags + tags_parsed['tags']
    for tag in tags:
        manifest_url = url._replace(path = url.path + '/manifests/' + tag)
        manifest_replies = read_with_next(manifest_url, headers, 'application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.v2+json,application/vnd.docker.distribution.manifest.v1+json')
        if not manifest_replies:
            continue
        reply_headers, body = manifest_replies[0]
        manifest_parsed = json.loads(body)
        cd = reply_headers.get('Docker-Content-Digest')
        if cd:
            manifest_parsed['DockerDigest'] = cd
        manifest_parsed['Reference'] = tag
        print(json.dumps(manifest_parsed))
    return

if __name__ == '__main__':
    main()
