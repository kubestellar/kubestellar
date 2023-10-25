<!--kubestellar-test-apache-start-->
Now, let's check that the deployment was created in the kind **ks-edge-cluster1** cluster (it may take up to 30 seconds to appear):
```shell
KUBECONFIG=~/.kube/config kubectl --context ks-edge-cluster1 get deployments -A
```

you should see output including:
``` { .sh .no-copy }
NAMESPACE           NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
my-namespace        my-first-kubestellar-deployment    1/1        1            1       6m48s
```

And, check the **ks-edge-cluster2** kind cluster for the same:
```shell
KUBECONFIG=~/.kube/config kubectl --context ks-edge-cluster2 get deployments -A
```

you should see output including:
``` { .sh .no-copy }
NAMESPACE           NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
my-namespace        my-first-kubestellar-deployment    1/1        1            1       7m54s
```

Finally, let's check that the workload is working in both clusters:
For **ks-edge-cluster1**:
```shell
while [[ $(KUBECONFIG=~/.kube/config kubectl --context ks-edge-cluster1 get pod \
  -l "app=common" -n my-namespace -o jsonpath='{.items[0].status.phase}') != "Running" ]]; do 
    sleep 5; 
  done;
curl http://localhost:8094
```

you should see the output:
```html
<!DOCTYPE html>
<html>
  <body>
    This web site is hosted on ks-edge-cluster1 and ks-edge-cluster2.
  </body>
</html>
```

For **ks-edge-cluster2**:
```shell
while [[ $(KUBECONFIG=~/.kube/config kubectl --context ks-edge-cluster2 get pod \
  -l "app=common" -n my-namespace -o jsonpath='{.items[0].status.phase}') != "Running" ]]; do 
    sleep 5; 
  done;
curl http://localhost:8096
```

you should see the output:
```html
<!DOCTYPE html>
<html>
  <body>
    This web site is hosted on ks-edge-cluster1 and ks-edge-cluster2.
  </body>
</html>
```
<!--kubestellar-test-apache-end-->