<!--where-resolver-imports-start-->
Use the user home workspace (\~) as the workload management workspace (WMW).
```shell
kubectl ws \~
```

Bind APIs.
```shell
kubectl apply -f config/imports/
```
<!--where-resolver-imports-end-->
