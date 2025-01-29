The following steps show how to create two new `kind` clusters and
register them with the hub as described in the
[official open cluster management docs](https://open-cluster-management.io/docs/getting-started/installation/start-the-control-plane/).

Note that `kind` does not support three or more concurrent clusters unless you raise some limits as described in this `kind` "known issue": [Pod errors due to “too many open files”](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

1. Execute the following commands to create two kind clusters, named `cluster1` and `cluster2`, and register them with the OCM hub. These clusters will serve as workload clusters. If you have previously executed these commands, you might already have contexts named `cluster1` and `cluster2`. If so, you can remove these contexts using the commands `kubectl config delete-context cluster1` and `kubectl config delete-context cluster2`.

    ```shell
    : set flags to "" if you have installed KubeStellar on an OpenShift cluster
    flags="--force-internal-endpoint-lookup"
    clusters=(cluster1 cluster2);
    for cluster in "${clusters[@]}"; do
       kind create cluster --name ${cluster}
       kubectl config rename-context kind-${cluster} ${cluster}
       clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --singleton '${flags}'"}' | sh
    done
    ```

    The `clusteradm` command grabs a token from the hub (`its1` context), and constructs the command to apply the new cluster
    to be registered as a managed cluster on the OCM hub.

2. Repeatedly issue the command:

    ```shell
    kubectl --context its1 get csr
    ```

    until you see that the certificate signing requests (CSR) for both cluster1 and cluster2 exist.
    Note that the CSRs condition is supposed to be `Pending` until you approve them in step 4.

3. Once the CSRs are created, approve the CSRs complete the cluster registration with the command:

    ```shell
    clusteradm --context its1 accept --clusters cluster1
    clusteradm --context its1 accept --clusters cluster2
    ```

4. Check the new clusters are in the OCM inventory and label them:

    ```shell
    kubectl --context its1 get managedclusters
    kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1
    kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2
    ```
