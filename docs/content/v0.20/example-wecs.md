# Create and Register WECs for examples

The following steps show how to create new clusters and
register them with the hub as descibed in the
[official open cluster management docs](https://open-cluster-management.io/getting-started/installation/start-the-control-plane/).

1. Run the following set of commands for creating a new kind cluster with name "cluster1" and registering it with the
OCM hub. This cluster will act as a workload cluster.

   ```shell
   export CLUSTER=cluster1
   kind create cluster --name ${CLUSTER}
   kubectl config rename-context kind-${CLUSTER} ${CLUSTER}
   clusteradm --context transport1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${CLUSTER}/" | awk '{print $0 " --context '${CLUSTER}' --force-internal-endpoint-lookup"}' | sh
   ```

   The last line grabs a token from the hub (`transport1` context), and constructs the command to apply on the new cluster
   to be registered as a managed cluster on the OCM hub.

2. Repeat for a second workload cluster:

   ```shell
   export CLUSTER=cluster2
   kind create cluster --name ${CLUSTER}
   kubectl config rename-context kind-${CLUSTER} ${CLUSTER}
   clusteradm --context transport1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${CLUSTER}/" | awk '{print $0 " --context '${CLUSTER}' --force-internal-endpoint-lookup"}' | sh
   ```

3. Issue the command:

   ```shell
   watch kubectl --context transport1 get csr
   ```

   and wait until the certificate signing requests (CSR) for both cluster1 and cluster2 are created, then
   ctrl+C.
   Note that the CSRs condition is supposed to be `Pending` until you approve them in step 4.

4. Once the CSRs are created approve the csrs to complete the cluster registration with the command:

   ```shell
   clusteradm --context transport1 accept --clusters cluster1
   clusteradm --context transport1 accept --clusters cluster2
   ```

5. Check the new clusters are in the OCM inventory and label them:

   ```shell
   kubectl --context transport1 get managedclusters
   kubectl --context transport1 label managedcluster cluster1 location-group=edge name=cluster1
   kubectl --context transport1 label managedcluster cluster2 location-group=edge name=cluster2
   ```
