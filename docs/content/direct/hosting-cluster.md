# Using existing Kubernetes cluster to host KubeStellar

Status of this document: it is the barest of a start. Much more needs to be written.

## Using an existing Kind cluster as the hosting cluster

This requires a pre-existing Kind cluster that has an Ingress controller that is listening on host port 9443 and configured with TLS passthrough.

The examples say to create a Kind cluster for hosting using the following command.

```shell
kflex init --create-kind
```

To use a pre-existing Kind cluster instead, make sure that your current kubeconfig context is for accessing that cluster and issue the following command.

```shell
kflex init
```

All the subsequent kubectl and helm commands that say to use the kubeconfig context named `kind-kubeflex` need to be modified to use the appropriate kubeconfig context for accessing the hosting cluster.

## Using an existing OpenShift cluster as the hosting cluster

This is similar to using an existing Kind cluster but requires an additional modification. Modify the `kflex` init command and subsequent kubeconfig context references as in the existing-kind-cluster scenario.

Additionally, the recipe for registering a WEC with the ITS needs to be modified. In the `clusteradm` command, omit the `--force-internal-endpoint-lookup` flag. If following the example commands literally, this means to define `flags=""` rather than `flags="--force-internal-endpoint-lookup"`.

## When everything is not on the same machine

Thus far we can only say how to handle this when the hosting cluster is OpenShift. The problem is getting URLs that work from everywhere. OpenShift is a hosted product, your clusters have domain names that are resolvable from everywhere. In other words, if you use an OpenShift cluster as your hosting cluster then this problem is already solved.

## Rule-based customization

KubeStellar can distribute one workload object to multiple WECs, and it is common for users to need some customization to each WEC. By _rule based_ we mean that the customization is not expressed via one or more literal expressions but rather can refer to _properties_ of each WEC by property name. As KubeStellar distributes or transports a workload object from WDS to a WEC, the object can be transformed in a way that depends on those properties.

At its current level of development, KubeStellar has a simple but limited way to specify rule-based customization, called "template expansion".

### Template Expansion

Template expansion is an optional feature that a user can request on an object-by-object basis. The way to request this feature on an object is to put the following annotation on the object.

```yaml
    control.kubestellar.io/expand-templates: "true"
```

The customization that template expansion does when distributing an object from a WDS to a WEC is applied independently to each leaf string of the object and is based on the "text/template" standard pacakge of Go. The string is parsed as a template and then replaced with the result of expanding the template with data supplied by the inventory object (the `ManagedCluster`) describing the WEC. The data is supplied by the object labels and annotations whose key is valid as a Go language identifier, with labels taking precedence over annotations. Errors from this process are reported in the status field of the Binding object involved.

Template expansion can only be applied when and where the un-expanded leaf strings pass the validation that the WDS applies, and can only epxress substring replacements.

For example, consider the following example workload object.

```yaml
apiVersion: logging.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: instance
  namespace: openshift-logging
  annotations:
    control.kubestellar.io/expand-templates: "true"
spec:
  outputs:
    - name: remote-loki
      type: loki
      url: "https://my.loki.server.com/{{.clustername}}"
...
```

The following inventory object provides a value for the `clustername` parameter.

```yaml
apiVersion: cluster.open-cluster-management.io/v1
kind: ManagedCluster
metadata:
  annotations:
    clustername: virgo
  name: virgo
...
```

When distributed to the virgo WEC, that ClusterLogForwarded would say the following.

```yaml
...
      url: "https://my.loki.server.com/virgo"
...
```
