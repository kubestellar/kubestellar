package options

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"

	workloadv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/workload/v1alpha1"
)

type Options struct {
	QPS             float32
	Burst           int
	FromKubeconfig  string
	FromContext     string
	FromClusterPath string
	ToKubeconfig    string
	ToContext       string
	SyncTargetName  string
	SyncTargetUID   string
}

func NewOptions() *Options {
	return &Options{
		QPS:   30,
		Burst: 20,
	}
}

func (options *Options) AddFlags(fs *pflag.FlagSet) {
	fs.Float32Var(&options.QPS, "qps", options.QPS, "QPS to use when talking to API servers.")
	fs.IntVar(&options.Burst, "burst", options.Burst, "Burst to use when talking to API servers.")
	fs.StringVar(&options.FromKubeconfig, "from-kubeconfig", options.FromKubeconfig, "Kubeconfig file for -from cluster.")
	fs.StringVar(&options.FromContext, "from-context", options.FromContext, "Context to use in the Kubeconfig file for -from cluster, instead of the current context.")
	fs.StringVar(&options.FromClusterPath, "from-cluster", options.FromClusterPath, "Path of the -from logical cluster.")
	fs.StringVar(&options.ToKubeconfig, "to-kubeconfig", options.ToKubeconfig, "Kubeconfig file for -to cluster. If not set, the InCluster configuration will be used.")
	fs.StringVar(&options.ToContext, "to-context", options.ToContext, "Context to use in the Kubeconfig file for -to cluster, instead of the current context.")
	fs.StringVar(&options.SyncTargetName, "sync-target-name", options.SyncTargetName,
		fmt.Sprintf("ID of the -to cluster. Resources with this ID set in the '%s' label will be synced.", workloadv1alpha1.ClusterResourceStateLabelPrefix+"<ClusterID>"))
	fs.StringVar(&options.SyncTargetUID, "sync-target-uid", options.SyncTargetUID, "The UID from the SyncTarget resource in KCP.")
}

func (options *Options) Complete() error {
	return nil
}

func (options *Options) Validate() error {
	if options.FromClusterPath == "" {
		return errors.New("--from-cluster is required")
	}
	if options.FromKubeconfig == "" {
		return errors.New("--from-kubeconfig is required")
	}
	if options.SyncTargetUID == "" {
		return errors.New("--sync-target-uid is required")
	}
	return nil
}
