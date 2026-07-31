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

	wrapees := []transport.Wrapee{wrapee1, wrapee2, wrapee3}

	kindToResource := func(gk schema.GroupKind) string {
		return gk.Kind + "s" // Simplistic mock
	}

	result := ocmTransport.WrapObjects(wrapees, kindToResource)

	manifestWork, ok := result.(*workv1.ManifestWork)
	if !ok {
		t.Fatalf("expected *workv1.ManifestWork, got %T", result)
	}

	if len(manifestWork.Spec.ManifestConfigs) != 2 {
		t.Fatalf("expected 2 manifest configs, got %d", len(manifestWork.Spec.ManifestConfigs))
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

	// Second config should be for CreateOnly deployment (wrapee3)
	deployConfig := manifestWork.Spec.ManifestConfigs[1]
	if deployConfig.ResourceIdentifier.Resource != "Deployments" {
		t.Errorf("expected Deployments resource, got %s", deployConfig.ResourceIdentifier.Resource)
	}
	if deployConfig.UpdateStrategy == nil || deployConfig.UpdateStrategy.Type != workv1.UpdateStrategyTypeCreateOnly {
		t.Errorf("expected CreateOnly update strategy for Deployment, got %v", deployConfig.UpdateStrategy)
	}
}
