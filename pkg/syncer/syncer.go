package syncer

import (
	"context"
	"time"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"
	kcpclientset "github.com/kcp-dev/edge-mc/pkg/syncer/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/edge-mc/pkg/syncer/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"

	"github.com/kcp-dev/logicalcluster/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/version"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type SyncerConfig struct {
	UpstreamConfig   *rest.Config
	DownstreamConfig *rest.Config
	SyncTargetPath   logicalcluster.Path
	SyncTargetName   string
	SyncTargetUID    string
}

type Syncer struct {
	logger               klog.Logger
	syncConfigNamespace  string
	syncConfigName       string
	syncConfigKubeClient *kubernetes.Clientset
	downSyncedRedources  []edgev1alpha1.EdgeSyncConfigResource
	upSyncedRedources    []edgev1alpha1.EdgeSyncConfigResource
}

type SyncerInterface interface {
	initializeClients([]edgev1alpha1.EdgeSyncConfigResource) error
	ReInitializeClients([]edgev1alpha1.EdgeSyncConfigResource) error
	getClients(edgev1alpha1.EdgeSyncConfigResource)
}

const (
	resyncPeriod = 10 * time.Hour
)

func StartSyncer(ctx context.Context, cfg *SyncerConfig, numSyncerThreads int, importPollInterval time.Duration, syncConfigNamespace string) error {
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("syncTarget.name", cfg.SyncTargetName)
	logger.V(2).Info("starting edge-mc syncer")
	kcpVersion := version.Get().GitVersion

	bootstrapConfig := rest.CopyConfig(cfg.UpstreamConfig)
	rest.AddUserAgent(bootstrapConfig, "edge-mc#syncer/"+kcpVersion)

	syncConfigClient, err := kcpclientset.NewForConfig(bootstrapConfig)
	if err != nil {
		return err
	}

	// syncConfigInformerFactory to watch a certain syncConfig on upstream
	syncConfigInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(syncConfigClient, resyncPeriod, kcpinformers.WithTweakListOptions(
		func(listOptions *metav1.ListOptions) {
			listOptions.FieldSelector = fields.OneTermEqualSelector("metadata.name", cfg.SyncTargetName).String()
		},
	))

	syncConfigInformerFactory.Start(ctx.Done())
	syncConfigInformerFactory.WaitForCacheSync(ctx.Done())

	upstreamConfig := rest.CopyConfig(cfg.UpstreamConfig)
	rest.AddUserAgent(upstreamConfig, "edge-mc#syncer/"+kcpVersion)
	downstreamConfig := rest.CopyConfig(cfg.DownstreamConfig)
	rest.AddUserAgent(downstreamConfig, "edge-mc#syncer/"+kcpVersion)

	return nil
}
