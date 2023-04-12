/*
Copyright 2022 The KCP Authors.

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
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	. "github.com/kcp-dev/edge-mc/pkg/syncer/clientfactory"
)

func resourceToString(resource edgev1alpha1.EdgeSyncConfigResource) string {
	return fmt.Sprintf("%s.%s/%s in %s", resource.Kind, resource.Group, resource.Name, resource.Namespace)
}

func initializeClients(
	logger klog.Logger,
	syncResources []edgev1alpha1.EdgeSyncConfigResource,
	upstreamClientFactory ClientFactory,
	downstreamClientFactory ClientFactory,
	upstreamClients map[schema.GroupKind]*Client,
	downstreamClients map[schema.GroupKind]*Client,
	conversions []edgev1alpha1.EdgeSynConversion,
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

func convertToUpstream(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) edgev1alpha1.EdgeSyncConfigResource {
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

func convertToDownstream(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) edgev1alpha1.EdgeSyncConfigResource {
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

func applyConversion(source *unstructured.Unstructured, target edgev1alpha1.EdgeSyncConfigResource) {
	source.SetAPIVersion(target.Group + "/" + target.Version)
	source.SetKind(target.Kind)
	source.SetName(target.Name)
	// CRD restricts the CRD name to be same as group
	if target.Kind == "CustomResourceDefinition" {
		names := strings.Split(target.Name, ".")[1:]
		unstructured.SetNestedField(source.Object, strings.Join(names, "."), "spec", "group")
	}
}

func getClients(resource edgev1alpha1.EdgeSyncConfigResource, upstreamClients map[schema.GroupKind]*Client, downstreamClients map[schema.GroupKind]*Client, conversions []edgev1alpha1.EdgeSynConversion) (*Client, *Client, error) {
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
