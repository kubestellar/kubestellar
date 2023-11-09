<!--kubestellar-syncer-1-syncer-gen-plugin-start-->
Generate UUID for Syncer identification.
```shell
syncer_id="syncer-"`uuidgen | tr '[:upper:]' '[:lower:]'`
```

Go to a workspace.
```shell
kubectl ws root
kubectl ws create ws1 --enter
```

Create the following APIBinding in the workspace (Note that in the case of mailbox workspaces, it's done by mailbox controller at creating the mailbox workspace.)
```shell
cat << EOL | kubectl apply -f -
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-espw
spec:
  reference:
    export:
      path: root:espw
      name: edge.kubestellar.io
EOL
```

Create a serviceaccount in the workspace.
```shell
cat << EOL | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: $syncer_id
EOL
```

Create clusterrole and clusterrolebinding to bind the serviceaccount to the role.
```shell
cat << EOL | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: $syncer_id
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
- nonResourceURLs: ["/"]
  verbs: ["access"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: $syncer_id
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: $syncer_id
subjects:
- apiGroup: ""
  kind: ServiceAccount
  name: $syncer_id
  namespace: default
EOL
```

Wait a little for the secret to be created for kubernetes cluster with LegacyServiceAccountTokenNoAutoGeneration disabled and get the token from the secret. 
```shell
token=""
count=0
while [[ $count -lt 5 ]]
do
  secret_name=`kubectl get secret -o custom-columns=":.metadata.name"| grep $syncer_id` || true
  if [[ "$secret_name" == "" ]];then
    echo retry in 1s
    sleep 1
    let count=count+1
  else
    token=`kubectl get secret $secret_name -o jsonpath='{.data.token}' | base64 -d`
    break
  fi
done
```

If the SA token secret is not created, it means the kubernetes version is equal or later then v1.24. In that case, generate SA token manually.
```shell
if [[ "$token" == "" ]];then
  token=`kubectl create token $syncer_id`
fi
```

Get the certificates that will be set in the upstream kubeconfig manifest.
```shell
cacrt=`kubectl config view --minify --raw | yq ".clusters[0].cluster.certificate-authority-data"`
```

Get server_url that will be set in the upstream kubeconfig manifest.
```shell
server_url=`kubectl config view --minify --raw | yq ".clusters[0].cluster.server" | sed -e 's|https://\(.*\):\([^/]*\)/.*|https://\1:\2|g'`
```

Set some other parameters.</br>
a. downstream_namespace where Syncer Pod runs
```shell
downstream_namespace="kubestellar-$syncer_id"
```
b. Syncer image
```shell
image="quay.io/kubestellar/syncer:v0.2.2"
```

Download manifest template.
```shell
curl -LO {{ config.repo_raw_url }}/main/pkg/syncer/scripts/kubestellar-syncer-bootstrap.template.yaml
```

Generate manifests to bootstrap KubeStellar-Syncer.
```shell
syncer_id=$syncer_id cacrt=$cacrt token=$token server_url=$server_url downstream_namespace=$downstream_namespace image=$image envsubst < kubestellar-syncer-bootstrap.template.yaml
```
```
---
apiVersion: v1
kind: Namespace
metadata:
  name: kubestellar-syncer-9ee90de6-eb76-4ddb-9346-c4c8d92075e1
---
apiVersion: v1
kind: ServiceAccount
metadata:
...
```
<!--kubestellar-syncer-1-syncer-gen-plugin-end-->