<!--create-files-and-contexts-for-openshift-clusters-start-->
**important:** alias the kubernetes contexts of the OpenShift clusters you provided to match their use in this guide
```
oc login <ks-core OpenShift cluster>
CURRENT_CONTEXT=$(KUBECONFIG=~/.kube/config kubectl config current-context) \
    && KUBECONFIG=~/.kube/config kubectl config set-context ks-core \
    --namespace=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $1}') \
    --cluster=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $2}') \
    --user=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $3"/"$2}')

oc login <ks-edge-cluster1 OpenShift cluster>
CURRENT_CONTEXT=$(KUBECONFIG=~/.kube/config kubectl config current-context) \
    && KUBECONFIG=~/.kube/config kubectl config set-context ks-edge-cluster1 \
    --namespace=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $1}') \
    --cluster=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $2}') \
    --user=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $3"/"$2}')

oc login <ks-edge-cluster2 OpenShift cluster>
CURRENT_CONTEXT=$(KUBECONFIG=~/.kube/config kubectl config current-context) \
    && KUBECONFIG=~/.kube/config kubectl config set-context ks-edge-cluster2 \
    --namespace=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $1}') \
    --cluster=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $2}') \
    --user=$(echo "$CURRENT_CONTEXT" | awk -F '/' '{print $3"/"$2}')

```
<!--create-files-and-contexts-for-openshift-clusters-end-->