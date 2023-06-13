package clusterproviderclient

import (
	"context"

	"sigs.k8s.io/logical-cluster"

	clusterprovider "github.com/kcp-dev/edge-mc/cluster-provider-client/cluster"
	kindprovider "github.com/kcp-dev/edge-mc/cluster-provider-client/kind"
	v1alpha1apis "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
)

var ProviderList map[string]ProviderClient

// ES: return and error and don't panic, move push to outside the case
func GetProviderClient(providerType v1alpha1apis.ClusterProviderType, providerName string) ProviderClient {
	key := string(providerType) + "-" + providerName
	newProvider, exists := ProviderList[key]
	if !exists {
		switch providerType {
		case v1alpha1apis.KindProviderType:
			newProvider = kindprovider.New(providerName)
			ProviderList[key] = newProvider
		default:
			panic("unknown provider type")
		}
	}

	return newProvider
}

// Provider defines methods to retrieve, list, and watch fleet of clusters.
// The provider is responsible for discovering and managing the lifecycle of each
// cluster.
//
// Example:
// ES: add example, remove list() add Get()
type ProviderClient interface {
	Create(ctx context.Context, name logical.Name, opts clusterprovider.Options) (clusterprovider.LogicalClusterInfo, error)
	Delete(ctx context.Context, name logical.Name, opts clusterprovider.Options) error

	// List returns a list of logical clusters.
	// This method is used to discover the initial set of logical clusters
	// and to refresh the list of logical clusters periodically.
	List() ([]logical.Name, error)

	// Watch returns a Watcher that watches for changes to a list of logical clusters
	// and react to potential changes.
	// TODO: I have yet to implement it for the kind provider type, so commenting it out for now
	Watch() (clusterprovider.Watcher, error)
}
