# Create and Register WECs for examples

The following steps show how to create new clusters and
register them with the hub as descibed in the
[official open cluster management docs](https://open-cluster-management.io/getting-started/installation/start-the-control-plane/).

1. Run the following set of commands for creating two kind clusters with name `cluster1` and `cluster2` and registering it with the
OCM hub. This cluster will act as a workload cluster.

   ```shell
   flags="--force-internal-endpoint-lookup" # set this to "" if you have installed KubeStellar on an OpenShift cluster
   clusters=(cluster1 cluster2);
   for cluster in "${clusters[@]}"; do
      kind create cluster --name ${cluster}
      kubectl config rename-context kind-${cluster} ${cluster}
      clusteradm --context imbs1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' '${flags}'"}' | sh
   done   
   ```

   The last line grabs a token from the hub (`imbs1` context), and constructs the command to apply on each cluster
   to be registered as a managed cluster on the OCM hub.

2. Issue the command:

   ```shell
   watch kubectl --context imbs1 get csr
   ```

   and wait until the certificate signing requests (CSR) for both cluster1 and cluster2 are created, then
   ctrl+C.
   Note that the CSRs condition is supposed to be `Pending` until you approve them in step 4.

3. Once the CSRs are created approve the csrs to complete the cluster registration with the command:

   ```shell
   clusteradm --context imbs1 accept --clusters cluster1
   clusteradm --context imbs1 accept --clusters cluster2
   ```

4. Check the new clusters are in the OCM inventory and label them:

   ```shell
   kubectl --context imbs1 get managedclusters
   kubectl --context imbs1 label managedcluster cluster1 location-group=edge name=cluster1
   kubectl --context imbs1 label managedcluster cluster2 location-group=edge name=cluster2
   ```
