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

package plugin

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"

	clientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
)

// Delete a Location.
// Don't return an error if it doesn't exist, instead return a boolean on
// whether a deletion had to take place.
func DeleteLocation(client *clientset.Clientset, ctx context.Context, locationName string) (bool, error) {
	// Delete the Location
	err := client.EdgeV2alpha1().Locations().Delete(ctx, locationName, metav1.DeleteOptions{})
	if err == nil {
		// Removed Location
		return true, nil
	} else if ! apierrors.IsNotFound(err) {
		// Some error other than a non-existant Location
		return false, err
	}
	// No Location to remove
	return false, nil
}

// Delete a SyncTarget.
// Don't return an error if it doesn't exist, instead return a boolean on
// whether a deletion had to take place.
func DeleteSyncTarget(client *clientset.Clientset, ctx context.Context, syncTargetName string) (bool, error) {
	// Delete the SyncTarget
	err := client.EdgeV2alpha1().SyncTargets().Delete(ctx, syncTargetName, metav1.DeleteOptions{})
	if err == nil {
		// Removed SyncTarget
		return true, nil
	} else if ! apierrors.IsNotFound(err) {
		// Some error other than a non-existant SyncTarget
		return false, err
	}
	// No SyncTarget to remove
	return false, nil
}

// Delete a KCP workspace.
// Don't return an error if it doesn't exist, instead return a boolean on
// whether a deletion had to take place.
func DeleteWorkspace(client *kcpclientset.Clientset, ctx context.Context, wsName string) (bool, error) {
	// Delete the workspace
	err := client.TenancyV1alpha1().Workspaces().Delete(ctx, wsName, metav1.DeleteOptions{})
	if err == nil {
		// Removed workspace
		return true, nil
	} else if ! apierrors.IsNotFound(err) {
		// Some error other than a non-existant workspace
		return false, err
	}
	// No workspace to remove
	return false, nil
}
