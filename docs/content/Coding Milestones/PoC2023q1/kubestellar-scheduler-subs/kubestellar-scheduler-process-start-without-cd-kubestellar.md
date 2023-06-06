<!--kubestellar-scheduler-process-start-without-cd-kubestellar-start-->
```shell
kubectl ws root:espw
go run cmd/kubestellar-scheduler/main.go -v 2 &
sleep 45
```
<!--kubestellar-scheduler-process-start-without-cd-kubestellar-end-->