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

package util

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	ocmWorkClient "open-cluster-management.io/api/client/work/clientset/versioned"
	v1 "open-cluster-management.io/api/work/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	ksClient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
)

const (
	timeout = 500 * time.Second
)

func GetConfig(context string) *rest.Config {
	ginkgo.GinkgoHelper()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return config
}

func CreateKubeClient(config *rest.Config) *kubernetes.Clientset {
	ginkgo.GinkgoHelper()
	clientset, err := kubernetes.NewForConfig(config)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return clientset
}

func CreateKSClient(config *rest.Config) *ksClient.Clientset {
	ginkgo.GinkgoHelper()
	clientset, err := ksClient.NewForConfig(config)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return clientset
}

func CreateOcmWorkClient(config *rest.Config) *ocmWorkClient.Clientset {
	ginkgo.GinkgoHelper()
	ocmWorkClient, err := ocmWorkClient.NewForConfig(config)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return ocmWorkClient
}

func CreateDynamicClient(config *rest.Config) *dynamic.DynamicClient {
	ginkgo.GinkgoHelper()
	dynamicClient, err := dynamic.NewForConfig(config)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return dynamicClient
}

func CreateNS(ctx context.Context, client *kubernetes.Clientset, name string) {
	ginkgo.GinkgoHelper()
	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = "nginx"
	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "",
			Labels:    labels,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, nsObj, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		err = nil
	}
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(func() error {
		_, err := client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
}

func Cleanup(ctx context.Context) {
	ginkgo.GinkgoHelper()
	var e, o bytes.Buffer
	cmd := exec.CommandContext(ctx, "../common/cleanup.sh")
	cmd.Stderr = &e
	cmd.Stdout = &o
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(ginkgo.GinkgoWriter, "%s", o.String())
		fmt.Fprintf(ginkgo.GinkgoWriter, "%s", e.String())
	}
	gomega.Expect(err).To(gomega.Succeed())
}

func SetupKubestellar(ctx context.Context, releasedFlag bool, otherFlags ...string) {
	ginkgo.GinkgoHelper()
	var e, o bytes.Buffer
	var args []string
	if releasedFlag {
		args = []string{"--released"}
	}
	args = append(args, otherFlags...)
	commandName := "../common/setup-kubestellar.sh"
	ginkgo.By(fmt.Sprintf("Execing command %#v", append([]string{commandName}, args...)))
	cmd := exec.CommandContext(ctx, commandName, args...)
	cmd.Stderr = &e
	cmd.Stdout = &o
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(ginkgo.GinkgoWriter, "%s", o.String())
		fmt.Fprintf(ginkgo.GinkgoWriter, "%s", e.String())
	}
	gomega.Expect(err).To(gomega.Succeed())
}

func int32Ptr(i int32) *int32 {
	return &i
}

