import re
import sys

# Usage: python extract_min_kflex_version.py [path_to_check_pre_req.sh]
if len(sys.argv) > 1:
    path = sys.argv[1]
else:
    path = 'scripts/check_pre_req.sh'

with open(path) as f:
    content = f.read()

# Look for the is_installed_kflex() function and extract the min version
match = re.search(r"is_installed_kflex\(\) \{.*?['\"]Kubeflex version: v([0-9.]+)['\"]", content, re.DOTALL)
if match:
    print(match.group(1))
else:
    print("ERROR: Could not find min kflex CLI version in {}".format(path), file=sys.stderr)
    sys.exit(1)
