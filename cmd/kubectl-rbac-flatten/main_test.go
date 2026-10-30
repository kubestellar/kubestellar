package main

import (
	"context"
	"testing"

	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclientset "k8s.io/client-go/kubernetes/fake"

	"github.com/kubestellar/kubestellar/pkg/util"
)

func TestGetFlatNonResourceURLs(t *testing.T) {
	client := fakeclientset.NewSimpleClientset(&rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "system:discovery"},
		Rules: []rbac.PolicyRule{
			{
				Verbs:           []string{"get"},
				NonResourceURLs: []string{"/api", "/api/*"},
			},
		},
	}, &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "system:discovery"},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:discovery",
		},
		Subjects: []rbac.Subject{
			{
				Kind:     "Group",
				Name:     "system:authenticated",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
	})

	rscMap := make(resourceMap)

	subjFilter := subjectFilter{
		UserName:       util.StringFilter{AllPass: true},
		UserGroup:      util.StringFilter{AllPass: true},
		ServiceAccount: util.StringFilter{AllPass: true},
	}
	verbFilter := util.StringFilter{AllPass: true}
	rscFilter := util.StringFilter{AllPass: true}

	flat, errs := getFlat(context.Background(), client.RbacV1(), rscMap, subjFilter, verbFilter, rscFilter)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(flat) != 1 {
		t.Fatalf("expected 1 flat tuple, got %d", len(flat))
	}

	tuple := flat[0]
	if len(tuple.Rule.NonResourcePaths) != 2 || tuple.Rule.NonResourcePaths[0] != "/api" {
		t.Errorf("unexpected non-resource paths: %v", tuple.Rule.NonResourcePaths)
	}
	if len(tuple.Rule.Resources) > 0 {
		t.Errorf("expected 0 resources for non-resource rule, got %d", len(tuple.Rule.Resources))
	}
}
