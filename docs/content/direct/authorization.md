# Authorization in Workload Execution Clusters (WECs)

KubeStellar follows the principle of least privilege. It does not grant broad write access in your Workload Execution Clusters (WECs) beyond what is required. Because KubeStellar uses Open Cluster Management (OCM) for distribution, the default permissions for downsyncing are exactly those of the OCM work agent ("klusterlet work agent").

- The work agent runs in the WEC as ServiceAccount `klusterlet-work-sa` in the namespace `open-cluster-management-agent`.
- The baseline, default permissions for the agent are defined by OCM. See OCM’s documentation: https://open-cluster-management.io/docs/concepts/work-distribution/manifestwork/#permission-setting-for-work-agent

If your workload requires the agent to manipulate additional API resources (for example, custom resources not covered by the defaults), you must explicitly expand authorization.

## Ways to expand authorization

There are two supported approaches. Choose whichever best fits your operational model and security controls.

1) Directly grant RBAC in the WEC
- Create a narrowly scoped `ClusterRole` for the needed API groups/resources/verbs and bind it to `klusterlet-work-sa`.
- Do this per WEC, using your preferred process/tooling.

Example (grant access to a CRD group):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: example-crd-access
rules:
- apiGroups: ["example.my.group"]
  resources: ["widgets"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: klusterlet-example-crd-access
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: example-crd-access
subjects:
- kind: ServiceAccount
  name: klusterlet-work-sa
  namespace: open-cluster-management-agent
```

2) Downsync RBAC objects
- Treat RBAC as part of your desired state in the WDS and downsync the `ClusterRole`/`(Cluster)RoleBinding` objects to selected WECs via a `BindingPolicy`.
- This keeps RBAC changes auditable and repeatable in the same pipeline you use for the rest of your workload.

Tip: OCM also supports an aggregated ClusterRole pattern. If you create a `ClusterRole` labeled `open-cluster-management.io/aggregate-to-work: "true"`, its rules are aggregated into the agent’s effective permissions. You can also downsync such a `ClusterRole`.

Example (aggregated ClusterRole):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: open-cluster-management:klusterlet-work:extra-rules
  labels:
    open-cluster-management.io/aggregate-to-work: "true"
rules:
- apiGroups: ["example.my.group"]
  resources: ["widgets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

Refer to OCM’s docs for details on both methods: https://open-cluster-management.io/docs/concepts/work-distribution/manifestwork/#permission-setting-for-work-agent

## Concrete example: Out-of-tree CRDs

Our Example Scenarios show how to grant the agent the additional rights needed to manipulate an out-of-tree CRD (AppWrapper). See Scenario 2 in `./example-scenarios.md#scenario-2---out-of-tree-workload`.

## Why KubeStellar does not auto-grant

Automating privilege expansion would remove the ability for platform teams to maintain independent controls over what is authorized. See Issue #1542 for discussion. Prior subsets of this documentation concern appeared in Issues #2915 and #2947.

## Troubleshooting authorization failures

If a downsynced resource fails with Forbidden errors in the WEC, or a `ManifestWork`/`AppliedManifestWork` shows reconciliation failures:
- Identify the missing verb/resource/apiGroup in the error message.
- Expand authorization using one of the approaches above, scoped as narrowly as possible (only the resources and verbs needed).
- Reconcile again and verify status.

Keep permissions minimal and specific to uphold least privilege.
