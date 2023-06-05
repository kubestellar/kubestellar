
cd /Users/andan02/projects/edge-mc/docs/scripts/
set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

git clone https://github.com/kcp-dev/edge-mc KubeStellar

git clone -b v0.11.0 https://github.com/kcp-dev/kcp kcp

cd kcp
make build
export PATH=$(pwd)/bin:$PATH

kcp start &
sleep 30  # wait for KCP to initialize

export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
export PATH=$(pwd)/bin:$PATH

kubectl ws root
kubectl ws create espw

kubectl ws root:espw
cd ../KubeStellar
go run ./cmd/mailbox-controller -v=2 &
sleep 15  # wait a few seconds for the mailbox controller to initialize

kubectl ws \~
kubectl ws create imw --enter

cat <<EOF | kubectl apply -f -
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: stest1
spec:
  cells:
    foo: bar
EOF

kubectl ws root:espw

kubectl get workspaces

kubectl ws \~
kubectl ws imw
kubectl delete SyncTarget stest1
