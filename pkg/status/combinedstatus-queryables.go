/*
Copyright 2024 The KubeStellar Authors.

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

package status

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtime2 "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/pkg/util"
)

// getCombinedContentMap returns a map of content for the given workstatus.
func getCombinedContentMap(listersConcurrentMap util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister],
	workStatus *workStatus, resolution *combinedStatusResolution) map[string]interface{} {

	// betting on `combinedStatusResolution::queryingContentRequirements` being faster
	// than fetching content that is not required.
	sourceObjectRequired, returnedRequired,
		inventoryRequired, propagationMetaRequired := resolution.queryingContentRequirements()

	content := map[string]interface{}{}

	if sourceObjectRequired {
		objMap, err := getObjectMetaAndSpec(listersConcurrentMap, workStatus.SourceObjectIdentifier)
		if err != nil {
			runtime2.HandleError(fmt.Errorf("failed to get meta & spec for source object %s: %w",
				workStatus.SourceObjectIdentifier, err))
		}

		content[sourceObjectKey] = objMap
	}

	if returnedRequired {
		content[returnedKey] = workStatus.Content()
	}

	if inventoryRequired {
		content[inventoryKey] = inventoryForWorkStatus(workStatus)
	}

	if propagationMetaRequired {
		content[propagationMetaKey] = propagateMetaForWorkStatus(workStatus, resolution)
	}

	return content
}

// getObjectMetaAndSpec fetches the metadata and spec of the object associated
// with the given workload object identifier.
// The function is guaranteed not to return a key of `status` in the map.
// If the resource contains any other subresources, they are fetched as well.
func getObjectMetaAndSpec(listersConcurrentMap util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister],
	objectIdentifier util.ObjectIdentifier) (map[string]interface{}, error) {
	// fetch object
	lister, exists := listersConcurrentMap.Get(objectIdentifier.GVR())
	if !exists {
		return nil, fmt.Errorf("lister not found for gvr %s", objectIdentifier.GVR())
	}

	obj, err := getObject(lister, objectIdentifier.ObjectName.Namespace, objectIdentifier.ObjectName.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get object (%v) with gvr (%v): %w", objectIdentifier.ObjectName,
			objectIdentifier.GVR(), err)
	}

	unstrObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("object cannot be cast to *unstructured.Unstructured: object: %s",
			util.RefToRuntimeObj(obj))
	}

	return unstrObj.Object, nil
}

func getObject(lister cache.GenericLister, namespace, name string) (runtime.Object, error) {
	if namespace != "" {
		return lister.ByNamespace(namespace).Get(name)
	}
	return lister.Get(name)
}

// inventoryForWorkStatus returns an inventory map for the given workstatus.
func inventoryForWorkStatus(ws *workStatus) map[string]interface{} {
	return map[string]interface{}{
		"name": ws.WECName,
	}
}

func propagateMetaForWorkStatus(ws *workStatus, resolution *combinedStatusResolution) map[string]interface{} {
	var protoLastUpdateTimestamp *timestamppb.Timestamp

	if ws.lastUpdateTime != nil {
		protoLastUpdateTimestamp = timestamppb.New(ws.lastUpdateTime.Time)
	}

	return map[string]interface{}{
		"lastReturnedUpdateTimestamp": protoLastUpdateTimestamp,
	}
}
