<!--where-resolver-process-start-without-cd-kubestellar-start-->
```shell
kubectl ws root:espw
go run cmd/kubestellar-where-resolver/main.go -v 2 &
sleep 45
```
<!--where-resolver-process-start-without-cd-kubestellar-end-->
