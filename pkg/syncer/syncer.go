package syncer

import (
	"context"
	"time"

	kcpdynamic "github.com/kcp-dev/client-go/dynamic"
	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"
	kcpclientset "github.com/kcp-dev/edge-mc/pkg/syncer/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/edge-mc/pkg/syncer/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kcp-dev/edge-mc/pkg/syncer/controller"
	"github.com/kcp-dev/logicalcluster/v3"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
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

	syncConfigClientSet, err := kcpclientset.NewForConfig(bootstrapConfig)
	if err != nil {
		return err
	}
	syncConfigClient := syncConfigClientSet.EdgeV1alpha1().EdgeSyncConfigs()

	// syncConfigInformerFactory to watch a certain syncConfig on upstream
	syncConfigInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(syncConfigClientSet, resyncPeriod, kcpinformers.WithTweakListOptions(
		func(listOptions *metav1.ListOptions) {
			listOptions.FieldSelector = fields.OneTermEqualSelector("metadata.name", cfg.SyncTargetName).String()
		},
	))

	syncConfigInformerFactory.Start(ctx.Done())
	syncConfigInformerFactory.WaitForCacheSync(ctx.Done())

	syncConfigInformer := syncConfigInformerFactory.Edge().V1alpha1().EdgeSyncConfigs()

	upstreamConfig := rest.CopyConfig(cfg.UpstreamConfig)
	rest.AddUserAgent(upstreamConfig, "edge-mc#syncer/"+kcpVersion)
	upstreamSyncerClusterClient, err := kcpdynamic.NewForConfig(upstreamConfig)
	if err != nil {
		return err
	}
	_ = upstreamSyncerClusterClient

	downstreamConfig := rest.CopyConfig(cfg.DownstreamConfig)
	rest.AddUserAgent(downstreamConfig, "edge-mc#syncer/"+kcpVersion)
	downstreamDynamicClient, err := dynamic.NewForConfig(downstreamConfig)
	if err != nil {
		return err
	}
	downstreamKubeClient, err := kubernetes.NewForConfig(downstreamConfig)
	if err != nil {
		return err
	}
	downstreamSyncerDiscoveryClient := discovery.NewDiscoveryClient(downstreamKubeClient.RESTClient())

	controller, err := controller.NewSyncConfigController(logger, syncConfigClient, syncConfigInformer, types.UID(cfg.SyncTargetUID), cfg.SyncTargetName, downstreamDynamicClient, downstreamSyncerDiscoveryClient)
	if err != nil {
		return err
	}
	_ = controller
	return nil
}
