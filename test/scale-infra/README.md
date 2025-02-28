A scalable Kubernetes-based testbed for KubeStellar Performance Tests
---------------------------------------------------------------------

<img src="images/ks-scale-test-infra.drawio.svg" width="60%" height="60%" title="ks-scale-test-infra">


### Prerequisites

In order to follow the instructions below, you must have [python3](https://www.python.org/downloads/) with all the dependencies listed [here](common/requirements.txt) installed. We recommend to create a python virtual environment `.venv` under `kubestellar/test/scale-infra`, for example: 

```bash
cd kubestellar/test/scale-infra 
python3 -m venv .venv
. .venv/bin/activate
pip3 install --upgrade pip
pip3 install -r requirements.txt
```

Additionally, you must have the following:

- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- AWS access key ID and AWS secret access key (to configure AWS CLI follow instructions [here](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-files.html#cli-configure-files-methods))
- An SSH "keypair" registered with EC2 in the region that you will be using. 


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
    ./deploy_ks_cp_infra.sh --region us-east-2 --vpc-name <vpc_name> --k8s-num-masters 1 --k8s-num-workers 2 --instances-type t2.xlarge --aws-key-name mykey  --arch x86_64 --ec2-image-id <aws-ami> --ks-release 0.26.0
    ```

    The above command creates the required AWS infrastructure including a VPC, security groups and EC2 instances. Then, it creates a Kubernetes cluster deployed using Kubeadm. Lastly, it deploys the KubeStellar core components. You can use the flag `--ks-release` to specify the KubeStellar release. Kubestellar is deployed using the [KS helmchart](https://github.com/kubestellar/kubestellar/tree/release-0.26.0/core-chart) configured with a ITS of type host. 

    Use the flag `--vpc-name` to specify the name for the [AWS virtual private cloud](https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html) to deploy your infrastructure in a logically isolated virtual network: *We highly advise utilizing a unique name or the AWS IAM user ID as the identifier for your VPC*. Furthermore, use the flag `--aws-ami` to specify the [Amazon machine image ID](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html), keeping in mind that it is region-specific.

    We advise utilizing the subsequent command to acquire the AMI ("ImageId") value for a specific region:

    ```bash
    aws ec2 describe-images --region <region_name> --filters "Name=architecture,Values=x86_64" "Name=name,Values=ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-*"
    ```

    Also, use the flag `--aws-key-name` to specify the name of the uploaded public ssh key to your AWS account in the target region. 

    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the nodes that constitute the Kubernetes cluster will be generated at the current directory at `.data/hosts_core`.


2. Create WEC hosting instances:

    ```
    ./deploy_wec_infra.sh --region us-east-2 --vpc-name <vpc_name> --wecs-hosting-instances 1 --instances-type t2.2xlarge --aws-key-name  mykey  --arch x86_64 --ec2-image-id <aws-ami>
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
    ./delete_all_infra.sh  --region us-east-2 --vpc-name <vpc_name>
    ```

    As an alternative to this quick start, a step-by-step bootstrapping of all the components can be done by following the instructions [here](INSTRUCTIONS.md).