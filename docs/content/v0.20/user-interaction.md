# User Interaction

This section describes the user interaction to bring up
a KubeStellar instance and to deploy workloads to a fleet
of Workload Execution Clusters (WECs).

## Setup 

From a user perspective, the steps to get a KubeStellar instance up and
running using the *kflex* CLI are the following:

1.  **“*kflex init*”** - installs KubeFlex on a hosting cluster. This
    step installs the KubeFlex ControlPlane CRD and controller manager,
    and sets up a shared Postgres DB, all in the *kubeflex-system*
    namespace.

2.  “***kflex create imbs1 --type vcluster -p ocm***” - creates the
    **imbs1** Inventory & Transport Space of type *vcluster* running the OCM
    (Open Cluster Management) Cluster Manager. Note that *-p ocm* runs a *post-create
    hook* on the vcluster control plane which installs OCM on it.

3.  “***kflex create wds1 -p kubestellar***” - creates the Workload
    Description Space **wds1**. Similarly to before, *-p kubestellar* runs a
    post-create hook on the k8s control plane that starts an instance of
    a KubeStellar controller manager, which connects to the **wds1**
    front-end and the **imbs1** OCM control plane back-end.

To deploy workloads on target clusters, users
can now interact with the hosted WDS and ITS API servers.

## Managed Clusters Registration

The next step is to register clusters with OCM. This step uses the same
set of commands defined in the [<u>Open Cluster
Management</u>](https://open-cluster-management.io/) documentation. To
register a cluster “cluster1” for example the steps are (assuming that a
context with name `cluster1` exists in the `kubeconfig` for the user):

1.  “***clusteradm --context imbs1 get token***” get a OCM registration
    token from the OCM instance running in imbs1.

2.  **“*clusteradm join –context cluster1 \<token\> \<ocm apiserver
    address\> –cluster-name cluster1*”** installs the OCM agent
    (Klusterlet) on cluster1 and issue a certificate signing request
    (CSR) to the OCM hub. The CSR is effectively used to request
    approval for joining the cluster under OCM Hub management.

3.  “***clusteradm –context imbs1 accept --clusters cluster1*”**
    approves the CSR from cluster1 and adds cluster1 to OCM Hub
    Management.

### Managed Clusters Labeling

After registering a cluster, users can list available clusters and their
status (via heartbeating from the `OCM Klusterlet`), and label clusters so
that they can define placement policies based on label selectors. For
example:

1.  **“*kubectl --context imbs1 get managedclusters***” lists managed
    clusters and their status

2.  “***kubectl --context imbs1 label managedcluster cluster1
    location-group=edge***” sets the label *location-group=edge* on the
    managed cluster **cluster1.**

### Applying Placement Policies

Placement Policies dictate “What” should be delivered and “Where” it
should be deployed. In the example below, the *nginx-placement* policy
is used to deliver an nginx workload to a set of clusters.

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: nginx-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx"}
EOF
```

This placement configuration determines **where** to deploy the workload by using 
the label selector expressions found in *clusterSelectors*. It also specifies **what** 
to deploy through the downsync.labelSelectors expressions. 
Each matchLabels expression is a criterion for selecting a set of objects based on 
their labels. Other criteria can be added to filter objects based on their namespace, 
api group, resource, and name. If these criteria are not specified, all objects with 
the matching labels are selected. If an object has multiple labels, it is selected 
only if it matches all the labels in the matchLabels expression. 
If there are multiple objectSelectors, an object is selected if it matches any of them. 

This structure ensures precise control over the deployment of
workloads to the appropriate clusters based on the defined criteria.</p>
Note that it is <strong>not</strong> necessary to list all the API
resources to be delivered in the <em><strong>downsync.labelSelectors</strong></em> section. 
If no API resources are specified, all objects matching the label selectors will
be delivered (<em>cluster-scoped</em> or <em>namespace-scoped</em>). API
resources could be specified mainly to provide an additional level of
control. If they are specified, then only objects that are instances of
the specified API resources and match the label selectors are selected
for delivery.

Users apply a placement policy as the one in the example above into
the WDS (e.g, <em><strong>wds1</strong></em>) using standard
<em>kubectl</em> commands:

```shell
kubectl apply -f nginx-placement.yaml
```

KubeStellar adheres to Kubernetes’ best practices by enabling users to apply
objects in any sequence they choose. The responsibility of aligning the
actual state of the system with the intended desired state falls on the
controllers, along with their informers and reconciliation loops.
Whether users decide to apply workloads before placements, or placements
before workloads, the outcome remains consistent.

### Submitting Workloads

Once placement policies are defined and in place, users can just submit
workloads to distribute to one or more clusters the same way that they
submit workloads to a single Kubernetes cluster. An example is
submitting a deployment such as the one below:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: nginx
  labels:
    app.kubernetes.io/name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: public.ecr.aws/nginx/nginx:latest 
        ports:
        - containerPort: 80
```

Assuming the deployment is stored in the file `nginx-deployment.yaml`, users can 
submit the above workload with the command:

```shell
kubectl apply -f nginx-deployment.yaml
```

Since there are no active deployment controllers in a WDS space, the
deployment object is not interpreted and as a result pods are not
started in the WDS space. Rather, the placements controller, which starts
informers on the appropriate Kubernetes API resources, reacts to the new object
being applied, and does the following:

1.  Evaluates if there is one or more placement policies with selectors
    that match the labels on the object.

2.  For each policy match,it evaluates if there is one or more clusters
    matching the *clusterSelectors*, and creates a merged list of
    clusters where the object should be delivered.

3.  If the object has a list of clusters where it should be delivered,
    it is wrapped in an Open Cluster Management `ManifestWork` object and
    copied in the mailbox namespace for each cluster in the ITS.

### Updating Workloads

Updating workloads on the target clusters follows the usual Kubernetes
usage patterns, for example users can update the yaml files for a
workload and re-deploy with kubectl apply.

### Deleting Workloads

Removing workloads from target clusters can be done using standard
Kubernetes commands, such as kubectl delete. Additionally, it’s possible
to retain the workloads within the Workload Description Space (WDS)
while uninstalling them from managed clusters. This can be achieved by
either deleting the placement policy responsible for their deployment or
by altering the placement label selectors so they no longer match the
workloads in question.