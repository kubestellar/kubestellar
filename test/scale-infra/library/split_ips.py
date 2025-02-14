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

from ansible.module_utils.basic import *

def main():

    fields = {
        "cluster_name": {"required": True, "type": "str"},
        "public_ips": {"required": True, "type": "list"},
        "position": {"required": True, "type": "int"},
    }

    module = AnsibleModule(argument_spec=fields)

    cluster_name = module.params["cluster_name"]
    public_ips = module.params["public_ips"]
    position = module.params["position"]
    master_public_ips = public_ips[:position]
    worker_public_ips = public_ips[position:]


    fout = open('_'.join(['.data/hosts', cluster_name]), 'w') # output ansible hosts file with ec2 public ip addresses
    fout.write('[masters]' + "\n")

    i=1

    for ip in master_public_ips:
       fout.write('master' + str(i) + ' ' + 'ansible_host=' + ip + '\n')
       i +=1

    fout.write("\n")
    fout.write('[add_workers]' + "\n")

    j=1

    for ip in worker_public_ips:
       fout.write('worker' + str(j) + ' ' + 'ansible_host=' + ip + '\n')
       j+=1

    fout.write('\n')
    fout.write('[remove_workers]' + '\n')
    fout.flush()
    fout.close()

    module.exit_json(changed=True,
                     master_public_ips=master_public_ips,
                     worker_public_ips=worker_public_ips)

if __name__ == '__main__':
    main()
