<!--check-command-syncerconfig-the-one-start-->
``` {.bash .hide-me}
let increment=10
let slept=1
while ! kubectl get SyncerConfig the-one -o yaml; do
    if (( slept >= 300 )); then
        echo "FAILURE to run command 'kubectl get SyncerConfig the-one -o yaml' (slept $slept)" >&2
        exit 86
    fi
    sleep $increment
    let slept=slept+increment
done
```
<!--check-command-syncerconfig-the-one-start-->