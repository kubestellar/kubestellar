package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrlutil "github.com/kubestellar/kubestellar/pkg/ctrlutil"
	kfv1alpha1 "github.com/kubestellar/kubeflex/api/v1alpha1"
	kslclient "github.com/kubestellar/kubestellar/pkg/client"
)

type ObjectIdentifier struct {
	Group     string
	Version   string
	Kind      string
	Name      string
	Namespace string
}

// DetectDuplicateObjectIdentifiersAcrossWDSes checks for duplicate object identifiers across all WDSes and logs warnings if found.
func DetectDuplicateObjectIdentifiersAcrossWDSes(logger logr.Logger) error {
	ctx := context.Background()
	kubeClient := *kslclient.GetClient()
	labelSelector := labels.SelectorFromSet(labels.Set(map[string]string{
		ctrlutil.ControlPlaneTypeLabel: ctrlutil.ControlPlaneTypeWDS,
	}))
	wdsList := &kfv1alpha1.ControlPlaneList{}
	if err := kubeClient.List(ctx, wdsList, &client.ListOptions{LabelSelector: labelSelector}); err != nil {
		return fmt.Errorf("failed to list WDS control planes: %w", err)
	}

	idToWDS := map[ObjectIdentifier][]string{}

	for _, wds := range wdsList.Items {
		wdsName := wds.Name
		restConfig, _, err := ctrlutil.GetWDSKubeconfig(logger, wdsName)
		if err != nil {
			logger.Error(err, "Failed to get kubeconfig for WDS", "wds", wdsName)
			continue
		}
		dynClient, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			logger.Error(err, "Failed to create dynamic client for WDS", "wds", wdsName)
			continue
		}
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
		if err != nil {
			logger.Error(err, "Failed to create discovery client for WDS", "wds", wdsName)
			continue
		}
		apiResourceLists, err := discoveryClient.ServerPreferredResources()
		if err != nil {
			logger.Error(err, "Failed to discover resources for WDS", "wds", wdsName)
			continue
		}
		for _, apiResourceList := range apiResourceLists {
			gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
			if err != nil {
				continue
			}
			for _, resource := range apiResourceList.APIResources {
				// Skip subresources, CRDs, and control objects
				if strings.Contains(resource.Name, "/") || !contains(resource.Verbs, "list") || !resource.Namespaced && !resource.KindIsNamespaced() {
					continue
				}
				gvr := gv.WithResource(resource.Name)
				var objList *unstructuredList
				var err error
				if resource.Namespaced {
					objList, err = listAllNamespaces(ctx, dynClient, gvr)
				} else {
					objList, err = listClusterScoped(ctx, dynClient, gvr)
				}
				if err != nil {
					logger.Error(err, "Failed to list objects", "gvr", gvr, "wds", wdsName)
					continue
				}
				for _, obj := range objList.Items {
					id := ObjectIdentifier{
						Group:     gvr.Group,
						Version:   gvr.Version,
						Kind:      resource.Kind,
						Name:      obj.GetName(),
						Namespace: obj.GetNamespace(),
					}
					idToWDS[id] = append(idToWDS[id], wdsName)
				}
			}
		}
	}

	// Report duplicates
	for id, wdses := range idToWDS {
		if len(wdses) > 1 {
			logger.Error(nil, "Duplicate object identifier across WDSes", "identifier", id, "wdses", wdses)
		}
	}
	return nil
}

// Helper: check if a verb is in the list
func contains(verbs []string, verb string) bool {
	for _, v := range verbs {
		if v == verb {
			return true
		}
	}
	return false
}

// Helper: list all objects in all namespaces for a namespaced resource
type unstructuredList struct {
	Items []metav1.Object
}

func listAllNamespaces(ctx context.Context, dynClient dynamic.Interface, gvr schema.GroupVersionResource) (*unstructuredList, error) {
	// List in all namespaces
	ul, err := dynClient.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	items := make([]metav1.Object, 0, len(ul.Items))
	for i := range ul.Items {
		items = append(items, &ul.Items[i])
	}
	return &unstructuredList{Items: items}, nil
}

// Helper: list all objects for a cluster-scoped resource
func listClusterScoped(ctx context.Context, dynClient dynamic.Interface, gvr schema.GroupVersionResource) (*unstructuredList, error) {
	ul, err := dynClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	items := make([]metav1.Object, 0, len(ul.Items))
	for i := range ul.Items {
		items = append(items, &ul.Items[i])
	}
	return &unstructuredList{Items: items}, nil
} 