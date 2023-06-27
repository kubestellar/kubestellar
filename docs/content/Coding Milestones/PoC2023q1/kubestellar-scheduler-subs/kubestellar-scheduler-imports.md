<!--kubestellar-scheduler-imports-start-->
Use the user home workspace (\~) as the workload management workspace (WMW).
```shell
kubectl ws \~
```

Bind APIs.
```shell
kubectl apply -f ../kubestellar/config/imports/
```
<!--kubestellar-scheduler-imports-end-->