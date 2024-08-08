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
KubeStellar rules are designed to not be very inconsistent with the
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

- Remember that for each of your BindingPolicy objects, there is a corresponding Binding object that reports what is matching the policy object.
- Although not part of the interface, when debugging you can look at the ManifestWork and WorkStatus objects in the ITS.
- Look at logs of controllers. If they have had container restarts that look relevant, look also at the previous logs. Do not forget OCM controllers. Do not forget that some Pods have more than one interesting container.
- If a controller's `-v` is not at least 5, increase it.
- Remember that Kubernetes controllers tend to report transient problems as errors without making it clear that the problem is transient and tend to not make it clear if/when the problem has been resolved (sigh).

## Making a good trouble report

Basic configuration information.

- Include the versions of all the relevant software; do not forget the OCM pieces.
- Report on each Kubernetes/OCP cluster involved. What sort of cluster is it (kind, k3d, OCP, ...)? What version of that?
- For each WDS and ITS involved, report on what sort of thing is playing that role (remember that a Space is a role) --- a new KubeFlex control plane (report type) or an existing cluster (report which one).

Do a simple clean demonstration of the problem, if possible.

Show the particulars of something going wrong.

- Report timestamps of when salient changes happened. Make it clear which timezone is involved in each one. Particularly interesting times are when KubeStellar did the wrong thing or failed to do anything at all in response to something.
- Look at the Binding and ManifestWork and WorkStatus objects and the controller logs. Include both in a problem report. Show the relevant workload objects, from WDS and from WEC. When the problem is behavior over time, show the objects contents from before and after the misbehavior.
- When reporting kube API object contents, include the `meta.managedFields`. For example, when using `kubectl get`, include `--show-managed-fields`.
