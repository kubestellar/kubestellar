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
	"errors"
	"fmt"
	"regexp"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
)

// Make sure user provided WDS name is valid
func CheckWdsName(wdsName string) error {
	// ensure characters are valid
	matched, _ := regexp.MatchString(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`, wdsName)
	if !matched {
		err := errors.New("WDS name must match regex '^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'")
		return err
	}
	return nil
}

// Check if WDS workspace exists, and if not create it
func VerifyOrCreateWDS(client *kcpclientset.Clientset, ctx context.Context, logger klog.Logger, wdsName string) error {
	// Check if WDS workspace exists
	_, err := client.TenancyV1alpha1().Workspaces().Get(ctx, wdsName, metav1.GetOptions{})
	if err == nil {
		logger.Info(fmt.Sprintf("Found WDS workspace root:%s", wdsName))
		return err
	}
	if ! apierrors.IsNotFound(err) {
		// Some error other than a non-existant workspace
		logger.Error(err, fmt.Sprintf("Error checking for root:WDS %s", wdsName))
		return err
	}

	// WDS workspace does not exist, create it
	logger.Info(fmt.Sprintf("No WDS workspace root:%s, creating it", wdsName))

	workspace := &tenancyv1alpha1.Workspace {
		TypeMeta: metav1.TypeMeta {
			Kind: "Workspace",
			APIVersion: "tenancy.kcp.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta {
			Name: wdsName,
		},
	}
	_, err = client.TenancyV1alpha1().Workspaces().Create(ctx, workspace, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to create WDS workspace root:%s", wdsName))
		return err
	}

	// Wait for workspace to become ready
	wait.Poll(time.Millisecond*100, time.Second*timeoutTime, func() (bool, error) {
		// See if we can get new workspace
		if _, err := client.TenancyV1alpha1().Workspaces().Get(ctx, wdsName, metav1.GetOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				// Failed to get due to not found, try until timeout
				return false, nil
			}
			// Some error happened beyond not finding the workspace
			return false, err
		}
		// We got the workspace, we're good to go
		logger.Info(fmt.Sprintf("Workspace root:%s ready", wdsName))
		// NOTE sometimes the next step of creating an APIBinding still fails,
		// so add more delay. If this is still not long enough, an error like
		// the following will be seen when creating an APIBinding:
		//
		// apibindings.apis.kcp.io "bind-flowcontrol.apiserver.k8s.io" is
		// forbidden: User "kcp-admin" cannot get resource "apibindings" in API
		// group "apis.kcp.io" at the cluster scope: access denied
		time.Sleep(time.Millisecond*1000)
		return true, nil
	})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Problem waiting for WDS workspace root:%s", wdsName))
		return err
	}

	return nil
}

// Check for Kube APIBindings
// If withKube is true, create any bindings that don't exist
// If withKube is false, delete any bindings that exist
func VerifyKubeAPIBindings(client *kcpclientset.Clientset, ctx context.Context, logger klog.Logger, withKube bool) error {
	// APIBindings to check
	binds := []string {
		"kubernetes",
		"apiregistration.k8s.io",
		"apps",
		"autoscaling",
		"batch",
		"core.k8s.io",
		"cluster-core.k8s.io",
		"discovery.k8s.io",
		"flowcontrol.apiserver.k8s.io",
		"networking.k8s.io",
		"cluster-networking.k8s.io",
		"node.k8s.io",
		"policy",
		"scheduling.k8s.io",
		"storage.k8s.io",
		"cluster-storage.k8s.io",
	}
	// Iterate over bindings
	for _, exportName := range binds {
		bindName := "bind-" + exportName
		if withKube {
			// Make sure these bindings exist
			err := VerifyOrCreateAPIBinding(client, ctx, logger, bindName, exportName, "root:compute")
			if err != nil {
				return err
			}
		} else {
			// Remove these bindings if they exist
			removed, err := DeleteAPIBinding(client, ctx, bindName)
			if err != nil {
				logger.Error(err, fmt.Sprintf("Problem removing APIBinding %s", bindName))
				return err
			}
			if removed {
				logger.Info(fmt.Sprintf("Removed APIBinding %s", bindName))
			} else {
				logger.Info(fmt.Sprintf("Verified no APIBinding %s", bindName))
			}
		}
	}
	return nil
}

// Delete an API binding.
// Don't return an error if it doesn't exist, instead return a boolean on
// whether a deletion had to take place.
func DeleteAPIBinding(client *kcpclientset.Clientset, ctx context.Context, bindName string) (bool, error) {
	// Delete the APIBinding
	err := client.ApisV1alpha1().APIBindings().Delete(ctx, bindName, metav1.DeleteOptions{})
	if err == nil {
		// Removed the APIBinding
		return true, nil
	} else if ! apierrors.IsNotFound(err) {
		// Some error other than a non-existant APIBinding
		return false, err
	}
	// No APIBinding to remove
	return false, nil
}