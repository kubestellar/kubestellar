/*
Copyright 2022 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestNewKubeStellarSyncerYAML(t *testing.T) {
	expectedYAML := `---
apiVersion: v1
kind: Namespace
metadata:
  name: kubestellar-syncer-sync-target-name-34b23c4k
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubestellar-syncer-sync-target-name-34b23c4k
  namespace: kubestellar-syncer-sync-target-name-34b23c4k
---
apiVersion: v1
kind: Secret
metadata:
  name: kubestellar-syncer-sync-target-name-34b23c4k-token
  namespace: kubestellar-syncer-sync-target-name-34b23c4k
  annotations:
    kubernetes.io/service-account.name: kubestellar-syncer-sync-target-name-34b23c4k
type: kubernetes.io/service-account-token
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubestellar-syncer-sync-target-name-34b23c4k
rules:
- apiGroups:
  - "rbac.authorization.k8s.io"
  resources:
  - clusterroles
  - clusterrolebindings
  verbs:
  - "*"
- apiGroups:
  - "*"
  resources:
  - "*"
  verbs:
  - "*"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubestellar-syncer-sync-target-name-34b23c4k
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubestellar-syncer-sync-target-name-34b23c4k
subjects:
- kind: ServiceAccount
  name: kubestellar-syncer-sync-target-name-34b23c4k
  namespace: kubestellar-syncer-sync-target-name-34b23c4k
---
apiVersion: v1
kind: Secret
metadata:
  name: kubestellar-syncer-sync-target-name-34b23c4k
  namespace: kubestellar-syncer-sync-target-name-34b23c4k
stringData:
  kubeconfig: |
    apiVersion: v1
    kind: Config
    clusters:
    - name: default-cluster
      cluster:
        certificate-authority-data: ca-data
        server: server-url
    contexts:
    - name: default-context
      context:
        cluster: default-cluster
        namespace: kubestellar-namespace
        user: default-user
    current-context: default-context
    users:
    - name: default-user
      user:
        token: token
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubestellar-syncer-sync-target-name-34b23c4k
  namespace: kubestellar-syncer-sync-target-name-34b23c4k
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: kubestellar-syncer-sync-target-name-34b23c4k
  template:
    metadata:
      labels:
        app: kubestellar-syncer-sync-target-name-34b23c4k
    spec:
      containers:
      - name: kubestellar-syncer
        command:
        - /ko-app/syncer
        args:
        - --from-kubeconfig=/kubestellar/kubeconfig
        - --sync-target-name=sync-target-name
        - --sync-target-uid=sync-target-uid
        - --qps=123.4
        - --burst=456
        - --v=3
        env:
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: image
        imagePullPolicy: IfNotPresent
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - name: kubestellar-config
          mountPath: /kubestellar/
          readOnly: true
      serviceAccountName: kubestellar-syncer-sync-target-name-34b23c4k
      volumes:
        - name: kubestellar-config
          secret:
            secretName: kubestellar-syncer-sync-target-name-34b23c4k
            optional: false
`

	actualYAML, err := renderKubeStellarSyncerResources(templateInputForEdge{
		ServerURL:     "server-url",
		Token:         "token",
		CAData:        "ca-data",
		KCPNamespace:  "kubestellar-namespace",
		Namespace:     "kubestellar-syncer-sync-target-name-34b23c4k",
		SyncTarget:    "sync-target-name",
		SyncTargetUID: "sync-target-uid",
		Image:         "image",
		Replicas:      1,
		QPS:           123.4,
		Burst:         456,
	}, "kubestellar-syncer-sync-target-name-34b23c4k")
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(expectedYAML, string(actualYAML)))
}
