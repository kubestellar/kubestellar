# kflex-get-kubeconfig

This is a simple utility for fetching a kubeconfig to use to access
the apiservers of a KubeFlex `ControlPlane`.

This utility is given either the name of the `ControlPlane` or a label
selector (in the same format as given to `kubectl`) that is expected to
match the labels of exactly 1 `ControlPlane`.

This utility is told whether to extract the in-cluster or external
kubeconfig.

This utility is also given a file pathname where the fetched
kubeconfig should be written. The special pathname `-` may be passed
to indicate writing to stdout.

Following is an example invocation.

```shell
kflex-get-kubeconfig --output-file taste \
    --control-plane-label-selector kflex.kubestellar.io/cptype=its \
    --in-cluster=false
```

## Command line flags

### Specific to this utility

```console
      --control-plane-label-selector string   label selector that identifies exactly one ControlPlane
      --control-plane-name string             name of ControlPlane to read; mutually exclusive with --control-plane-label-selector
      --in-cluster                            whether to extract the kubeconfig for use in the kubeflex hosting cluster (default true)
      --output-file string                    pathname of file where the kubeconfig will be written; '-' means stdout
```

### Kubernetes client

```console
      ---burst int                            Allowed burst in requests/sec for reading ControlPlane (default 10)
      ---qps float                            Max average requests/sec for reading ControlPlane (default 5)
      --cluster string                        The name of the kubeconfig cluster to use for reading ControlPlane
      --context string                        The name of the kubeconfig context to use for reading ControlPlane
      --kubeconfig string                     Path to the kubeconfig file to use for reading ControlPlane
      --user string                           The name of the kubeconfig user to use for reading ControlPlane
```

### Go logging

```console
      --add_dir_header                        If true, adds the file directory to the header of the log messages
      --alsologtostderr                       log to standard error as well as files (no effect when -logtostderr=true)
      --log_backtrace_at traceLocation        when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                        If non-empty, write log files in this directory (no effect when -logtostderr=true)
      --log_file string                       If non-empty, use this log file (no effect when -logtostderr=true)
      --log_file_max_size uint                Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                           log to standard error instead of files (default true)
      --one_output                            If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true)
      --skip_headers                          If true, avoid header prefixes in the log messages
      --skip_log_headers                      If true, avoid headers when opening log files (no effect when -logtostderr=true)
      --stderrthreshold severity              logs at or above this threshold go to stderr when writing to files and stderr (no effect when -logtostderr=true or -alsologtostderr=true) (default 2)
  -v, --v Level                               number for the log level verbosity
      --vmodule moduleSpec                    comma-separated list of pattern=N settings for file-filtered logging
```
