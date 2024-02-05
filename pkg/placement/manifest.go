/*
Copyright 2023 The KubeStellar Authors.

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

package placement

import (
	"context"

	"github.com/go-logr/logr"
	workv1 "open-cluster-management.io/api/work/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubestellar/kubestellar/pkg/ocm"
)

// TODO (maroon): this file should be deleted when transport is ready
func deleteObjectOnManagedClusters(logger logr.Logger, cl client.Client, obj runtime.Object,
	managedClusters sets.Set[string]) {
	for managedCluster := range managedClusters {
		err := deleteManifestForObject(cl, obj, managedCluster)
		if err != nil {
			logger.Error(err, "Error deleting object on mailbox")
		}
	}
}

func reconcileManifest(cl client.Client, manifest *workv1.ManifestWork, namespace string) error {
	manifest.SetNamespace(namespace)

	// use the deep copy for applying the manifest as the manifest object gets updated by the first cluster.
	applyManifest := manifest.DeepCopy()

	// use server-side apply
	if err := cl.Patch(context.TODO(), applyManifest, client.Apply, client.FieldOwner("kubestellar")); err != nil {
		return err
	}

	return nil
}

func deleteManifestForObject(cl client.Client, obj runtime.Object, namespace string) error {
	manifest := ocm.BuildEmptyManifestFromObject(obj)
	manifest.SetNamespace(namespace)
	if err := cl.Delete(context.TODO(), manifest, &client.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
