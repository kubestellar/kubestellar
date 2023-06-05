<!--kubestellar-scheduler-process-start-start-->
```shell
kubectl ws root:espw
cd ../KubeStellar
go run cmd/kubestellar-scheduler/main.go -v 2 &
sleep 45  # wait a few seconds for the kubestellar scheduler to initialize
```
<!--kubestellar-scheduler-process-start-end-->