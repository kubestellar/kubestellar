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

#### Deployment
Starting from a local directory containing the git repo, do the following:

1. Deploy the KS control plane hosting infra: 

    First, set the following variables:
    ```console
    VPC="<vpc_name>"
    EC2_AMI="<aws-ami>"
    KS_RELEASE="<ks-release>"
    EC2_SSH_PUBLIC_KEY="<mykey>"
    ARCH="x86_64"
    REGION="us-east-2"
    CLUSTER_NAME="core"
    NUM_WORKER_NODES=2
    EC2_INSTANCE_TYPE="t2.xlarge"
    ```
     
    Set var `KS_RELEASE` to release 0.25.1 or later. Then, deploy the core infrastructure which includes a VPC, security groups, EC2 instances, etc.
    ```bash
    cd test/scale-infra
    ./deploy_ks_cp_infra.sh --cluster-name $CLUSTER_NAME --region $REGION --vpc-name $VPC --k8s-num-workers $NUM_WORKER_NODES --instances-type $EC2_INSTANCE_TYPE --aws-key-name $EC2_SSH_PUBLIC_KEY --arch $ARCH --ec2-image-id $EC2_AMI --ks-release $KS_RELEASE
    ```

    The above command creates the required AWS infrastructure including a VPC, security groups and EC2 instances. Then, it creates a Kubernetes cluster deployed using Kubeadm. Lastly, it deploys the KubeStellar core components. You can use the flag `--ks-release` to specify the KubeStellar release. Kubestellar is deployed using the [KS helmchart](https://github.com/kubestellar/kubestellar/tree/release-0.26.0/core-chart) configured with a ITS of type host. 

    Use the flag `--vpc-name` to specify the name for the [AWS virtual private cloud](https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html) to deploy your infrastructure in a logically isolated virtual network: *We highly advise utilizing a unique name or the AWS IAM user ID as the identifier for your VPC*. Furthermore, use the flag `--ec2-image-id` to specify the [Amazon machine image ID](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html), keeping in mind that it is region-specific.

    We advise to use the Ubuntu Amazon EC2 AMI Locator to acquire the AMI value for a specific region: https://cloud-images.ubuntu.com/locator/ec2/. Select Ubuntu release `22.04` or `24.04`. 

    Also, use the flag `--aws-key-name` to specify the name of the uploaded public ssh key to your AWS account in the target region. 

    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the nodes that constitute the Kubernetes cluster will be generated at the current directory at `.data/${REGION}_${VPC}/hosts_core`.

    Furthermore, use the generated kubeconfig file `.data/${REGION}_${VPC}/admin.conf` to access the Kubernetes cluster and the control plane components for KubeStellar, for example:
    
    Check the created kubeconfig contexts:

    ```bash 
    kubectl --kubeconfig=$PWD/.data/${REGION}_${VPC}/admin.conf config get-contexts
    ```

    Output:
    ```console
    CURRENT   NAME     CLUSTER        AUTHINFO           NAMESPACE
    *         kscore   kubernetes     kubernetes-admin   
              wds1     wds1-cluster   wds1-admin         default
    ```


    To access the `wds1` k8s context first extract the command displayed at the end of the last Ansible task to edit your `etc/hosts` file with wds1 host entry:

    (*Sample Ansible task command line output*):

    ```console
    TASK [Command to edit your etc/hosts file with wds1 host entry] ***********************************************************************
    ok: [master1] => {
        "msg": "echo '192.168.56.20 wds1.ec2-192-168-56-20.compute-1.amazonaws.com' | sudo tee -a /etc/hosts"
    }

    PLAY RECAP ****************************************************************************************************************************
    master1                    : ok=24   changed=20   unreachable=0    failed=0    skipped=2    rescued=0    ignored=1   
    ```

    Then, extract and execute the command displayed in the Ansible task above:

    ```bash
    echo '192.168.56.20 wds1.ec2-192-168-56-20.compute-1.amazonaws.com' | sudo tee -a /etc/
    ```

    Lastly, check the namespaces created in the `wds1` context:

    ```bash 
    kubectl --kubeconfig=$PWD/.data/${REGION}_${VPC}/admin.conf --context wds1 get ns
    ```

    Output:
    ```console
    NAME                 STATUS   AGE
    kube-system          Active   29m
    kube-public          Active   29m
    kube-node-lease      Active   29m
    default              Active   29m
    kubestellar-report   Active   28m
    ```


2. Create WEC hosting instances:
    
    Set the following variables:
    ```console
    CLUSTER_NAME="wec"
    NUM_HOSTING_INSTANCES=2
    EC2_INSTANCE_TYPE="t2.xlarge"
    ```

    Then, deploy the WEC infra:

    ```bash
    ./deploy_wec_infra.sh --cluster-name $CLUSTER_NAME --region $REGION --vpc-name $VPC --wecs-hosting-instances $NUM_HOSTING_INSTANCES --instances-type $EC2_INSTANCE_TYPE --aws-key-name $EC2_SSH_PUBLIC_KEY --arch $ARCH --ec2-image-id $EC2_AMI
    ```

    Use the flag `--wecs-hosting-instances` to specify the number of ec2 instances to be created to host the WECs. You must create the WEC hosting instances in the same region as the KS control plane hosting infra created a step 1 - multiple regions deployment is not supported at the moment.  

    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the ec2 WEC hosting instances will be generated in the present directory at `.data/${REGION}_${VPC}/hosts_wec`.

3. Create WEC kind clusters:

    a) Update the inventory of the WEC instances to include the IP address of a master node from the K8s cluster created in step 1 above:

    First, find the required master's IP address info in the following file: `.data/${REGION}_${VPC}/hosts_core`.

    Next, use the following command to see the content of the WEC ansible inventory:
    ```bash
    cat .data/${REGION}_${VPC}/hosts_wec
    ```

    Sample output: 
    ```console
    [masters]

    [add_workers]
    worker1 ansible_host=192.168.56.1
    worker1 ansible_host=192.168.56.2
    ```

    Lastly, edit the contents of the file `.data/${REGION}_${VPC}/hosts_wec` to include the IP address and name of a master node in the [masters] section, for example:
    ```console
    [masters]
    master1 ansible_host=192.168.56.20

    [add_workers]
    worker1 ansible_host=192.168.56.1
    worker1 ansible_host=192.168.56.2
    ```

    b) Create WEC kind clusters and connect to KS Core cluster:

    ```bash
    ansible-playbook -i .data/${REGION}_${VPC}/hosts_wec deploy_ks_wec.yaml --ssh-common-args="-o StrictHostKeyChecking=no" -e "region=$REGION vpc_name=$VPC num_wecs=1"
    ```

    Use the input paramater `num_wecs` to specify the number of kind clusters to be created for each WEC Hosting Instances. The above command creates kind WEC clusters and connects them to the KubeStellar core cluster created in step 1. Furthermore, it attaches a [KWOK](https://github.com/kubernetes-sigs/kwok) fake node to each kind cluster. 

    c) Access the WEC kind clusters from Ansible control machine:

    A kubeconfig file is generated for each WEC hosting instances at the following directory `.data/${REGION}_${VPC}/wecs_config/`. For example:

    ```bash
    ls .data/${REGION}_${VPC}/wecs_config/
    ``` 

    Output:
    ```console
    192-168-56-1.conf	192-168-56-2.conf
    ```
    
    Use these kubeconfig to access the WEC kind clusters, example:

    ```bash
    kubectl --kubeconfig=$PWD/.data/${REGION}_${VPC}/wecs_config/192-168-56-1.conf config get-contexts --insecure-skip-tls-verify=true
    ```

    Output:
    ```console
    CURRENT   NAME              CLUSTER                AUTHINFO               NAMESPACE
    *         192-168-56-1-0   kind-192-168-56-1-0   kind-192-168-56-1-0 
    ```

    Lastly, check that the kind WEC clusters are connected to the KubeStellar control plane:

    ```bash
    kubectl --kubeconfig=$PWD/.data/${REGION}_${VPC}/admin.conf get managedclusters
    ```

    Output:
    ```console
    NAME             HUB ACCEPTED   MANAGED CLUSTER URLS                        JOINED   AVAILABLE   AGE
    192-168-56-1-0    true        https://192-168-56-1-0-control-plane:6443       True     True        5m2s
    192-168-56-2-0    true        https://192-168-56-2-0-control-plane:6443       True     True        5m2s
    ```


#### Uninstall

To destroy the infrastructure run the following command:

```bash
./delete_all_infra.sh  --region $REGION --vpc-name $VPC
```

As an alternative to this quick start, a step-by-step bootstrapping of all the components can be done by following the instructions [here](INSTRUCTIONS.md).