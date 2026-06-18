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
	"slices"
	"strings"

	"github.com/spf13/pflag"

	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	discovery "k8s.io/client-go/discovery"
	kubeclient "k8s.io/client-go/kubernetes"
	auth_client "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/abstract"
	"github.com/kubestellar/kubestellar/pkg/util"
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
	userNameFilterOptions := util.NewStringFilterOptions()
	userNameFilterOptions.AddToFlags(fs, "subject-user-names", "subject user names to focus on; '*' means all")
	userGroupFilterOptions := util.NewStringFilterOptions().SeparateBySpacesToo()
	userGroupFilterOptions.AddToFlags(fs, "subject-user-groups", "subject user groups to focus on; '*' means all")
	serviceAccountFilterOptions := util.NewStringFilterOptions()
	serviceAccountFilterOptions.AddToFlags(fs, "subject-service-accounts", "subject service accounts to focus on; syntax for one is namespace:name; '*' means all")
	verbFilterOptions := util.NewStringFilterOptions("*")
	verbFilterOptions.AddToFlags(fs, "verbs", "verbs to focus on; '*' means all")
	resourceFilterOptions := util.NewStringFilterOptions("*")
	resourceFilterOptions.AddToFlags(fs, "resources", "resources to focus on; resource syntax is plural.group/subresource; .group and /subresource are omitted when appropriate; '*' means all resources")
	showRole := true
	fs.BoolVar(&showRole, "show-role", showRole, "include role in listing for resource grans")
	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.FromContext(ctx)
	ctx = klog.NewContext(ctx, logger)

	subjFilter := newSubjectFilter(userNameFilterOptions, userGroupFilterOptions, serviceAccountFilterOptions)
	verbFilter, _ := verbFilterOptions.ToFilter()
	resourceFilter, _ := resourceFilterOptions.ToFilter()

	config, err := cliOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to build config from flags")
		os.Exit(5)
	}

	kubeClient := kubeclient.NewForConfigOrDie(config)
	discoClient := kubeClient.Discovery()
	rscMap, errs := getResourceMap(discoClient)
	if len(errs) != 0 {
		for _, err := range errs {
			logger.Error(err, "Failed to fetch all resources")
		}
	}
	for grStr := range resourceFilter.Literals {
		gr := schema.ParseGroupResource(grStr)
		if _, have := rscMap[metav1.GroupResource(gr)]; !have {
			logger.Error(nil, "Given resource does not exist", "resource", grStr)
		}
	}
	client := kubeClient.RbacV1()
	flat, errs := getFlat(ctx, client, rscMap, subjFilter, verbFilter, resourceFilter)
	switch outputFormat {
	case "table":
		showRsc := resourceFilter.AllPass || len(resourceFilter.Literals) != 1
		tw := printers.GetNewTabWriter(os.Stdout)
		tw.Write([]byte("BINDING\t"))
		if showRole {
			tw.Write([]byte("ROLE\t"))
		}
		tw.Write([]byte("SUBJECT\t"))
		tw.Write([]byte("VERB\t"))
		if showRsc {
			tw.Write([]byte("RESOURCE\t"))
		}
		tw.Write([]byte("OBJNAME\n"))
		for _, tup := range flat {
			for _, verb := range tup.Rule.Verbs {
				for _, rsc := range tup.Rule.Resources {
					objNames := tup.Rule.ObjectNames
					if len(objNames) == 0 {
						objNames = []string{"*"}
					}
					for _, objName := range objNames {
						tw.Write([]byte(tup.Binding.String() + "\t"))
						if showRole {
							if tup.RoleInCluster {
								tw.Write([]byte("/"))
							}
							tw.Write([]byte(tup.RoleName + "\t"))
						}
						tw.Write([]byte(fmtSubj(tup.Subject) + "\t"))
						tw.Write([]byte(verb + "\t"))
						if showRsc {
							tw.Write([]byte(rsc + "\t"))
						}
						tw.Write([]byte(objName + "\n"))
					}
				}
			}
		}
		err := tw.Flush()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		if resourceFilter.AllPass {
			fmt.Println()
			tw = printers.GetNewTabWriter(os.Stdout)
			tw.Write([]byte("BINDING\t"))
			if showRole {
				tw.Write([]byte("ROLE\t"))
			}
			tw.Write([]byte("SUBJECT\t"))
			tw.Write([]byte("VERB\t"))
			tw.Write([]byte("PATH\n"))
			for _, tup := range flat {
				for _, verb := range tup.Rule.Verbs {
					for _, path := range tup.Rule.NonResourcePaths {
						tw.Write([]byte(tup.Binding.String() + "\t"))
						if showRole {
							if tup.RoleInCluster {
								tw.Write([]byte("/"))
							}
							tw.Write([]byte(tup.RoleName + "\t"))
						}
						tw.Write([]byte(fmtSubj(tup.Subject) + "\t"))
						tw.Write([]byte(verb + "\t"))
						tw.Write([]byte(path + "\n"))
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

// resourceMap maps resource to `clusterScoped bool`.
type resourceMap map[metav1.GroupResource]bool

func getResourceMap(discoClient discovery.DiscoveryInterface) (resourceMap, []error) {
	_, rscList, err := discoClient.ServerGroupsAndResources()
	if err != nil {
		return nil, []error{err}
	}
	rm := resourceMap{}
	errs := []error{}
	for _, rsc := range rscList {
		gv, err := schema.ParseGroupVersion(rsc.GroupVersion)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for _, ar := range rsc.APIResources {
			gr := metav1.GroupResource{Group: gv.Group, Resource: ar.Name}
			rm[gr] = !ar.Namespaced
		}
	}
	return rm, errs
}

type subjectFilter struct {
	UserName       util.StringFilter
	UserGroup      util.StringFilter
	ServiceAccount util.StringFilter
}

func newSubjectFilter(userName, userGroup, serviceAccount util.StringFilterOptions) subjectFilter {
	userNameFilter, filteringOnName := userName.ToFilter()
	userGroupFilter, filteringOnGroup := userGroup.ToFilter()
	svcAcctFilter, filteringOnSvcAcct := serviceAccount.ToFilter()
	if !(filteringOnName || filteringOnGroup || filteringOnSvcAcct) {
		userNameFilter = util.StringFilter{AllPass: true}
		userGroupFilter = util.StringFilter{AllPass: true}
		svcAcctFilter = util.StringFilter{AllPass: true}
	}
	return subjectFilter{
		UserName:       userNameFilter,
		UserGroup:      userGroupFilter,
		ServiceAccount: svcAcctFilter}
}

func (sf *subjectFilter) Passes(subj rbac.Subject) bool {
	if subj.APIGroup != "" && subj.APIGroup != "rbac.authorization.k8s.io" {
		return false
	}
	switch subj.Kind {
	case "User":
		return sf.UserName.Passes(subj.Name, false)
	case "Group":
		return sf.UserGroup.Passes(subj.Name, false)
	case "ServiceAccount":
		return sf.ServiceAccount.Passes(subj.Namespace+":"+subj.Name, false)
	default:
		return false
	}
}

type Flat []Tuple

type Tuple struct {
	Binding       NamespacedName
	RoleInCluster bool
	RoleName      string
	Subject       rbac.Subject
	Rule          PolicyRule
}

type PolicyRule struct {
	Verbs            []string
	Resources        []string
	ObjectNames      []string
	NonResourcePaths []string
}

func getFlat(ctx context.Context, client auth_client.RbacV1Interface, rscMap resourceMap, subjFilter subjectFilter, verbFilter, rscFilter util.StringFilter) (Flat, []error) {
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
	complainedCSRoles := sets.New[NamespacedName]()
	complainedNRRoles := sets.New[NamespacedName]()
	ans := Flat{}
	addProducts := func(rules []rbac.PolicyRule, subjects []rbac.Subject, source NamespacedName, roleInCluster bool, roleName string) {
		roleNN := NamespacedName{Name: roleName}
		if !roleInCluster {
			roleNN.Namespace = source.Namespace
		}
		complainCS, complainNR := false, false
		complainedResources := sets.New[metav1.GroupResource]()
		for ruleIdx, rule := range rules {
			ruledResources := composeRuleResources(rscMap, rule.APIGroups, rule.Resources)
			if len(ruledResources) == 0 {
				complainNR = true
				if !complainedNRRoles.Has(roleNN) {
					errs = append(errs, fmt.Errorf("rule[%d] in role-ism %s refers to no existing resources (APIGroups=%#v, Resources=%#v)", ruleIdx, roleNN.String(), rule.APIGroups, rule.Resources))
				}
				continue
			}
			var ruledResourceStrs []string
			for rr := range ruledResources {
				clusterScoped := rscMap[rr]
				if clusterScoped && source.Namespace != metav1.NamespaceNone {
					complainCS = true
					if !complainedCSRoles.Has(roleNN) && !complainedResources.Has(rr) {
						errs = append(errs, fmt.Errorf("namespace-bound role-ism %s refers to cluster-scoped resource %s", roleNN.String(), rr.String()))
						complainedResources.Insert(rr)
					}
					continue
				}
				rrStr := rr.String()
				if rscFilter.Passes(rrStr, false) {
					ruledResourceStrs = append(ruledResourceStrs, rrStr)
				}
			}
			if len(ruledResourceStrs) == 0 {
				continue
			}
			slices.Sort(ruledResourceStrs)
			if len(rule.ResourceNames) == 0 {
				rule.ResourceNames = []string{"*"}
			}
			pr := PolicyRule{
				Verbs:            verbFilter.FilterSlice(rule.Verbs, true),
				Resources:        ruledResourceStrs,
				ObjectNames:      rule.ResourceNames,
				NonResourcePaths: rule.NonResourceURLs,
			}
			for _, subj := range subjects {
				if subjFilter.Passes(subj) {
					tup := Tuple{Binding: source,
						RoleInCluster: roleInCluster,
						RoleName:      roleName,
						Subject:       subj,
						Rule:          pr}
					ans = append(ans, tup)
				}
			}
		}
		if complainCS {
			errs = append(errs, fmt.Errorf("binding %s refers to role-ism %s with cluster-scoped resources", source.String(), roleNN.String()))
			complainedCSRoles.Insert(roleNN)
		}
		if complainNR {
			complainedNRRoles.Insert(roleNN)
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
		addProducts(rules, crb.Subjects, br, true, crb.RoleRef.Name)
	}
	for _, rb := range rbList {
		br := NamespacedName{Namespace: rb.Namespace, Name: rb.Name}
		var rules []rbac.PolicyRule
		var roleInCluster bool
		switch rb.RoleRef.Kind {
		case "ClusterRole":
			if cr, ok := crMap[rb.RoleRef.Name]; ok {
				rules = cr.Rules
			} else {
				errs = append(errs, fmt.Errorf("RoleBinding %s/%s references unknown ClusterRole: %q", rb.Namespace, rb.Name, rb.RoleRef.Name))
			}
			roleInCluster = true
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
		addProducts(rules, rb.Subjects, br, roleInCluster, rb.RoleRef.Name)
	}
	return ans, errs
}

func composeRuleResources(rscMap resourceMap, apiGroups, resources []string) sets.Set[metav1.GroupResource] {
	if len(apiGroups) == 0 {
		apiGroups = []string{"*"}
	}
	apiGroupSet := sets.New(apiGroups...)
	allAPIGroups := apiGroupSet.Has("*")
	if len(resources) == 0 {
		resources = []string{"*"}
	}
	resourceSet := sets.New[string]()
	for _, rsc := range resources {
		rscParts := strings.SplitN(rsc, "/", 2)
		resourceSet.Insert(rscParts[0])
	}
	allResources := resourceSet.Has("*")
	ans := sets.New[metav1.GroupResource]()
	for gr := range rscMap {
		if !allAPIGroups && !apiGroupSet.Has(gr.Group) {
			continue
		}
		if allResources || resourceSet.Has(gr.Resource) {
			ans.Insert(gr)
		}
	}
	return ans
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
