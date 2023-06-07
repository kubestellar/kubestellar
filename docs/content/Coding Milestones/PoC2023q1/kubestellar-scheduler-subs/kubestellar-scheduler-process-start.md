<!--kubestellar-scheduler-process-start-start-->
```shell
kubectl ws root:espw
cd ../KubeStellar
go run cmd/kubestellar-scheduler/main.go -v 2 &
sleep 45
```
<!--kubestellar-scheduler-process-start-end-->