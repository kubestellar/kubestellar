A scalable Kubernetes-based testbed for KubeStellar Performance Tests
---------------------------------------------------------------------

<img src="images/ks-scale-test-infra.drawio.svg" width="60%" height="60%" title="ks-scale-test-infra">

### Prerequisites
- Ansible, with 
  - the `amazon.aws` collection installed (see, e.g., https://docs.ansible.com/ansible/latest/collections/amazon/aws/ec2_instance_module.html), and
  - [PyPI package boto3](https://pypi.org/project/boto/) installed for the Python interpreter that Ansible uses on your machine;
- AWS CLI;
- AWS access key ID and AWS secret access key (to configure AWS CLI [see](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-files.html#cli-configure-files-methods))
- An SSH "keypair" registered with EC2

### Overview

The testbed consists of (from bottom to top):
- Networking and storage resources from AWS;
- AWS EC2 instances;
- Kubernetes cluster(s);
- Workload(s): Kubestellar Core and WEC components 


### Quick Start Deployment Instructions:

Starting from a local directory containing the git repo, do the following:

1. Deploy the KS control plane hosting infra: 

    ```
    cd test/scale-infra
    ./deploy_ks_cp_infra.sh --region us-east-2 --vpc-name ks-core --k8s-num-masters 1 --k8s-num-workers 2 --instances_type t2.xlarge --aws-key-name mykey  --arch x86_64 --ec2-image-id <aws-ami> --ks-release 0.25.1
    ```

    The above command creates the required AWS infrastructure including a VPC, security groups and EC2 instances. Then, it creates a Kubernetes cluster deployed using Kubeadm. Lastly, it deploys the KubeStellar core components. You can use the flag `--ks-release` to specify the KubeStellar release. Kubestellar is deployed using the [KS helmchart](https://github.com/kubestellar/kubestellar/tree/main/core-chart) configured with a ITS of type host. 

    You can also use the flag `--vpc-name` to specify the name for the [AWS virtual private cloud](https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html) to deploy your infrastructure in a logically isolated virtual network: *we strongly recommend to use your name or AWS IAM user ID*.

    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the nodes that constitute the Kubernetes cluster will be generated at the current directory at `.data/hosts_core`.


2. Create WEC hosting instances:

    ```
    ./deploy_wec_infra.sh --region us-east-2 --vpc_name ks-core --wecs_hosting_instances 1 --instance_type t2.2xlarge --aws_key_name  mykey  --arch x86_64 --ec2_image_id <aws-ami>
    ```

    Use the flag `--wecs-hosting-instances` to specify the number of ec2 instances to be created to host the WECs. You must create the WEC hosting instances in the same region as the KS control plane hosting infra created a step 1 - multiple regions deployment is not supported at the moment.  

    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the ec2 WEC hosting instances will be generated in the present directory at `.data/hosts_wec`.

3. Create the WEC kind clusters:

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


4. Destroy the infrastructure.

    ```
    ./delete_all_infra.sh  --region us-east-2 --vpc-name <name>
    ```

    As an alternative to this quick start, a step-by-step bootstrapping of all the components can be done by following the instructions [here](INSTRUCTIONS.md).