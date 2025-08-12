import re
import sys
import argparse
import os

parser = argparse.ArgumentParser(description='Extract minimum kflex CLI version from a script.')
parser.add_argument('path', nargs='?', default='scripts/check_pre_req.sh', help='Path to check_pre_req.sh script')
args = parser.parse_args()

path = args.path

if not os.path.isfile(path):
    print(f"ERROR: File not found: {path}", file=sys.stderr)
    sys.exit(1)

with open(path) as f:
    content = f.read()

match = re.search(r"is_installed_kflex\(\) \{.*?['\"]Kubeflex version: v([0-9]+\.[0-9]+\.[0-9]+(?:[-a-zA-Z0-9\.]*)?)['\"]", content, re.DOTALL)
if match:
    print(match.group(1))
else:
    print(f"ERROR: Could not find min kflex CLI version in {path}", file=sys.stderr)
    sys.exit(1)
