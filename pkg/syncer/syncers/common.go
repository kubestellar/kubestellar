/*
Copyright 2022 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package syncers

import (
	"errors"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	. "github.com/kubestellar/kubestellar/pkg/syncer/clientfactory"
)

func resourceToString(resource edgev2alpha1.EdgeSyncConfigResource) string {
	return fmt.Sprintf("%s.%s/%s in %s", resource.Kind, resource.Group, resource.Name, resource.Namespace)
}

func initializeClients(
	logger klog.Logger,
	syncResources []edgev2alpha1.EdgeSyncConfigResource,
	upstreamClientFactory ClientFactory,
	downstreamClientFactory ClientFactory,
	upstreamClients map[schema.GroupKind]*Client,
	downstreamClients map[schema.GroupKind]*Client,
	conversions []edgev2alpha1.EdgeSynConversion,
) error {
	logger.V(3).Info("initialize clients")
	for _, syncResource := range syncResources {
		logger.V(3).Info(fmt.Sprintf("  setup ResourceClient for %q", resourceToString(syncResource)))

		syncResourceForUpstream := convertToUpstream(syncResource, conversions)

		groupForUp := syncResourceForUpstream.Group
		kindForUp := syncResourceForUpstream.Kind
		gkForUp := schema.GroupKind{
			Group: groupForUp,
			Kind:  kindForUp,
		}
		_, ok := upstreamClients[gkForUp]
		if ok {
			logger.V(3).Info(fmt.Sprintf("  skip since upstreamClientFactory is already setup for %q", resourceToString(syncResourceForUpstream)))
		} else {
			logger.V(3).Info(fmt.Sprintf("  create upstreamClientFactory for %q", resourceToString(syncResourceForUpstream)))
			upstreamClient, err := upstreamClientFactory.GetResourceClient(groupForUp, kindForUp)
			if err != nil {
				logger.Error(err, fmt.Sprintf("failed to create kcpResourceClient '%s.%s'", groupForUp, kindForUp))
				return err
			}
			upstreamClients[gkForUp] = &upstreamClient
		}

		syncResourceForDownstream := convertToDownstream(syncResource, conversions)
		groupForDown := syncResourceForDownstream.Group
		kindForDown := syncResourceForDownstream.Kind
		gkForDown := schema.GroupKind{
			Group: groupForDown,
			Kind:  kindForDown,
		}

		_, ok = downstreamClients[gkForDown]
		if ok {
			logger.V(3).Info(fmt.Sprintf("  skip since downstreamClientFactory is already setup for %q", resourceToString(syncResourceForDownstream)))
		} else {
			logger.V(3).Info(fmt.Sprintf("  create downstreamClientFactory for %q", resourceToString(syncResourceForDownstream)))
			k8sClient, err := downstreamClientFactory.GetResourceClient(groupForDown, kindForDown)
			if err != nil {
				logger.Error(err, fmt.Sprintf("failed to create k8sResourceClient '%s.%s'", groupForDown, kindForDown))
				return err
			}
			downstreamClients[gkForDown] = &k8sClient
		}
	}
	return nil
}

// TODO: Disable dinaturing/re-naturing feature as default. Remove the feature flag once it's fully supported.
func isDenaturingEnabled() bool {
	env, ok := os.LookupEnv("ENABLE_DENATURING")
	return ok && env == "true"
}

func convertToUpstream(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) edgev2alpha1.EdgeSyncConfigResource {
	if !isDenaturingEnabled() {
		return resource
	}
	for _, conversion := range conversions {
		if conversion.Downstream.Group == resource.Group && conversion.Downstream.Kind == resource.Kind && conversion.Downstream.Name == resource.Name {
			resource.Group = conversion.Upstream.Group
			resource.Kind = conversion.Upstream.Kind
			resource.Name = conversion.Upstream.Name
			return resource
		}
	}
	return resource
}

func convertToDownstream(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) edgev2alpha1.EdgeSyncConfigResource {
	if !isDenaturingEnabled() {
		return resource
	}
	for _, conversion := range conversions {
		if conversion.Upstream.Group == resource.Group && conversion.Upstream.Kind == resource.Kind && conversion.Upstream.Name == resource.Name {
			resource.Group = conversion.Downstream.Group
			resource.Kind = conversion.Downstream.Kind
			resource.Name = conversion.Downstream.Name
			return resource
		}
	}
	return resource
}

func applyConversion(source *unstructured.Unstructured, target edgev2alpha1.EdgeSyncConfigResource) {
	if !isDenaturingEnabled() {
		return
	}
	source.SetAPIVersion(target.Group + "/" + target.Version)
	source.SetKind(target.Kind)
	source.SetName(target.Name)
	// CRD restricts the CRD name to be same as group
	if target.Kind == "CustomResourceDefinition" {
		names := strings.Split(target.Name, ".")[1:]
		unstructured.SetNestedField(source.Object, strings.Join(names, "."), "spec", "group")
	}
}

func getClients(resource edgev2alpha1.EdgeSyncConfigResource, upstreamClients map[schema.GroupKind]*Client, downstreamClients map[schema.GroupKind]*Client, conversions []edgev2alpha1.EdgeSynConversion) (*Client, *Client, error) {
	upstreamResource := convertToUpstream(resource, conversions)
	upstreamGk := schema.GroupKind{
		Group: upstreamResource.Group,
		Kind:  upstreamResource.Kind,
	}
	upstreamClient, ok := upstreamClients[upstreamGk]
	if !ok {
		msg := fmt.Sprintf("upstreamClient for '%s.%s' is not registered", upstreamResource.Group, upstreamResource.Kind)
		return nil, nil, errors.New(msg)
	}

	downstreamResource := convertToDownstream(resource, conversions)
	downstreamGk := schema.GroupKind{
		Group: downstreamResource.Group,
		Kind:  downstreamResource.Kind,
	}
	downstreamClient, ok := downstreamClients[downstreamGk]
	if !ok {
		msg := fmt.Sprintf("downstreamClient for '%s.%s' is not registered", downstreamResource.Group, downstreamResource.Kind)
		return nil, nil, errors.New(msg)
	}
	return upstreamClient, downstreamClient, nil
}

// Compute diff between srcResourceList and destResourceList
//
//   - newResources: list of srcResource that exists in srcResourceList but not destResourceList
//   - updatedResources: list of srcResource that exists in both srcResourceList and destResourceList
//   - deletedResources: list of destResource that exists in not srcResourceList but destResourceList
func diff(logger klog.Logger, srcResourceList *unstructured.UnstructuredList, destResourceList *unstructured.UnstructuredList, setAnnotation func(resource *unstructured.Unstructured), hasAnnotation func(resource *unstructured.Unstructured) bool) (
	newResources []unstructured.Unstructured,
	updatedResources []unstructured.Unstructured,
	deletedResources []unstructured.Unstructured,
) {
	for _, srcResource := range srcResourceList.Items {
		destResource, ok := findWithObject(srcResource, destResourceList)
		if ok {
			srcResource.SetResourceVersion(destResource.GetResourceVersion())
			srcResource.SetUID(destResource.GetUID())
			srcResource.SetManagedFields([]metav1.ManagedFieldsEntry{})

			// Avoid to overwrite status field. Though, not sure this workaround is required.
			// Actually, when Syncer donwsynces, Syncer doesn't call UpdateStatus() method. Status fields at downstream side aren't updated by downsyncing.
			setStatusFieldToDestinationStatus(logger, &srcResource, destResource)

			// if hasAnnotation(destResource) {
			setAnnotation(&srcResource)
			updatedResources = append(updatedResources, srcResource)
			// } else {
			// logger.V(2).Info(fmt.Sprintf("  ignore adding %s to updatedResources since annotation is not set.", destResource.GetName()))
			// }
		} else {
			srcResource.SetResourceVersion("")
			srcResource.SetUID("")
			srcResource.SetManagedFields([]metav1.ManagedFieldsEntry{})
			setAnnotation(&srcResource)
			newResources = append(newResources, srcResource)
		}
	}
	for _, destResource := range destResourceList.Items {
		_, ok := findWithObject(destResource, srcResourceList)
		if !ok {
			if hasAnnotation(&destResource) {
				logger.V(3).Info(fmt.Sprintf("  %s is added to deletedResources since annotation is set.", destResource.GetName()))
				deletedResources = append(deletedResources, destResource)
			} else {
				logger.V(2).Info(fmt.Sprintf("  ignore adding %s to deletedResources since annotation is not set.", destResource.GetName()))
			}
		}
	}

	logger.V(3).Info(fmt.Sprintf("  new resource names: %v", mapToNames(newResources)))
	logger.V(3).Info(fmt.Sprintf("  updated resource names: %v", mapToNames(updatedResources)))
	logger.V(3).Info(fmt.Sprintf("  deleted resource names: %v", mapToNames(deletedResources)))

	return newResources, updatedResources, deletedResources
}

func setAnnotation(resource *unstructured.Unstructured, key string, value string) {
	annotations := resource.GetAnnotations()
	if annotations != nil {
		annotations[key] = value
	} else {
		annotations = map[string]string{key: value}
	}
	resource.SetAnnotations(annotations)
}

func getAnnotation(resource *unstructured.Unstructured, key string) string {
	annotations := resource.GetAnnotations()
	return annotations[key]
}

func setStatusFieldToDestinationStatus(logger klog.Logger, srcResource *unstructured.Unstructured, destResource *unstructured.Unstructured) {
	status, found, err := unstructured.NestedMap(destResource.Object, "status")
	if found {
		srcResource.Object["status"] = status
	} else {
		if err != nil {
			logger.V(2).Info(fmt.Sprintf("  %v", err))
		}
		logger.V(2).Info(fmt.Sprintf("  failed to extract status from destination object %q. Nothing to do with status field", destResource.GetName()))
	}
}

func mapToNames(resources []unstructured.Unstructured) []string {
	names := []string{}
	for _, resource := range resources {
		names = append(names, resource.GetName())
	}
	return names
}
