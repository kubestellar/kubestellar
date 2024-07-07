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

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	ksClient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
)

const (
	timeout = 500 * time.Second
)

func GetConfig(context string) *rest.Config {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return config
}

func CreateKubeClient(config *rest.Config) *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return clientset
}

func CreateKSClient(config *rest.Config) *ksClient.Clientset {
	clientset, err := ksClient.NewForConfig(config)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return clientset
}

func CreateOcmWorkClient(config *rest.Config) *ocmWorkClient.Clientset {
	ocmWorkClient, err := ocmWorkClient.NewForConfig(config)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return ocmWorkClient
}

func CreateDynamicClient(config *rest.Config) *dynamic.DynamicClient {
	dynamicClient, err := dynamic.NewForConfig(config)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return dynamicClient
}

func CreateNS(ctx context.Context, client *kubernetes.Clientset, name string) {
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
	if errors.IsAlreadyExists(err) {
		err = nil
	}
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(func() error {
		_, err := client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Succeed())
}

func Cleanup(ctx context.Context) {
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

func SetupKubestellar(ctx context.Context, releasedFlag bool) {
	var e, o bytes.Buffer
	var args []string
	if releasedFlag {
		args = []string{"--released"}
	}
	commandName := "../common/setup-kubestellar.sh"
	ginkgo.By(fmt.Sprintf("Execing command %v", append([]string{commandName}, args...)))
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
	wds.ControlV1alpha1().BindingPolicies().Delete(ctx, name, metav1.DeleteOptions{})
	gomega.Eventually(func() error {
		_, err := wds.ControlV1alpha1().BindingPolicies().Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Not(gomega.Succeed()))
}

func CreateBindingPolicy(ctx context.Context, wds *ksClient.Clientset, name string,
	clusterSelector []metav1.LabelSelector, testAndStatusCollection []ksapi.DownsyncObjectTestAndStatusCollection) {
	bindingPolicy := ksapi.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ksapi.BindingPolicySpec{
			ClusterSelectors: clusterSelector,
			Downsync:         testAndStatusCollection,
		},
	}
	_, err := wds.ControlV1alpha1().BindingPolicies().Create(ctx, &bindingPolicy, metav1.CreateOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(func() *ksapi.BindingPolicy {
		p, _ := wds.ControlV1alpha1().BindingPolicies().Get(ctx, name, metav1.GetOptions{})
		return p
	}, timeout).Should(gomega.Not(gomega.BeNil()))
}

func CreateStatusCollector(ctx context.Context, wds *ksClient.Clientset, name string, spec ksapi.StatusCollectorSpec) {
	satusCollector := ksapi.StatusCollector{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatusCollector",
			APIVersion: "control.kubestellar.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       spec,
	}
	_, err := wds.ControlV1alpha1().StatusCollectors().Create(ctx, &satusCollector, metav1.CreateOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
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
	var cs *ksapi.CombinedStatus

	p, err := ksClient.ControlV1alpha1().BindingPolicies().Get(ctx, policyName, metav1.GetOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())

	d, err := kubeClient.AppsV1().Deployments(ns).Get(ctx, objectName, metav1.GetOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())

	cs_name := string(d.UID) + "." + string(p.UID)

	gomega.Eventually(func() error {
		cs, err = ksClient.ControlV1alpha1().CombinedStatuses(ns).Get(ctx, cs_name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Not(gomega.HaveOccurred()))

	gomega.Eventually(func() error {
		cs, err = ksClient.ControlV1alpha1().CombinedStatuses(ns).Get(ctx, cs_name, metav1.GetOptions{})
		return err
	}, timeout).ShouldNot(gomega.HaveOccurred())

	// now that CombinedStatus exists, we need to wait some time for it to be completed
	// TODO: find a way to determine completion
	time.Sleep(40 * time.Second)
	cs, err = ksClient.ControlV1alpha1().CombinedStatuses(ns).Get(ctx, cs_name, metav1.GetOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())

	return cs
}

func DeleteDeployment(ctx context.Context, wds *kubernetes.Clientset, ns string, name string) {
	wds.AppsV1().Deployments(ns).Delete(ctx, name, metav1.DeleteOptions{})
	gomega.Eventually(func() error {
		_, err := wds.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		return err
	}, timeout).Should(gomega.Not(gomega.Succeed()))
}

func CreateDeployment(ctx context.Context, wds *kubernetes.Clientset, ns string, name string, labels map[string]string) {
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
	count    int
	problems []string
}

// ValidateNumDeployments waits a limited amount of time for the number of Deployjment objects to equal the given count and
// all the problemFuncs to return the empty string for every Deployment.
func ValidateNumDeployments(ctx context.Context, wec *kubernetes.Clientset, ns string, num int, problemFuncs ...func(*appsv1.Deployment) string) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() countAndProblems {
		deployments, err := wec.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return countAndProblems{-1, []string{err.Error()}}
		}
		ans := countAndProblems{count: len(deployments.Items)}
		for _, pf := range problemFuncs {
			for _, deployment := range deployments.Items {
				if problem := pf(&deployment); len(problem) > 0 {
					ans.problems = append(ans.problems, fmt.Sprintf("deployment %q has a problem: %s", deployment.Name, problem))
				}
			}
		}
		return ans
	}, timeout).Should(gomega.Equal(countAndProblems{count: num}))
}

func ValidateNumServices(ctx context.Context, wec *kubernetes.Clientset, ns string, num int) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		services, err := wec.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(services.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateNumJobs(ctx context.Context, wec *kubernetes.Clientset, ns string, num int) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		jobs, err := wec.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(jobs.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateSingletonStatusZeroValue(ctx context.Context, wds *kubernetes.Clientset, ns string, name string) {
	gomega.Eventually(func() []appsv1.DeploymentCondition {
		deployment, err := wds.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return deployment.Status.Conditions
	}, timeout).Should(gomega.BeNil())
}

func ValidateSingletonStatus(ctx context.Context, wds *kubernetes.Clientset, ns string, name string) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		deployment, err := wds.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return int(deployment.Status.AvailableReplicas)
	}, timeout).Should(gomega.Equal(1))
}

func ValidateNumManifestworks(ctx context.Context, ocmWorkIts *ocmWorkClient.Clientset, ns string, num int) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		list, err := ocmWorkIts.WorkV1().ManifestWorks(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(list.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateNumDeploymentReplicas(ctx context.Context, wec *kubernetes.Clientset, ns string, numReplicas int) {
	ginkgo.GinkgoHelper()
	gomega.Eventually(func() int {
		deployments, err := wec.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		if len(deployments.Items) != 1 {
			return 0
		}
		d := deployments.Items[0]
		print()
		return int(*d.Spec.Replicas)
	}, timeout).Should(gomega.Equal(numReplicas))
}

func DeleteWECDeployments(ctx context.Context, wec *kubernetes.Clientset, ns string) {
	err := wec.AppsV1().Deployments(ns).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
}

// CleanupWDS: removes all deployments and services from ns and all bindingPolicies from cluster
func CleanupWDS(ctx context.Context, wds *kubernetes.Clientset, ksWds *ksClient.Clientset, ns string) {
	DeleteAll[*appsv1.DeploymentList](ctx, wds.AppsV1().Deployments(ns), func(objList *appsv1.DeploymentList) []string {
		return objectsToNames((*appsv1.Deployment).GetName, objList.Items)
	})
	Delete1By1[*corev1.ServiceList](ctx, wds.CoreV1().Services(ns), func(objList *corev1.ServiceList) []string {
		return objectsToNames((*corev1.Service).GetName, objList.Items)
	})
	DeleteAll[*batchv1.JobList](ctx, wds.BatchV1().Jobs(ns), func(objList *batchv1.JobList) []string {
		return objectsToNames((*batchv1.Job).GetName, objList.Items)
	})
	DeleteAll[*ksapi.BindingPolicyList](ctx, ksWds.ControlV1alpha1().BindingPolicies(), func(objList *ksapi.BindingPolicyList) []string {
		return objectsToNames((*ksapi.BindingPolicy).GetName, objList.Items)
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

func Delete1By1[ObjectListType metav1.ListInterface](ctx context.Context, client ResourceInterface[ObjectListType], listNames func(ObjectListType) []string) {
	gomega.Eventually(func() error {
		list, err := client.List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		remaining := listNames(list)
		if len(remaining) == 0 {
			return nil
		}
		errs := []error{fmt.Errorf("some objects remain; their names are: %v", remaining)}
		for _, objName := range remaining {
			err := client.Delete(ctx, objName, metav1.DeleteOptions{})
			if err != nil {
				errs = append(errs, err)
			}
		}
		return goerrors.Join(errs...)
	}, timeout/3).Should(gomega.Succeed())
}

func DeleteAll[ObjectListType metav1.ListInterface](ctx context.Context, client ResourceCollectionInterface[ObjectListType], listNames func(ObjectListType) []string) {
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
	ginkgo.By(fmt.Sprintf("Enumerating Pods in namespace %q with names starting with %q", ns, namePrefix))
	var pods *corev1.PodList
	gomega.EventuallyWithOffset(1, func() error {
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
	gomega.EventuallyWithOffset(1, func() map[string]error {
		problems := map[string]error{}
		for podName := range stillNeedsDelete {
			err := client.CoreV1().Pods(ns).Delete(ctx, podName, metav1.DeleteOptions{})
			if err == nil || errors.IsNotFound(err) {
				delete(stillNeedsDelete, podName)
			} else {
				problems[podName] = err
			}
		}
		return problems
	}, timeout).Should(gomega.BeEmpty())
	ginkgo.By(fmt.Sprintf("Deleted pods %v", goners))
	gomega.EventuallyWithOffset(1, func() map[string]types.UID {
		remaining := map[string]types.UID{}
		for podName, podUID := range goners {
			pod, err := client.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
			if err == nil && pod.UID == podUID || err != nil && !errors.IsNotFound(err) {
				remaining[pod.Name] = pod.UID
			}
		}
		return remaining
	}, timeout).Should(gomega.BeEmpty())
}

func Expect1PodOfEach(ctx context.Context, client *kubernetes.Clientset, ns string, namePrefixes ...string) {
	gomega.EventuallyWithOffset(1, func() map[string]int {
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
