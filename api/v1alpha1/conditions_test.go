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
	conditions1 := []PlacementCondition{
		generateCondition("ConditionTypeA", "ReasonA", "MessageA",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionTrue, metav1.Now(), addTime(2)),
	}

	newCondition := generateCondition("ConditionTypeA", "ReasonAUpdated", "MessageAUpdated",
		corev1.ConditionTrue, addTime(4), addTime(5))

	expectedConditions := []PlacementCondition{
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
	c1 := []PlacementCondition{
		generateCondition("ConditionTypeA", "ReasonA", "MessageA",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionFalse, metav1.Now(), addTime(2)),
	}
	c2 := []PlacementCondition{
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionFalse, metav1.Now(), addTime(3)),
		generateCondition("ConditionTypeA", "ReasonA", "MessageA",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
	}

	if !AreConditionSlicesSame(c1, c2) {
		t.Errorf("AreConditionSlicesSame failed: expected true, but got false")
	}

	// Create two slices of conditions with different elements
	c3 := []PlacementCondition{
		generateCondition("ConditionTypeA", "ReasonA", "MessageA",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
		generateCondition("ConditionTypeB", "ReasonB", "MessageB",
			corev1.ConditionTrue, metav1.Now(), addTime(2)),
	}
	c4 := []PlacementCondition{
		generateCondition("ConditionTypeC", "ReasonC", "MessageC",
			corev1.ConditionFalse, metav1.Now(), metav1.Now()),
	}

	if AreConditionSlicesSame(c3, c4) {
		t.Errorf("AreConditionSlicesSame failed: expected false, but got true")
	}
}

func generateCondition(ctype ConditionType, reason ConditionReason, message string, status corev1.ConditionStatus, ltt, ltu metav1.Time) PlacementCondition {
	return PlacementCondition{
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
