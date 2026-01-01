# Authorization in Workload Execution Clusters (WECs)

KubeStellar aims to follow the principle of least privilege by not adding any privileges beyond those required by its dependencies. Because KubeStellar uses Open Cluster Management (OCM) for distribution, the default permissions for downsyncing are exactly those of the OCM work agent ("klusterlet work agent"). OCM provides broad default permissions for convenience, which KubeStellar inherits. You can customize these permissions to meet your specific security requirements.

- The work agent runs in the WEC as ServiceAccount `klusterlet-work-sa` in the namespace `open-cluster-management-agent`.
- The baseline, default permissions for the agent are defined by OCM. See OCM's documentation: [Permission setting for work agent](https://open-cluster-management.io/docs/concepts/work-distribution/manifestwork/#permission-setting-for-work-agent)

If you want the agent to manipulate additional API resources (for example, custom resources not covered by the defaults), you must explicitly expand authorization.

## Ways to expand authorization

There are two supported approaches. Choose whichever best fits your operational model and security controls.

1) Directly grant RBAC in the WEC
- Create a `ClusterRole` for the needed API groups/resources/verbs and bind it to `klusterlet-work-sa` (consider limiting scope to only what's necessary for your use case).
- Do this per WEC, using your preferred process/tooling.

Example (grant access to custom resources):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: example-crd-access  # Use a descriptive name for your use case
rules:
- apiGroups: ["example.my.group"]  # Replace with your actual API group
  resources: ["widgets"]           # Replace with your custom resources
  verbs: ["get", "list", "watch", "create", "update", "patch"]  # Recommendation: add only verbs you need
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: klusterlet-example-crd-access  # Should identify your ClusterRole purpose
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: example-crd-access  # Must match the ClusterRole name above
subjects:
- kind: ServiceAccount
  name: klusterlet-work-sa           # This is the OCM work agent service account
  namespace: open-cluster-management-agent  # This is the OCM agent namespace
```

2) Downsync RBAC objects
- Treat RBAC as part of your desired state in the WDS and downsync the `ClusterRole`/`(Cluster)RoleBinding` objects to selected WECs via a `BindingPolicy`.
- This keeps RBAC changes auditable and repeatable in the same pipeline you use for the rest of your workload.
- Note that this approach is a specific example of approach 1, where KubeStellar serves as your preferred process/tooling for managing RBAC.

Tip: OCM also supports an aggregated ClusterRole pattern. If you create a `ClusterRole` labeled `open-cluster-management.io/aggregate-to-work: "true"`, its rules are automatically included in the existing `klusterlet-work-sa` permissions through OCM's aggregation mechanism. You can also downsync such a `ClusterRole`. With the aggregated pattern, no separate ClusterRoleBinding is needed - OCM handles the permission binding automatically.

Example (aggregated ClusterRole):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: example-open-cluster-management:klusterlet-work:extra-rules  # Use descriptive naming
  labels:
    open-cluster-management.io/aggregate-to-work: "true"  # This label enables auto-aggregation
rules:
- apiGroups: ["example.my.group"]  # Replace with your actual API group
  resources: ["widgets"]           # Replace with your custom resources
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]  # Include needed verbs
```

Refer to OCM's docs for details on both methods: [Permission setting for work agent](https://open-cluster-management.io/docs/concepts/work-distribution/manifestwork/#permission-setting-for-work-agent)

## Concrete example: Out-of-tree CRDs

Scenario 2 shows how to grant the agent the additional rights needed to manipulate an out-of-tree CRD (AppWrapper). See [Scenario 2](./example-scenarios.md#scenario-2-out-of-tree-workload) in the Example Scenarios.

## Why KubeStellar does not auto-grant

Automating privilege expansion would remove the ability for platform teams to maintain independent controls over what is authorized. See [Issue #1542](https://github.com/kubestellar/kubestellar/issues/1542) for discussion. Prior subsets of this documentation concern appeared in [Issue #2915](https://github.com/kubestellar/kubestellar/issues/2915) and [Issue #2947](https://github.com/kubestellar/kubestellar/issues/2947).

## Troubleshooting authorization failures

If a downsynced resource fails with Forbidden errors in the WEC, or a `ManifestWork`/`AppliedManifestWork` shows reconciliation failures:
- Identify the missing verb/resource/apiGroup in the error message.
- Expand authorization using one of the approaches above (consider your security requirements when determining scope).
- Reconcile again and verify status.
