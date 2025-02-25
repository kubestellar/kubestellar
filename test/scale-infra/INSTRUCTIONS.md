### Examples of Individual Steps
As an alternative to the quick-start deployment bootstrapping instructions, you can also run the individual steps starting from a local directory containing the git repo as follows:

1. Create the KubeStellar core infra on AWS. The infrastructure includes a VPC, security groups, EC2 instances, etc.

    ```bash
    cd test/scale-infra
    ansible-playbook deploy_vpc_core.yaml -e "region=us-east-2 vpc_name=<vpc_name>"
    ansible-playbook create-ec2.yaml -e "cluster_name=core region=us-east-2 vpc_name=<vpc_name> aws_key_name=mykey  num_masters=1 num_workers=2 instance_type=t2.xlarge arch=x86_64 ec2_image='<your-aws-ami>'"
    ```

    Use the variable `vpc_name` to specify the name for the [AWS virtual private cloud](https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html) to deploy your infrastructure in a logically isolated virtual network: *We highly advise utilizing a unique name or the AWS IAM user ID as the identifier for your VPC*. Furthermore, use the flag `--aws-ami` to specify the [Amazon machine image ID](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html), keeping in mind that it is region-specific.

    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the master and worker nodes will be generated in the present directory at `.data/hosts_core`.

2. Deploy a Kubernetes cluster:

    ```bash
    ansible-playbook -i .data/hosts_core deploy-masters.yaml --ssh-common-args='-o StrictHostKeyChecking=no'
    ansible-playbook -i .data/hosts_core deploy-workers.yaml --ssh-common-args='-o StrictHostKeyChecking=no'
    ```

3. Deploy KubeStellar in the hosting cluster:

    ```bash
    ansible-playbook -i .data/hosts_core deploy_ks_core.yaml --ssh-common-args='-o StrictHostKeyChecking=no' -e 'ks_release=0.26.0'
    ```

    You can use the variable `ks_release` to specify the KubeStellar release. Kubestellar is deployed using the [KS helmchart](https://github.com/kubestellar/kubestellar/tree/release-0.26.0/core-chart) configured with a ITS of type host. 

4. Create the WEC hosting instances:

    ```bash
    ansible-playbook create-ec2.yaml -e "cluster_name=wec region=us-east-2 vpc_name=<vpc_name> aws_key_name=mykey wecs_hosting_instances=1 instance_type=t2.xlarge archt=x86_64 ec2_image='<your-aws-ami>" 
    ```

    Use the variable `wecs_hosting_instances` to specify the number of ec2 instances to be created to host WEC kind clusters.
    
    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the ec2 WEC hosting instances will be generated in the present directory at `.data/hosts_wec`.

5. Create the WEC kind clusters:

    a) Update the inventory of the WEC instances to include the IP address of a master node from the K8s cluster created in step 1 above:

    First, find the required master's IP address info in the following file: `.data/hosts_core`.

    Next, use the following command to see the content of the WEC ansible inventory:
    ```bash
    cat .data/hosts_wec
    ```

    Sample output: 
    ```console
    [masters]
    <add master node info here!>

    [add_workers]
    worker1 ansible_host=192.168.56.1
    ```

    Lastly, edit the contents of the file `.data/hosts_wec` to include the IP address of a master node:
    ```bash
    vi .data/hosts_wec
    ```

    b) Create Kind cluster WECs and connect to KS Core cluster

    ```
    ansible-playbook -i .data/hosts_wec deploy_ks_wec.yaml --ssh-common-args='-o StrictHostKeyChecking=no' -e 'num_wecs=1'
    ```

    Use the input paramater `num_wecs` to specify the number of kind clusters to be created for each WEC Hosting Instances. The above command creates kind WEC clusters and connects them to the KubeStellar core cluster created in step 1. Furthermore, it attaches a [KWOK](https://github.com/kubernetes-sigs/kwok) fake node to each kind cluster.

6. (Optional) Delete worker nodes from the cluster.
    Edit `.data/hosts_core` by adding the corresponding entries in the `[remove_workers]` Ansible inventory group.
    Edit `delete-worker.yaml` by changing the `node_name`, which is shown by `kubectl get nodes`.
    Run the playbook:

    ```bash
    ansible-playbook -i .data/hosts_core delete-worker.yaml
    ```

7. Destroy the infrastructure.

    a) Delete WECs infra:
    ```bash
    ansible-playbook -i .data/hosts_wec delete-ec2.yaml -e "cluster_name=wec region=us-east-2 group=<aws-iam-user>"
    ```

    b) Delete KubeStellar core infra: 

    ```bash
    ansible-playbook -i .data/hosts_core delete-ec2.yaml -e "cluster_name=core region=us-east-2 group=<aws-iam-user>"
    ```

    c) Delete VPC:

    ```bash
    ansible-playbook delete_vpc_infra.yaml -e "region=us-east-2  name=<aws-iam-user>"
    ```