func DeleteBindingPolicy(ctx context.Context, wds *ksClient.Clientset, name string) {
	ginkgo.GinkgoHelper()
	wds.ControlV1alpha1().BindingPolicies().Delete(ctx, name, metav1.DeleteOptions{})
	gomega.Eventually(func() error {
		_, err := wds.ControlV1alpha1().BindingPolicies().Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Not(gomega.Succeed()))
}

func CreateBindingPolicy(ctx context.Context, wds *ksClient.Clientset, name string,
	clusterSelector []metav1.LabelSelector, testAndStatusCollection []ksapi.DownsyncPolicyClause,
	mutators ...func(*ksapi.BindingPolicy) error) {
	ginkgo.GinkgoHelper()
	bindingPolicy := ksapi.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ksapi.BindingPolicySpec{
			ClusterSelectors: clusterSelector,
			Downsync:         testAndStatusCollection,
		},
	}
	for _, m := range mutators {
		err := m(&bindingPolicy)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
	_, err := wds.ControlV1alpha1().BindingPolicies().Create(ctx, &bindingPolicy, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(func() *ksapi.BindingPolicy {
		p, _ := wds.ControlV1alpha1().BindingPolicies().Get(ctx, name, metav1.GetOptions{})
		return p
	}, timeout).Should(gomega.Not(gomega.BeNil()))
}

func CreateStatusCollector(ctx context.Context, wds *ksClient.Clientset, name string, spec ksapi.StatusCollectorSpec) {
	ginkgo.GinkgoHelper()
	satusCollector := ksapi.StatusCollector{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatusCollector",
			APIVersion: "control.kubestellar.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       spec,
	}
	_, err := wds.ControlV1alpha1().StatusCollectors().Create(ctx, &satusCollector, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(func() *ksapi.StatusCollector {
		p, _ := wds.ControlV1alpha1().StatusCollectors().Get(ctx, name, metav1.GetOptions{})
		return p
	}, timeout).Should(gomega.Not(gomega.BeNil()))
}

// GetCombinedStatus is a helper func which expects the workload object to be a Deployment.
// CombinedStatus name is the concatenation of:
// - the UID of the workload object
// - the string "."
// - the UID of the BindingPolicy object.
func GetCombinedStatus(ctx context.Context, ksClient *ksClient.Clientset, kubeClient *kubernetes.Clientset, ns, objectName, policyName string) *ksapi.CombinedStatus {
	ginkgo.GinkgoHelper()
	var cs *ksapi.CombinedStatus

	p, err := ksClient.ControlV1alpha1().BindingPolicies().Get(ctx, policyName, metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	d, err := kubeClient.AppsV1().Deployments(ns).Get(ctx, objectName, metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	cs_name := string(d.UID) + "." + string(p.UID)

	gomega.Eventually(func() error {
		cs, err = ksClient.ControlV1alpha1().CombinedStatuses(ns).Get(ctx, cs_name, metav1.GetOptions{})
		return err
	}, timeout).ShouldNot(gomega.HaveOccurred())

	// now that CombinedStatus exists, we need to wait some time for it to be completed
	// TODO: find a way to determine completion
	time.Sleep(40 * time.Second)
	cs, err = ksClient.ControlV1alpha1().CombinedStatuses(ns).Get(ctx, cs_name, metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return cs
}

func DeleteDeployment(ctx context.Context, wds *kubernetes.Clientset, ns string, name string) {
	ginkgo.GinkgoHelper()
	wds.AppsV1().Deployments(ns).Delete(ctx, name, metav1.DeleteOptions{})
	gomega.Eventually(func() error {
		_, err := wds.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Not(gomega.Succeed()))
}

func CreateDeployment(ctx context.Context, wds *kubernetes.Clientset, ns string, name string, labels map[string]string) {
	ginkgo.GinkgoHelper()
	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "public.ecr.aws/nginx/nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	gomega.Eventually(func() error {
		_, err := wds.AppsV1().Deployments(ns).Create(ctx, &deployment, metav1.CreateOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
}

func CreateCustomTransform(ctx context.Context, wds *ksClient.Clientset, name, apiGroup, resource string, remove ...string) {
	ginkgo.GinkgoHelper()
	ct := ksapi.CustomTransform{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ksapi.CustomTransformSpec{
			APIGroup: apiGroup,
			Resource: resource,
			Remove:   remove,
		},
	}
	gomega.Eventually(func() error {
		_, err := wds.ControlV1alpha1().CustomTransforms().Create(ctx, &ct, metav1.CreateOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
}

func CreateService(ctx context.Context, wds *kubernetes.Clientset, ns string, name string, appName string) {
	ginkgo.GinkgoHelper()
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"app.kubernetes.io/name": appName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name": appName,
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	gomega.Eventually(func() error {
		_, err := wds.CoreV1().Services(ns).Create(ctx, &service, metav1.CreateOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
}

func CreateJob(ctx context.Context, wds *kubernetes.Clientset, ns string, name string, appName string) {
	ginkgo.GinkgoHelper()
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pi-",
			Labels: map[string]string{
				"app.kubernetes.io/name": appName,
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pi",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "pi",
							Image: "perl",
							Command: []string{
								"perl",
								"-Mbignum=bpi",
								"-wle",
								"print bpi(2000)",
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
	gomega.Eventually(func() error {
		_, err := wds.BatchV1().Jobs(ns).Create(ctx, &job, metav1.CreateOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
}

type countAndProblems struct {
	Count    int
	Problems []string
}

func GetDeploymentTime(ctx context.Context, clientset *kubernetes.Clientset, ns string, name string) time.Time {
	var deployment *appsv1.Deployment
	gomega.Eventually(func() error {
		var err error
		deployment, err = clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
	return deployment.CreationTimestamp.Time
}

func GetBindingTime(ctx context.Context, clientset *ksClient.Clientset, name string) time.Time {
	var binding *ksapi.Binding
	gomega.Eventually(func() error {
		var err error
		binding, err = clientset.ControlV1alpha1().Bindings().Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
	return binding.CreationTimestamp.Time
}

func GetManifestworkTime(ctx context.Context, ocmWorkImbs *ocmWorkClient.Clientset, ns string, name string) time.Time {
	var manifestwork *v1.ManifestWork
	gomega.Eventually(func() error {
		var err error
		manifestwork, err = ocmWorkImbs.WorkV1().ManifestWorks(ns).Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
	return manifestwork.CreationTimestamp.Time
}

// ValidateNumDeployments waits a limited amount of time for the number of Deployment objects to equal the given count and
// all the problemFuncs to return the empty string for every Deployment.
func ValidateNumDeployments(ctx context.Context, where string, wec *kubernetes.Clientset, ns string, num int, problemFuncs ...func(*appsv1.Deployment) string) {
	ginkgo.GinkgoHelper()
	logger := klog.FromContext(ctx)
	lastAns := countAndProblems{}
	gomega.Eventually(func(ctx context.Context) countAndProblems {
		deployments, err := wec.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.V(2).Info("Failed to list Deployment objects", "ns", ns)
			return countAndProblems{-1, []string{err.Error()}}
		}
		ans := countAndProblems{Count: len(deployments.Items)}
		for _, pf := range problemFuncs {
			for _, deployment := range deployments.Items {
				if problem := pf(&deployment); len(problem) > 0 {
					ans.Problems = append(ans.Problems, fmt.Sprintf("deployment %q has a problem: %s", deployment.Name, problem))
				}
			}
		}
		if ans.Count == num || ans.Count != lastAns.Count || len(ans.Problems) != len(lastAns.Problems) {
			ginkgo.GinkgoLogr.Info("Got Deployment survey result", "where", where, "ns", ns, "expected", num, "result", ans, "now", time.Now())
			logger.V(3).Info("Current Deployment survey result", "where", where, "ns", ns, "expected", num, "result", ans)
		}
		lastAns = ans
		return ans
	}, timeout).WithContext(ctx).Should(gomega.Equal(countAndProblems{Count: num}))
}

func ValidateNumServices(ctx context.Context, wec *kubernetes.Clientset, ns string, num int) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		services, err := wec.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return len(services.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateNumJobs(ctx context.Context, wec *kubernetes.Clientset, ns string, num int) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		jobs, err := wec.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return len(jobs.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateSingletonStatusZeroValue(ctx context.Context, wds *kubernetes.Clientset, ns string, name string) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() []appsv1.DeploymentCondition {
		deployment, err := wds.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return deployment.Status.Conditions
	}, timeout).Should(gomega.BeNil())
}

func ValidateSingletonStatusNonZeroValue(ctx context.Context, wds *kubernetes.Clientset, ns string, name string) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() []appsv1.DeploymentCondition {
		deployment, err := wds.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return deployment.Status.Conditions
	}, timeout).Should(gomega.Not(gomega.BeNil()))
}

func ValidateSingletonStatus(ctx context.Context, wds *kubernetes.Clientset, ns string, name string) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		deployment, err := wds.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return int(deployment.Status.AvailableReplicas)
	}, timeout).Should(gomega.Equal(1))
}

func ValidateBinding(ctx context.Context, wds ksClient.Interface, name string, validate func(*ksapi.Binding) bool) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() bool {
		binding, err := wds.ControlV1alpha1().Bindings().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			klog.FromContext(ctx).V(2).Info("Get Binding failed", "name", name, "err", err)
			return false
		}
		return validate(binding)
	}, timeout).Should(gomega.Equal(true))

}

func ValidateNumManifestworks(ctx context.Context, ocmWorkIts *ocmWorkClient.Clientset, ns string, num int) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		list, err := ocmWorkIts.WorkV1().ManifestWorks(ns).List(ctx, metav1.ListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return len(list.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateNumDeploymentReplicas(ctx context.Context, wec *kubernetes.Clientset, ns string, numReplicas int) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		deployments, err := wec.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		if len(deployments.Items) != 1 {
			return 0
		}
		d := deployments.Items[0]
		print()
		return int(*d.Spec.Replicas)
	}, timeout).Should(gomega.Equal(numReplicas))
}

func DeleteWECDeployments(ctx context.Context, wec *kubernetes.Clientset, ns string) {
	ginkgo.GinkgoHelper()
	err := wec.AppsV1().Deployments(ns).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// CleanupWDS: removes all deployments and services from ns and all bindingPolicies from cluster
func CleanupWDS(ctx context.Context, wds *kubernetes.Clientset, ksWds *ksClient.Clientset, ns string) {
	ginkgo.GinkgoHelper()
	DeleteAll[*appsv1.DeploymentList](ctx, wds.AppsV1().Deployments(ns), func(objList *appsv1.DeploymentList) []string {
		return objectsToNames((*appsv1.Deployment).GetName, objList.Items)
	})
	Delete1By1[*corev1.ServiceList](ctx, "Service", wds.CoreV1().Services(ns), func(objList *corev1.ServiceList) []string {
		return objectsToNames((*corev1.Service).GetName, objList.Items)
	})
	DeleteAll[*batchv1.JobList](ctx, wds.BatchV1().Jobs(ns), func(objList *batchv1.JobList) []string {
		return objectsToNames((*batchv1.Job).GetName, objList.Items)
	})
	DeleteAll[*ksapi.BindingPolicyList](ctx, ksWds.ControlV1alpha1().BindingPolicies(), func(objList *ksapi.BindingPolicyList) []string {
		return objectsToNames((*ksapi.BindingPolicy).GetName, objList.Items)
	})
	DeleteAll[*ksapi.StatusCollectorList](ctx, ksWds.ControlV1alpha1().StatusCollectors(), func(objList *ksapi.StatusCollectorList) []string {
		return objectsToNames((*ksapi.StatusCollector).GetName, objList.Items)
	})
}

type ResourceInterface[ObjectListType metav1.ListInterface] interface {
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	List(ctx context.Context, opts metav1.ListOptions) (ObjectListType, error)
}

type ResourceCollectionInterface[ObjectListType metav1.ListInterface] interface {
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	List(ctx context.Context, opts metav1.ListOptions) (ObjectListType, error)
}

func Delete1By1[ObjectListType metav1.ListInterface](ctx context.Context, kind string, client ResourceInterface[ObjectListType], listNames func(ObjectListType) []string) {
	ginkgo.GinkgoHelper()
	iteration := 1
	gomega.Eventually(func() error {
		list, err := client.List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		remaining := listNames(list)
		if len(remaining) == 0 {
			return nil
		}
		errs := []error{fmt.Errorf("some objects remained at start of iteration %d; their names are: %v", iteration, remaining)}
		for _, objName := range remaining {
			err := client.Delete(ctx, objName, metav1.DeleteOptions{})
			if err != nil {
				ginkgo.GinkgoLogr.Error(err, "Failed to delete an object", "iteration", iteration, "kind", kind, "name", objName)
				errs = append(errs, err)
			} else {
				ginkgo.GinkgoLogr.Info("Deleted object", "iteration", iteration, "kind", kind, "name", objName)
			}
		}
		iteration++
		return goerrors.Join(errs...)
	}, timeout/3).Should(gomega.Succeed())
}

func DeleteAll[ObjectListType metav1.ListInterface](ctx context.Context, client ResourceCollectionInterface[ObjectListType], listNames func(ObjectListType) []string) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() error {
		err := client.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
		if err != nil {
			return err
		}
		list, err := client.List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		remaining := listNames(list)
		if len(remaining) != 0 {
			return fmt.Errorf("some objects remain; their names are: %v", remaining)
		}
		return nil
	}, timeout/3).Should(gomega.Succeed())
}

func objectsToNames[ObjectType any](getName func(*ObjectType) string, objects []ObjectType) []string {
	ans := make([]string, len(objects))
	for idx := range objects {
		ans[idx] = getName(&objects[idx])
	}
	return ans
}

func DeletePods(ctx context.Context, client *kubernetes.Clientset, ns string, namePrefix string) {
	ginkgo.GinkgoHelper()
	ginkgo.By(fmt.Sprintf("Enumerating Pods in namespace %q with names starting with %q", ns, namePrefix))
	var pods *corev1.PodList
	gomega.Eventually(func() error {
		var err error
		pods, err = client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
	goners := map[string]types.UID{}
	stillNeedsDelete := map[string]types.UID{}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, namePrefix) {
			goners[pod.Name] = pod.UID
			stillNeedsDelete[pod.Name] = pod.UID
		}
	}
	gomega.Eventually(func() map[string]error {
		problems := map[string]error{}
		for podName := range stillNeedsDelete {
			err := client.CoreV1().Pods(ns).Delete(ctx, podName, metav1.DeleteOptions{})
			if err == nil || k8serrors.IsNotFound(err) {
				delete(stillNeedsDelete, podName)
			} else {
				problems[podName] = err
			}
		}
		return problems
	}, timeout).Should(gomega.BeEmpty())
	ginkgo.By(fmt.Sprintf("Deleted pods %v", goners))
	gomega.Eventually(func() map[string]types.UID {
		remaining := map[string]types.UID{}
		for podName, podUID := range goners {
			pod, err := client.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
			if err == nil && pod.UID == podUID || err != nil && !k8serrors.IsNotFound(err) {
				remaining[pod.Name] = pod.UID
			}
		}
		return remaining
	}, timeout).Should(gomega.BeEmpty())
}

func Expect1PodOfEach(ctx context.Context, client *kubernetes.Clientset, ns string, namePrefixes ...string) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() map[string]int {
		pods, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return map[string]int{err.Error(): 1}
		}
		problems := map[string]int{}
		for _, namePrefix := range namePrefixes {
			count := 0
			for _, pod := range pods.Items {
				if strings.HasPrefix(pod.Name, namePrefix) {
					for _, cond := range pod.Status.Conditions {
						if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
							count += 1
						}
					}
				}
			}
			if count != 1 {
				problems[namePrefix] = count
			}
		}
		return problems
	}, timeout).Should(gomega.BeEmpty())
	ginkgo.By(fmt.Sprintf("Waited for ready pods in namespace %q with name prefixes %v", ns, namePrefixes))
}

func ScaleDeployment(ctx context.Context, client *kubernetes.Clientset, ns string, name string, target int32) {
	ginkgo.GinkgoHelper()
	ginkgo.By(fmt.Sprintf("Scale Deployment %q in namespace %q to %d", name, ns, target))
	gomega.Eventually(ctx, func(ctx context.Context) error {
		gotSc, err := client.AppsV1().Deployments(ns).GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		gotSc.Spec.Replicas = target
		_, err = client.AppsV1().Deployments(ns).UpdateScale(ctx, name, gotSc, metav1.UpdateOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
}

func WaitForDepolymentAvailability(ctx context.Context, client kubernetes.Interface, ns string, name string) {
	ginkgo.GinkgoHelper()
	var target int32
	gomega.Eventually(func() error {
		gotDeploy, err := client.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if gotDeploy.Generation != gotDeploy.Status.ObservedGeneration {
			return fmt.Errorf("got Status.ObservedGeneration=%d but Generation=%d", gotDeploy.Status.ObservedGeneration, gotDeploy.Generation)
		}
		if gotDeploy.Spec.Replicas == nil {
			return fmt.Errorf("desired number of replicas is not set")
		}
		target = *gotDeploy.Spec.Replicas
		if gotDeploy.Status.AvailableReplicas != *gotDeploy.Spec.Replicas {
			return fmt.Errorf("got Status.AvailableReplicas=%d but Spec.Replicas=%d", gotDeploy.Status.AvailableReplicas, *gotDeploy.Spec.Replicas)
		}
		if gotDeploy.Status.UnavailableReplicas != 0 {
			return fmt.Errorf("got UnavailableReplicas=%d", gotDeploy.Status.UnavailableReplicas)
		}
		return nil
	}, timeout).Should(gomega.Succeed())
	ginkgo.GinkgoLogr.Info("Deployment has desired AvailableReplicas", "name", name, "namespace", ns, "target", target)
}

func ReadContainerArgsInDeployment(ctx context.Context, client *kubernetes.Clientset, ns string, deploymentName string, containerName string) []string {
	ginkgo.GinkgoHelper()
	var args []string
	gomega.Eventually(func() error {
		deploy, err := client.AppsV1().Deployments(ns).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		foundContainer := false
		for _, c := range deploy.Spec.Template.Spec.Containers {
			if c.Name == containerName {
				foundContainer = true
				args = c.Args
				break
			}
		}
		if !foundContainer {
			return fmt.Errorf("container %q in Deployment %q is not found", containerName, deploymentName)
		}
		return nil
	}, timeout).Should(gomega.Succeed())
	return args
}
