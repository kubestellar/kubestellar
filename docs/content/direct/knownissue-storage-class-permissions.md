# StorageClass Permissions Issue in WEC Clusters

## Issue

When attempting to deploy cluster-scoped resources like `StorageClass` objects from a WDS to a Workload Edge Cluster (WEC), the synchronization fails with permission errors like:

```yaml
Failed to apply manifest: storageclasses.storage.k8s.io "local-storage" is forbidden: User "system:serviceaccount:open-cluster-management-agent:klusterlet-work-sa" cannot get resource "storageclasses" in API group "storage.k8s.io" at the cluster scope
```

## Root Cause

In the Open Cluster Management (OCM) setup that KubeStellar uses for resource distribution, the agent running on each managed cluster (WEC) uses a service account called `klusterlet-work-sa` in the `open-cluster-management-agent` namespace. By default, this service account doesn't have sufficient RBAC permissions to interact with certain cluster-scoped resources, including `StorageClass`.

## Solution

You need to apply additional RBAC permissions to grant the `klusterlet-work-sa` service account access to manage `StorageClass` resources. We provide a helper script that applies these permissions:

```bash
# Apply to a specific WEC cluster context
scripts/apply-ocm-storage-permissions.sh --context <your-wec-context>
```

If you're setting up a new environment using the standard setup script, these permissions are now applied automatically.

## Manual Fix

If you prefer to apply the permissions manually or want to understand what's being done, here are the RBAC resources that need to be applied to each WEC:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: klusterlet-work-sa-storage
rules:
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: klusterlet-work-sa-storage
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: klusterlet-work-sa-storage
subjects:
- kind: ServiceAccount
  name: klusterlet-work-sa
  namespace: open-cluster-management-agent
```

## Other Cluster-Scoped Resources

If you encounter similar permission issues with other cluster-scoped resources, you may need to expand the ClusterRole to include those resource types. The same pattern applies - add the relevant API group and resources to the ClusterRole's rules section.
