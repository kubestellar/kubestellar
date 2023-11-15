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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	v2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	clientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
)

// Make sure user provided location name is valid
func CheckLocationName(locationName string) error {
	// ensure characters are valid
	matched, _ := regexp.MatchString(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`, locationName)
	if !matched {
		err := errors.New("Location name must match regex '^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'")
		return err
	}
	// check for reserved words
	if locationName == "default" {
		err := errors.New("Location name 'default' may not be used")
		return err
	}
	return nil
}

// Verify that user provided key=value arguments are valid
func CheckLabelArgs(labels []string) error {
	if len(labels) < 1 {
		err := errors.New("At least one label must be provided")
		return err
	}
	// Iterate over labels
	for _, labelString := range labels {
		// Ensure the raw string contains a =
		if !strings.Contains(labelString, "=") {
			err := fmt.Errorf("Invalid label %s, missing '='", labelString)
			return err
		}
		// Split substring on =
		labelSlice := strings.Split(labelString, "=")
		// We should have only a key and value now
		if len(labelSlice) != 2 {
			err := fmt.Errorf("Invalid label %s, must use exactly one '='", labelString)
			return err
		}
		key := labelSlice[0]
		value := labelSlice[1]
		// Make sure the key and value contain only valid characters
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_./-]*$`, key)
		if !matched {
			err := fmt.Errorf("Invalid label %s, key must match regex '^[a-zA-Z0-9][a-zA-Z0-9_./-]*$'", labelString)
			return err
		}
		matched, _ = regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9_.-]{0,61}[a-zA-Z0-9])?$`, value)
		if !matched {
			err := fmt.Errorf("Invalid label %s, value must match regex '^[a-zA-Z0-9]([a-zA-Z0-9_.-]{0,61}[a-zA-Z0-9])?$'", labelString)
			return err
		}
		// Make sure no invalid keys are passed
		if key == "id" {
			err := errors.New("Invalid key, 'id' is handled internally and may not be specified")
			return err
		}
	}
	return nil
}

// Check if SyncTarget exists; if not, create one
func VerifyOrCreateSyncTarget(client *clientset.Clientset, ctx context.Context, imw, locationName string, labels []string) error {
	logger := klog.FromContext(ctx)

	// Get the SyncTarget object
	syncTarget, err := client.EdgeV2alpha1().SyncTargets().Get(ctx, locationName, metav1.GetOptions{})
	if err == nil {
		logger.Info(fmt.Sprintf("Found SyncTarget %s in workspace root:%s", locationName, imw))
		// Check that SyncTarget has an "id" label matching locationName
		err = VerifySyncTargetId(syncTarget, client, ctx, imw, locationName)
		if err != nil {
			return err
		}
		// Check that SyncTarget has user provided key=value pairs, add them if not
		err = VerifySyncTargetLabels(syncTarget, client, ctx, imw, locationName, labels)
		return err
	} else if ! apierrors.IsNotFound(err) {
		// Some error other than a non-existant SyncTarget
		logger.Error(err, fmt.Sprintf("Problem checking for SyncTarget %s in workspace root:%s", locationName, imw))
		return err
	}
	// SyncTarget does not exist, must create
	logger.Info(fmt.Sprintf("No SyncTarget %s in workspace root:%s, creating it", locationName, imw))

	syncTarget = &v2alpha1.SyncTarget {
		TypeMeta: metav1.TypeMeta {
			Kind: "SyncTarget",
			APIVersion: "edge.kubestellar.io/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta {
			Name: locationName,
			Labels: map[string]string{"id": locationName},
		},
	}
	// Add any provided labels
	for _, labelString := range labels {
		// Split raw label string into key and value
		labelSlice := strings.Split(labelString, "=")
		key := labelSlice[0]
		value := labelSlice[1]
		syncTarget.ObjectMeta.Labels[key] = value
	}
	_, err = client.EdgeV2alpha1().SyncTargets().Create(ctx, syncTarget, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to create SyncTarget %s in workspace root:%s", locationName, imw))
		return err
	}

	return nil
}

// Make sure the SyncTarget has an id label matching locationName (update if not)
func VerifySyncTargetId(syncTarget *v2alpha1.SyncTarget, client *clientset.Clientset, ctx context.Context, imw, locationName string) error {
	logger := klog.FromContext(ctx)

	if syncTarget.ObjectMeta.Labels != nil {
		// We're missing a labels field, create it
		id := syncTarget.ObjectMeta.Labels["id"]
		if id == locationName {
			// id matches locationName, all good
			logger.Info(fmt.Sprintf("SyncTarget 'id' label matches %s", locationName))
			return nil
		}
		// ID label does not match locationName, update it
		logger.Info(fmt.Sprintf("SyncTarget %s 'id' label is '%s', changing to '%s'", locationName, id, locationName))
		syncTarget.ObjectMeta.Labels["id"] = locationName
	} else {
		// There are no labels, create it with id: locationName
		logger.Info(fmt.Sprintf("SyncTarget %s is missing labels, adding 'id'", locationName))
		syncTarget.ObjectMeta.Labels = map[string]string{"id": locationName}
	}

	// Apply updates to SyncTarget
	_, err := client.EdgeV2alpha1().SyncTargets().Update(ctx, syncTarget, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to update SyncTarget %s in workspace root:%s", locationName, imw))
		return err
	}

	return nil
}

// Check that SyncTarget has user provided key=value pairs, add them if not
func VerifySyncTargetLabels(syncTarget *v2alpha1.SyncTarget, client *clientset.Clientset, ctx context.Context, imw, locationName string, labels []string) error {
	logger := klog.FromContext(ctx)

	updateSyncTarget := false // bool to see if we need to update SyncTarget
	// Check for labels missing or not matching those provide by user
	for _, labelString := range labels {
		// Split raw label string into key and value
		labelSlice := strings.Split(labelString, "=")
		key := labelSlice[0]
		value := labelSlice[1]
		// Make sure we have a labels field
		if syncTarget.ObjectMeta.Labels == nil {
			// There are no labels, create the label map with first label
			logger.Info("SyncTarget is missing labels, adding it")
			logger.Info(fmt.Sprintf("SyncTarget label %s=, updating value to %s", key, value))
			syncTarget.ObjectMeta.Labels = map[string]string{key: value}
			updateSyncTarget = true
			continue
		}
		valueCurrent := syncTarget.ObjectMeta.Labels[key]
		// Make sure label matches user provided value
		if valueCurrent != value {
			logger.Info(fmt.Sprintf("SyncTarget label %s=%s, updating value to %s", key, valueCurrent, value))
			syncTarget.ObjectMeta.Labels[key] = value
			updateSyncTarget = true
		} else {
			logger.Info(fmt.Sprintf("SyncTarget has label %s=%s", key, value))
		}
	}
	// Update SyncTarget if needed
	if updateSyncTarget {
		// Apply updates to SyncTarget
		_, err := client.EdgeV2alpha1().SyncTargets().Update(ctx, syncTarget, metav1.UpdateOptions{})
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to update SyncTarget %s in workspace root:%s", locationName, imw))
			return err
		}
	}
	return nil
}

// Check if Location exists; if not, create one
func VerifyOrCreateLocation(client *clientset.Clientset, ctx context.Context, imw, locationName string, labels []string) error {
	logger := klog.FromContext(ctx)

	// Get the Location object
	location, err := client.EdgeV2alpha1().Locations().Get(ctx, locationName, metav1.GetOptions{})
	if err == nil {
		logger.Info(fmt.Sprintf("Found Location %s in workspace root:%s", locationName, imw))
		// Check that Location has user provided key=value pairs, add them if not
		err = VerifyLocationLabels(location, client, ctx, imw, locationName, labels)
		return err
	} else if ! apierrors.IsNotFound(err) {
		// Some error other than a non-existant SyncTarget
		logger.Error(err, fmt.Sprintf("Problem checking for Location %s in workspace root:%s", locationName, imw))
		return err
	}
	// Location does not exist, must create
	logger.Info(fmt.Sprintf("No Location %s in workspace root:%s, creating it", locationName, imw))

	location = &v2alpha1.Location {
		TypeMeta: metav1.TypeMeta {
			Kind: "Location",
			APIVersion: "edge.kubestellar.io/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta {
			Name: locationName,
		},
		Spec: v2alpha1.LocationSpec {
			Resource: v2alpha1.GroupVersionResource {
				Group: "edge.kubestellar.io",
				Version: "v2alpha1",
				Resource: "synctargets",
			},
			InstanceSelector: &metav1.LabelSelector {
				MatchLabels: map[string]string{"id": locationName},
			},
		},
	}
	// Add any provided labels
	for _, labelString := range labels {
		// Split raw label string into key and value
		labelSlice := strings.Split(labelString, "=")
		key := labelSlice[0]
		value := labelSlice[1]
		if location.ObjectMeta.Labels != nil {
			// Add key=value
			location.ObjectMeta.Labels[key] = value
		} else {
			// No labels field exists, add the labels map along with this key=value
			location.ObjectMeta.Labels = map[string]string{key: value}
		}
	}
	_, err = client.EdgeV2alpha1().Locations().Create(ctx, location, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to create Location %s in workspace root:%s", locationName, imw))
		return err
	}

	return nil
}

// Check that Location has user provided key=value pairs, add them if not
func VerifyLocationLabels(location *v2alpha1.Location, client *clientset.Clientset, ctx context.Context, imw, locationName string, labels []string) error {
	logger := klog.FromContext(ctx)

	updateLocation := false // bool to see if we need to update Location
	// Check for labels missing or not matching those provide by user
	for _, labelString := range labels {
		// Split raw label string into key and value
		labelSlice := strings.Split(labelString, "=")
		key := labelSlice[0]
		value := labelSlice[1]
		// Make sure we have a labels field
		if location.ObjectMeta.Labels == nil {
			// There are no labels, create the label map with first label
			logger.Info("Location is missing labels, adding it")
			logger.Info(fmt.Sprintf("Location label %s=, updating value to %s", key, value))
			location.ObjectMeta.Labels = map[string]string{key: value}
			updateLocation = true
			continue
		}
		valueCurrent := location.ObjectMeta.Labels[key]
		// Make sure label matches user provided value
		if valueCurrent != value {
			logger.Info(fmt.Sprintf("Location label %s=%s, updating value to %s", key, valueCurrent, value))
			location.ObjectMeta.Labels[key] = value
			updateLocation = true
		} else {
			logger.Info(fmt.Sprintf("Location has label %s=%s", key, value))
		}
	}
	// Update Location if needed
	if updateLocation {
		// Apply updates to Location
		_, err := client.EdgeV2alpha1().Locations().Update(ctx, location, metav1.UpdateOptions{})
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to update Location %s in workspace root:%s", locationName, imw))
			return err
		}
	}
	return nil
}

// Check if default Location exists, delete it if so
func VerifyNoDefaultLocation(client *clientset.Clientset, ctx context.Context, imw string) error {
	logger := klog.FromContext(ctx)

	// Check for "default" Location object
	_, err := client.EdgeV2alpha1().Locations().Get(ctx, "default", metav1.GetOptions{})
	if err != nil {
		// Check if error is due to the lack of a "default" location object (what we want)
		if apierrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Verified no default Location in workspace root:%s", imw))
			return nil
		}
		// There is some error other than trying to get a non-existent object
		logger.Error(err, fmt.Sprintf("Could not check if default Location in workspace root:%s", imw))
		return err
	}

	// "default" Location exists, delete it
	logger.Info(fmt.Sprintf("Found default Location in workspace root:%s, deleting it", imw))
	err = client.EdgeV2alpha1().Locations().Delete(ctx, "default", metav1.DeleteOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to delete default Location in workspace root:%s", imw))
		return err
	}
	return nil
}
