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

package v1alpha1

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAreConditionsEqual(t *testing.T) {
	// Create two conditions with only the LastUpdateTime field different
	c1 := generateCondition("ConditionType", "Reason", "Message", corev1.ConditionTrue, metav1.Now(), metav1.Now())
	c2 := generateCondition("ConditionType", "Reason", "Message", corev1.ConditionTrue, metav1.Now(),
		addTime(1))

	if !AreConditionsEqual(c1, c2) {
		t.Errorf("AreConditionsEqual failed: expected true, but got false")
	}

	// Create two conditions with all fields different
	c3 := generateCondition("ConditionTypeA", "ReasonA", "MessageA", corev1.ConditionTrue, metav1.Now(), addTime(1))
	c4 := generateCondition("ConditionTypeB", "ReasonB", "MessageB", corev1.ConditionTrue, metav1.Now(), addTime(2))
	if AreConditionsEqual(c3, c4) {
		t.Errorf("AreConditionsEqual failed: expected false, but got true")
	}
}

func TestSetCondition(t *testing.T) {
	// Create a slice of conditions and set a new condition
	conditions1 := []BindingPolicyCondition{
		generateCondition("ConditionTypeA", "ReasonA", "MessageA",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionTrue, metav1.Now(), addTime(2)),
	}

	newCondition := generateCondition("ConditionTypeA", "ReasonAUpdated", "MessageAUpdated",
		corev1.ConditionTrue, addTime(4), addTime(5))

	expectedConditions := []BindingPolicyCondition{
		newCondition,
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionTrue, metav1.Now(), addTime(2)),
	}

	actualConditions := SetCondition(conditions1, newCondition)

	if !AreConditionSlicesSame(actualConditions, expectedConditions) {
		t.Errorf("SetCondition failed: expected %+v, but got %+v", expectedConditions, actualConditions)
	}
}

func TestAreConditionSlicesSame(t *testing.T) {
	// Create two slices of conditions with the same elements in different orders
	c1 := []BindingPolicyCondition{
		generateCondition("ConditionTypeA", "ReasonA", "MessageA",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionFalse, metav1.Now(), addTime(2)),
	}
	c2 := []BindingPolicyCondition{
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionFalse, metav1.Now(), addTime(3)),
		generateCondition("ConditionTypeA", "ReasonA", "MessageA",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
	}

	if !AreConditionSlicesSame(c1, c2) {
		t.Errorf("AreConditionSlicesSame failed: expected true, but got false")
	}

	// Create two slices of conditions with different elements
	c3 := []BindingPolicyCondition{
		generateCondition("ConditionTypeA", "ReasonA", "MessageA",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionTrue, metav1.Now(), addTime(2)),
	}
	c4 := []BindingPolicyCondition{
		generateCondition("ConditionTypeC", "ReasonC", "MessageC",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
	}

	if AreConditionSlicesSame(c3, c4) {
		t.Errorf("AreConditionSlicesSame failed: expected false, but got true")
	}
}

func generateCondition(ctype ConditionType, reason ConditionReason, message string, status corev1.ConditionStatus, ltt, ltu metav1.Time) BindingPolicyCondition {
	return BindingPolicyCondition{
		Type:               ctype,
		Status:             status,
		LastTransitionTime: ltt,
		LastUpdateTime:     ltu,
		Reason:             reason,
		Message:            message,
	}
}

func addTime(t time.Duration) metav1.Time {
	return metav1.NewTime(time.Now().Add(2 * time.Hour))
}
