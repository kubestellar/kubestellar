/*
Copyright 2025 The KubeStellar Authors.

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

package main

// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
// to ensure that exec-entrypoint and run can make use of them.

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	auth_client "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/abstract"
)

//const noNamespace = "**"

type NamespacedName = apimachtypes.NamespacedName

func main() {
	klog.InitFlags(flag.CommandLine)
	fs := pflag.NewFlagSet("kubectl-rbac-flatten", pflag.ExitOnError)
	fs.AddGoFlagSet(flag.CommandLine)
	cliOpts := genericclioptions.NewConfigFlags(true)
	cliOpts.AddFlags(fs)
	outputFormat := "table"
	fs.StringVarP(&outputFormat, "output-format", "o", outputFormat, "output format, either json or table")
	apiGroupList := []string{"*"}
	fs.StringSliceVar(&apiGroupList, "api-groups", apiGroupList, "comma-separated list of API groups to include; '*' means all")
	resourceList := []string{"*"}
	fs.StringSliceVar(&resourceList, "resources", resourceList, "comma-separated list of resources to include; '*' means all")
	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.FromContext(ctx)
	ctx = klog.NewContext(ctx, logger)

	apiGroups := sets.New(apiGroupList...)
	if apiGroups.Len() == 0 || apiGroups.Has("*") {
		apiGroups = nil
	}

	resources := sets.New(resourceList...)
	if resources.Len() == 0 || resources.Has("*") {
		resources = nil
	}

	config, err := cliOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to build config from flags")
		os.Exit(5)
	}

	client := auth_client.NewForConfigOrDie(config)
	flat, errs := getFlat(ctx, client, apiGroups, resources)
	switch outputFormat {
	case "table":
		showAG := len(apiGroups) != 1
		showRsc := len(resources) != 1
		tw := printers.GetNewTabWriter(os.Stdout)
		tw.Write([]byte("BINDING\t"))
		tw.Write([]byte("SUBJECT\t"))
		tw.Write([]byte("VERB\t"))
		if showAG {
			tw.Write([]byte("APIGROUP\t"))
		}
		if showRsc {
			tw.Write([]byte("RESOURCE\t"))
		}
		tw.Write([]byte("OBJNAME\n"))
		for _, tup := range flat {
			for _, verb := range tup.Rule.Verbs {
				for _, apiGroup := range tup.Rule.APIGroups {
					for _, rsc := range tup.Rule.Resources {
						objNames := tup.Rule.ResourceNames
						if len(objNames) == 0 {
							objNames = []string{"*"}
						}
						for _, objName := range objNames {
							tw.Write([]byte(tup.Binding.String() + "\t"))
							tw.Write([]byte(fmtSubj(tup.Subject) + "\t"))
							tw.Write([]byte(verb + "\t"))
							if showAG {
								tw.Write([]byte(apiGroup + "\t"))
							}
							if showRsc {
								tw.Write([]byte(rsc + "\t"))
							}
							tw.Write([]byte(objName + "\n"))
						}
					}
				}
			}
		}
		err := tw.Flush()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		if apiGroups == nil {
			fmt.Println()
			tw = printers.GetNewTabWriter(os.Stdout)
			tw.Write([]byte("BINDING\t"))
			tw.Write([]byte("ROLE\t"))
			tw.Write([]byte("SUBJECT\t"))
			tw.Write([]byte("VERB\t"))
			tw.Write([]byte("NRURL\n"))
			for _, tup := range flat {
				for _, verb := range tup.Rule.Verbs {
					for _, nrURL := range tup.Rule.NonResourceURLs {
						tw.Write([]byte(tup.Binding.String() + "\t"))
						tw.Write([]byte(tup.RoleName + "\t"))
						tw.Write([]byte(fmtSubj(tup.Subject) + "\t"))
						tw.Write([]byte(verb + "\t"))
						tw.Write([]byte(nrURL + "\n"))
					}
				}
			}
			err = tw.Flush()
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}
	case "json":
		fmt.Println("[")
		first := true
		for _, tup := range flat {
			if first {
				first = false
			} else {
				fmt.Println(",")
			}
			enc, err := json.Marshal(tup)
			if err != nil {
				logger.Error(err, "Failed to encode product as JSON")
			}
			fmt.Println(string(enc))
		}
		fmt.Println("]")
	}
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

var allAPIGroups = sets.New[string]("*")

type Flat []Tuple

type Tuple struct {
	Binding  NamespacedName
	RoleName string
	Subject  rbac.Subject
	Rule     rbac.PolicyRule
}

func getFlat(ctx context.Context, client auth_client.RbacV1Interface, apiGroups, resources sets.Set[string]) (Flat, []error) {
	var errs []error
	var crMap map[string]rbac.ClusterRole
	var crbList []rbac.ClusterRoleBinding
	var rMap map[NamespacedName]rbac.Role
	var rbList []rbac.RoleBinding
	if list, err := client.ClusterRoles().List(ctx, metav1.ListOptions{}); err != nil {
		errs = append(errs, err)
	} else {
		crMap = abstract.SliceToPrimitiveMap(list.Items, func(x rbac.ClusterRole) string { return x.Name }, Id)
	}
	if list, err := client.ClusterRoleBindings().List(ctx, metav1.ListOptions{}); err != nil {
		errs = append(errs, err)
	} else {
		crbList = list.Items
	}
	if list, err := client.Roles(metav1.NamespaceAll).List(ctx, metav1.ListOptions{}); err != nil {
		errs = append(errs, err)
	} else {
		rMap = abstract.SliceToPrimitiveMap(list.Items, func(x rbac.Role) NamespacedName { return NamespacedName{Namespace: x.Namespace, Name: x.Name} }, Id)
	}
	if list, err := client.RoleBindings(metav1.NamespaceAll).List(ctx, metav1.ListOptions{}); err != nil {
		errs = append(errs, err)
	} else {
		rbList = list.Items
	}
	ans := Flat{}
	filterGroups := abstract.SliceFilter(isStarOrInSet(apiGroups), true)
	filterResources := abstract.SliceFilter(isStarOrInSet(resources), true)
	addProducts := func(rules []rbac.PolicyRule, subjects []rbac.Subject, source NamespacedName, roleName string) {
		for _, rule := range rules {
			if apiGroups != nil && len(rule.APIGroups) != 0 {
				rule.APIGroups = filterGroups(rule.APIGroups)
				if len(rule.APIGroups) == 0 {
					continue
				}
			}
			if resources != nil && len(rule.Resources) != 0 {
				rule.Resources = filterResources(rule.Resources)
				if len(rule.Resources) == 0 {
					continue
				}
			}
			for _, subj := range subjects {
				ans = append(ans, Tuple{Binding: source, RoleName: roleName, Subject: subj, Rule: rule})
			}
		}
	}
	for _, crb := range crbList {
		br := NamespacedName{Name: crb.Name}
		var rules []rbac.PolicyRule
		switch crb.RoleRef.Kind {
		case "ClusterRole":
			if cr, ok := crMap[crb.RoleRef.Name]; ok {
				rules = cr.Rules
			} else {
				errs = append(errs, fmt.Errorf("ClusterRoleBinding %s references unknown ClusterRole: %q", crb.Name, crb.RoleRef.Name))
			}
		default:
			errs = append(errs, fmt.Errorf("ClusterRoleBinding %s references unknown Kind of Role: %q", crb.Name, crb.RoleRef.Kind))
		}
		addProducts(rules, crb.Subjects, br, crb.RoleRef.Name)
	}
	for _, rb := range rbList {
		br := NamespacedName{Namespace: rb.Namespace, Name: rb.Name}
		var rules []rbac.PolicyRule
		switch rb.RoleRef.Kind {
		case "ClusterRole":
			if cr, ok := crMap[rb.RoleRef.Name]; ok {
				rules = cr.Rules
			} else {
				errs = append(errs, fmt.Errorf("RoleBinding %s/%s references unknown ClusterRole: %q", rb.Namespace, rb.Name, rb.RoleRef.Name))
			}
		case "Role":
			nn := NamespacedName{Namespace: rb.Namespace, Name: rb.RoleRef.Name}
			if r, ok := rMap[nn]; ok {
				rules = r.Rules
			} else {
				errs = append(errs, fmt.Errorf("RoleBinding %s/%s references unknown Role: %q", rb.Namespace, rb.Name, rb.RoleRef.Name))
			}
		default:
			errs = append(errs, fmt.Errorf("RoleBinding %s/%s references unknown Kind of Role: %q", rb.Namespace, rb.Name, rb.RoleRef.Kind))
		}
		addProducts(rules, rb.Subjects, br, rb.RoleRef.Name)
	}
	return ans, errs
}

func isStarOrInSet(set sets.Set[string]) func(string) bool {
	return func(str string) bool {
		return str == "*" || set.Has(str)
	}
}

func Id[T any](in T) T { return in }

func fmtSubj(subj rbac.Subject) string {
	switch subj.Kind {
	case "User":
		return "U:" + subj.Name
	case "Group":
		return "G:" + subj.Name
	case "ServiceAccount":
		return "SA:" + subj.Namespace + "/" + subj.Name
	default:
		return subj.Kind + ":" + subj.Name
	}
}
