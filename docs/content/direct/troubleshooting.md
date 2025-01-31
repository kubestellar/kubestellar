# Troubleshooting

This guide is a work in progress.

## Debug log levels

The KubeStellar controllers take an optional command line flag that
sets the level of debug logging to emit. Each debug log message is
associated with a _log level_, which is a non-negative integer. Higher
numbers correspond to messages that appear more frequently and/or give
more details. The flag's name is `-v` and its value sets the highest
log level that gets emitted; higher level messages are suppressed.

The KubeStellar debug log messages are assigned to log levels roughly
according to the following rules. Note that the various Kubernetes
libraries used in these controllers also emit leveled debug log
messages, according to their own numbering conventions. The
KubeStellar rules are designed to be mostly consistent with the
Kubernetes practice.

- **0**: messages that appear O(1) times per run.
- **1**: more detailed messages that appear O(1) times per run.
- **2**: messages that appear O(1) times per lifecycle event of an API object or important conjunction of them (e.g., when a Binding associates a workload object with a WEC).
- **3**: more detailed messages that appear O(1) times per lifecycle event of an API object or important conjunction of them.
- **4**: messages that appear O(1) times per sync. A sync is when a controller reads the current state of one API object and reacts to that.
- **5**: more detailed messages that appear O(1) times per sync.

The [core Helm chart](core-chart.md) has "values" that set the
verbosity (`-v`) of various controllers.

## Things to look at

- While double-checking your input is never bad, using `kubectl get -o yaml --show-managed-fields` to examine the live API objects adds some good stuff: confirmation that your input was received and parsed as expected, display of any error messages in your API objects, timestamps in the metadata (helpful for comparing with log messages), indication of what last wrote to each part of your API objects and when.
- When basic stuff is not working, survey the Pod objects in the KubeFlex hosting cluster to look for ones that are damaged in some way. For example: you can get a summary with the command `kubectl --context kind-kubeflex get pods -A` --- adjust as necessary for the name of your kubeconfig context to use for the KubeFlex hosting cluster.
- Remember that for each of your BindingPolicy objects, there is a corresponding Binding object that reports what is matching the policy object.
- Although not part of the interface, when debugging you can look at the ManifestWork and WorkStatus objects in the ITS.
- More broadly, remember that KubeStellar uses OCM.
- Look at logs of controllers. If they have had container restarts that look relevant, look also at the previous logs. Do not forget OCM controllers. Do not forget that some Pods have more than one interesting container.
    - Remember that the amount of log retained is typically a configured option in the relevant container runtime. If your logs are too short, look into increasing that log retention.
- If a controller's `-v` is not at least 5, increase it.
- Remember that Kubernetes controllers tend to report transient problems as errors without making it clear that the problem is transient and tend to not make it clear if/when the problem has been resolved (sigh).

## Some known problems

We have [the start of a list](known-issues.md).

## Making a good trouble report

Basic configuration information.

- Include the versions of all the relevant software; do not forget the OCM pieces.
- Report on each Kubernetes/OCP cluster involved. What sort of cluster is it (kind, k3d, OCP, ...)? What version of that?
- For each WDS and ITS involved, report on what sort of thing is playing that role (remember that a Space is a role) --- a new KubeFlex control plane (report type) or an existing cluster (report which one).

Do a simple clean demonstration of the problem, if possible.

Show the particulars of something going wrong.

- Show a shell session, starting from scratch
- Report timestamps of when salient changes happened. Make it clear which timezone is involved in each one. Particularly interesting times are when KubeStellar did the wrong thing or failed to do anything at all in response to something.
- Show the relevant API objects. When the problem is behavior over time, show the objects contents from before and after the misbehavior.
    - In the WDS: the workload objects involved; any `BidingPolicy` involved, and the corresponding `Binding` for each; any `CustomTransform`, `StatusCollector`, or `CombinedStatus` involved.
    - Any involved objects in the WEC(s).
    - Implementation objects in the ITS: `ManifestWork`, `WorkStatus`.
    - Here is one way to show the evolution of a relevant set of objects over time. The following command displays the `ManifestWork` objects, after creation and after each update (modulo the gaps allowed by eventual consistency), in an ITS as addressed by the kubeconfig context named `its1` --- after first listing the existing objects. Each line is prefixed with the hour:minute:second at which it appears.
        ```shell
        kubectl --context its1 get manifestworks -A --show-managed-fields -o yaml --watch | while IFS="" read line; do echo "$(date +%T)| $line"; done
        ```
- When reporting kube API object contents, include the `meta.managedFields`. For example, when using `kubectl get`, include `--show-managed-fields`.
- Show the logs from relevant controllers. The most active and directly relevant ones are the following.
    - The KubeStellar controller-manager (running in the KubeFlex hosting cluster) for the WDS
    - KubeStellar's OCM-based transport-controller (running in the KubeFlex hosting cluster) for the WDS+ITS
    - The OCM Status Add-On Agent in the WEC.
    - OCM's klusterlet-agent in the WEC.

### Use the snapshot script

There is a script that is intended to capture a lot of relevant state;
using it can help make a good trouble report.

You can use a command like the following to invoke the script.

```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/refs/heads/main/scripts/kubestellar-snapshot.sh) -V -Y -L
```

Report the log of running the script.

If the script is successful then it will create an archive file and
tell you about it; include that file in your trouble report.
