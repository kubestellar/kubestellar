package syncers

import edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"

type SyncerInterface interface {
	ReInitializeClients(resources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error
	SyncOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error
	BackStatusOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error
}
