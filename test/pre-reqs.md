The following pre-requisites are required:
- kubectl 

- kubeflex version 0.3 or higher
    To install kubeflex go to [https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation)

- OCM cli (clusteradm)
    To install OCM cli use:
    ```
    curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh | bash
    ```
    Note that the default installation of clusteradm will install in /usr/local/bin which will require root access. If you prefer to avoid root, you can specify an alternative installation path using the INSTALL_DIR environment variable, as follows:
    ```
    mkdir -p ocm
    export INSTALL_DIR="$PWD/ocm"
    curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh | bash
    export PATH=%PWD/ocm:$PATH
    ```

- helm - to deploy the kubestellar and kubeflex charts
- kind - to create a few small kubernetes clusters


