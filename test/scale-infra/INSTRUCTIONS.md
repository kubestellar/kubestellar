### Examples of Individual Steps
As an alternative to the quick-start deployment bootstrapping instructions, you can also run the individual steps starting from a local directory containing the git repo as follows:


## Deployment
1. Create the KubeStellar core infra on AWS:

    First, set the following variables:

    ```console
    VPC="<vpc_name>"
    EC2_AMI="<aws-ami>"
    EC2_SSH_PUBLIC_KEY="<mykey>"
    REGION="us-east-2"
    ARCH="x86_64"
    CLUSTER_NAME="core"
    NUM_WORKER_NODES=2
    EC2_INSTANCE_TYPE="t2.xlarge"
    ```

    Then, deploy the core infrastructure which includes a VPC, security groups, EC2 instances, etc.

    ```bash
    cd test/scale-infra
    ansible-playbook deploy_vpc_core.yaml -e "region=$REGION vpc_name=$VPC"
    ansible-playbook create-ec2.yaml -e "cluster_name=$CLUSTER_NAME region=$REGION vpc_name=$VPC aws_key_name=$EC2_SSH_PUBLIC_KEY num_workers=$NUM_WORKER_NODES instance_type=$EC2_INSTANCE_TYPE arch=$ARCH ec2_image=$EC2_AMI"
    ```

    Use the variable `VPC` to specify the name for the [AWS virtual private cloud](https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html) to deploy your infrastructure in a logically isolated virtual network: *We highly advise utilizing a unique name or the AWS IAM user ID as the identifier for your VPC*. Furthermore, use the variable `EC2_AMI` to specify the [Amazon machine image ID](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html), keeping in mind that it is region-specific.

    We advise to use the Ubuntu Amazon EC2 AMI Locator to acquire the AMI value for a specific region: https://cloud-images.ubuntu.com/locator/ec2/. Select Ubuntu release `22.04` or `24.04`.

    Also, use the var `EC2_SSH_PUBLIC_KEY` to specify the name of the uploaded public ssh key to your AWS account in the target region.

    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the master and worker nodes will be generated in the present directory at `.data/${REGION}_${VPC}/hosts_core`.

2. Deploy a Kubernetes cluster:

    ```bash
    ansible-playbook -i .data/${REGION}_${VPC}/hosts_core deploy-masters.yaml --ssh-common-args='-o StrictHostKeyChecking=no'
    ansible-playbook -i .data/${REGION}_${VPC}/hosts_core deploy-workers.yaml --ssh-common-args='-o StrictHostKeyChecking=no'
    ```

3. Deploy KubeStellar in the hosting cluster:

    Set var `KS_RELEASE` to release 0.25.1 or later. Then deploy KubeStellar: 
    ```bash
    ansible-playbook -i .data/${REGION}_${VPC}/hosts_core deploy_ks_core.yaml --ssh-common-args="-o StrictHostKeyChecking=no" -e "region=$REGION vpc_name=$VPC cluster_name=$CLUSTER_NAME ks_release=$KS_RELEASE"
    ```

    You can use the variable `ks_release` to specify the KubeStellar release. Kubestellar is deployed using the [KS helmchart](https://github.com/kubestellar/kubestellar/tree/release-0.26.0/core-chart) configured with a ITS of type host. 

    Upon completion of the script's execution, a kubeconfig file is generated at `.data/${REGION}_${VPC}/admin.conf` to access the Kubernetes cluster and the control plane components for KubeStellar, for example:
    
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

4. Create the WEC hosting instances:

    Set the following variables:
    ```console
    CLUSTER_NAME="wec"
    NUM_HOSTING_INSTANCES=2
    EC2_INSTANCE_TYPE="t2.xlarge"
    ```

    Then, deploy the WEC infra:
    ```bash
    ansible-playbook create-ec2.yaml -e "cluster_name=$CLUSTER_NAME region=$REGION vpc_name=$VPC aws_key_name=$EC2_SSH_PUBLIC_KEY wecs_hosting_instances=$NUM_HOSTING_INSTANCES instance_type=$EC2_INSTANCE_TYPE archt=$ARCH ec2_image=$EC2_AMI" 
    ```

    Use the variable `NUM_HOSTING_INSTANCES` to specify the number of ec2 instances to be created to host WEC kind clusters.
    
    Upon completion of the script's execution, an Ansible inventory file containing the IP addresses of the ec2 WEC hosting instances will be generated in the present directory at `.data/${REGION}_${VPC}/hosts_wec`.

5. Create the WEC kind clusters:

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

    b) Create WEC kind clusters and connect to KS Core cluster

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


## Uninstall

To destroy the infrastructure proceed as following:
   
1. Set the following variables to specify the region and vpc name:

    ```console
    REGION="us-east-2"
    VPC="<vpc_name>"
    ```

2. Delete the deployed infra:

    a) Delete WECs infra:
    ```bash
    ansible-playbook -i .data/${REGION}_${VPC}/hosts_wec delete-ec2.yaml -e "cluster_name=wec region=$REGION group=$VPC"
    ```

    b) Delete KubeStellar core infra: 

    ```bash
    ansible-playbook -i .data/${REGION}_${VPC}/hosts_core delete-ec2.yaml -e "cluster_name=core region=$REGION group=$VPC"
    ```

    c) Delete VPC:

    ```bash
    ansible-playbook delete_vpc_infra.yaml -e "region=$REGION  vpc_name=$VPC"
    ```
