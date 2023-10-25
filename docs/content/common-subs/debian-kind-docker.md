<!--debian-kind-docker-start-->
on Debian, the syncers on ks-edge-cluster1 and ks-edge-cluster2 will not resolve the kubestellar.core hostname 

You have 2 choices:  

1. Use the value of `hostname -f` instead of kubestellar.core as your "EXTERNAL_HOSTNAME" in "Step 1:  Deploy the KubeStellar Core Component", or

2. Just before step 6 in the KubeStellar User Quickstart for Kind do the following

Add IP/domain to /etc/hosts of cluster1/cluster2 containers (replace with appropriate IP address):
 
```
docker exec -it $(docker ps | grep ks-edge-cluster1 | cut -d " " -f 1) \
    sh -c "echo '192.168.122.144 kubestellar.core' >> /etc/hosts"
docker exec -it $(docker ps | grep ks-edge-cluster2 | cut -d " " -f 1) \
    sh -c "echo '192.168.122.144 kubestellar.core' >> /etc/hosts"
```

Edit coredns ConfigMap for cluster1 and cluster1 (see added lines in example):

```
KUBECONFIG=~/.kube/config kubectl edit cm coredns -n kube-system --context=ks-edge-cluster1
KUBECONFIG=~/.kube/config kubectl edit cm coredns -n kube-system --context=ks-edge-cluster2
```

add the highlighted information
``` hl_lines="9 10 11 27 28"
apiVersion: v1
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
        hosts /etc/coredns/customdomains.db core {  
          fallthrough                               
        }                                          
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . /etc/resolv.conf {
           max_concurrent 1000
        }
        cache 30
        loop
        reload
        loadbalance
    }
  customdomains.db: |                   
    192.168.122.144 kubestellar.core  
kind: ConfigMap
metadata:
  creationTimestamp: "2023-10-24T19:18:05Z"
  name: coredns
  namespace: kube-system
  resourceVersion: "10602"
  uid: 3930c18f-23e8-4d0b-9ddf-658fdf3cb20f
```

Edit Deployment for coredns on cluster1 and cluster2, adding the key/path at the given location:
```
KUBECONFIG=~/.kube/config kubectl edit -n kube-system \
    deployment coredns --context=ks-edge-cluster1
KUBECONFIG=~/.kube/config kubectl edit -n kube-system \
    deployment coredns --context=ks-edge-cluster2
```

``` hl_lines="7 8"
spec:
  template:
    spec:
      volumes:
      - configMap:
          items:
          - key: customdomains.db
            path: customdomains.db
```

Restart coredns pods:
```
KUBECONFIG=~/.kube/config kubectl rollout restart \
    -n kube-system deployment/coredns --context=ks-edge-cluster1
KUBECONFIG=~/.kube/config kubectl rollout restart \
    -n kube-system deployment/coredns --context=ks-edge-cluster2
```

__(adapted from "The Cluster-wise solution" at [https://stackoverflow.com/questions/37166822/is-there-a-way-to-add-arbitrary-records-to-kube-dns](https://stackoverflow.com/questions/37166822/is-there-a-way-to-add-arbitrary-records-to-kube-dns))__

<!--debian-kind-docker-end-->
