<!--kubestellar-scheduler-process-start-without-cd-kubestellar-start-->
```shell
go run cmd/kubestellar-scheduler/main.go -v 2 &
sleep 15  # wait a few seconds for the kubestellar scheduler to initialize
```
<!--kubestellar-scheduler-process-start-without-cd-kubestellar-end-->