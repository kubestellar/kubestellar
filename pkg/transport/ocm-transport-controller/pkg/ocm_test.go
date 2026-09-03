package ocm

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	workv1 "open-cluster-management.io/api/work/v1"

	"github.com/kubestellar/kubestellar/pkg/transport"
)

func TestWrapObjects(t *testing.T) {
	ocmTransport := NewOCMTransport()

	// 1. Create a non-Job object that is NOT createOnly
	configMap := &unstructured.Unstructured{}
	configMap.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})
	configMap.SetName("test-cm")
	configMap.SetNamespace("default")
	wrapee1 := transport.NewWrapee(configMap, false)

	// 2. Create a Job object
	job := &unstructured.Unstructured{}
	job.SetGroupVersionKind(schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"})
	job.SetName("test-job")
	job.SetNamespace("default")
	wrapee2 := transport.NewWrapee(job, false)

	// 3. Create a CreateOnly object
	createOnlyObj := &unstructured.Unstructured{}
	createOnlyObj.SetGroupVersionKind(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"})
	createOnlyObj.SetName("test-deploy")
	createOnlyObj.SetNamespace("default")
	wrapee3 := transport.NewWrapee(createOnlyObj, true)

	// 4. Create an object with the ignore-fields annotation
	annotatedObj := &unstructured.Unstructured{}
	annotatedObj.SetGroupVersionKind(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"})
	annotatedObj.SetName("test-annotated")
	annotatedObj.SetNamespace("default")
	annotatedObj.SetAnnotations(map[string]string{
		"kubestellar.io/ignore-fields": ".spec.replicas, .spec.template.metadata.labels",
	})
	wrapee4 := transport.NewWrapee(annotatedObj, false)

	wrapees := []transport.Wrapee{wrapee1, wrapee2, wrapee3, wrapee4}

	kindToResource := func(gk schema.GroupKind) string {
		return gk.Kind + "s" // Simplistic mock
	}

	result := ocmTransport.WrapObjects(wrapees, kindToResource)

	manifestWork, ok := result.(*workv1.ManifestWork)
	if !ok {
		t.Fatalf("expected *workv1.ManifestWork, got %T", result)
	}

	if len(manifestWork.Spec.ManifestConfigs) != 3 {
		t.Fatalf("expected 3 manifest configs, got %d", len(manifestWork.Spec.ManifestConfigs))
	}

	// First config should be for Job (wrapee2)
	jobConfig := manifestWork.Spec.ManifestConfigs[0]
	if jobConfig.ResourceIdentifier.Resource != "Jobs" {
		t.Errorf("expected Jobs resource, got %s", jobConfig.ResourceIdentifier.Resource)
	}
	if jobConfig.UpdateStrategy == nil || jobConfig.UpdateStrategy.Type != workv1.UpdateStrategyTypeServerSideApply {
		t.Errorf("expected ServerSideApply update strategy for Job, got %v", jobConfig.UpdateStrategy)
	}
	if len(jobConfig.UpdateStrategy.ServerSideApply.IgnoreFields) == 0 || len(jobConfig.UpdateStrategy.ServerSideApply.IgnoreFields[0].JSONPaths) != 3 {
		t.Errorf("expected 3 IgnoreFields for Job")
	}
	if len(jobConfig.ConditionRules) != 1 {
		t.Fatalf("expected 1 ConditionRule for Job, got %d", len(jobConfig.ConditionRules))
	}
	if jobConfig.ConditionRules[0].Condition != "Complete" || jobConfig.ConditionRules[0].Type != workv1.WellKnownConditionsType {
		t.Errorf("expected WellKnownConditions Complete rule for Job, got %v", jobConfig.ConditionRules[0])
	}

	// Second config should be for CreateOnly deployment (wrapee3)
	deployConfig := manifestWork.Spec.ManifestConfigs[1]
	if deployConfig.ResourceIdentifier.Resource != "Deployments" {
		t.Errorf("expected Deployments resource, got %s", deployConfig.ResourceIdentifier.Resource)
	}
	if deployConfig.UpdateStrategy == nil || deployConfig.UpdateStrategy.Type != workv1.UpdateStrategyTypeCreateOnly {
		t.Errorf("expected CreateOnly update strategy for Deployment, got %v", deployConfig.UpdateStrategy)
	}

	// Third config should be for annotated object (wrapee4)
	annotatedConfig := manifestWork.Spec.ManifestConfigs[2]
	if annotatedConfig.ResourceIdentifier.Resource != "Deployments" {
		t.Errorf("expected Deployments resource, got %s", annotatedConfig.ResourceIdentifier.Resource)
	}
	if annotatedConfig.UpdateStrategy == nil || annotatedConfig.UpdateStrategy.Type != workv1.UpdateStrategyTypeServerSideApply {
		t.Errorf("expected ServerSideApply update strategy for Annotated Obj, got %v", annotatedConfig.UpdateStrategy)
	}
	if len(annotatedConfig.UpdateStrategy.ServerSideApply.IgnoreFields) == 0 {
		t.Fatalf("expected IgnoreFields for Annotated Obj")
	}
	if len(annotatedConfig.UpdateStrategy.ServerSideApply.IgnoreFields[0].JSONPaths) != 2 {
		t.Errorf("expected 2 IgnoreFields for Annotated Obj, got %v", annotatedConfig.UpdateStrategy.ServerSideApply.IgnoreFields[0].JSONPaths)
	}
	if annotatedConfig.UpdateStrategy.ServerSideApply.IgnoreFields[0].JSONPaths[0] != ".spec.replicas" || annotatedConfig.UpdateStrategy.ServerSideApply.IgnoreFields[0].JSONPaths[1] != ".spec.template.metadata.labels" {
		t.Errorf("unexpected IgnoreFields parsed: %v", annotatedConfig.UpdateStrategy.ServerSideApply.IgnoreFields[0].JSONPaths)
	}
}
