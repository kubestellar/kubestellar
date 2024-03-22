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

func Cleanup() {
	var e, o bytes.Buffer
	cmd := exec.Command("../common/cleanup.sh")
	cmd.Stderr = &e
	cmd.Stdout = &o
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(ginkgo.GinkgoWriter, "%s", o.String())
		fmt.Fprintf(ginkgo.GinkgoWriter, "%s", e.String())
	}
	gomega.Expect(err).To(gomega.Succeed())
}

func SetupKubestellar(releasedFlag bool) {
	var cmd *exec.Cmd
	var e, o bytes.Buffer
	if releasedFlag {
		fmt.Fprintf(ginkgo.GinkgoWriter, "%s", "releasedFlag=true")
		cmd = exec.Command("../common/setup-kubestellar.sh", "--released")
	} else {
		fmt.Fprintf(ginkgo.GinkgoWriter, "%s", "releasedFlag=false")
		cmd = exec.Command("../common/setup-kubestellar.sh")
	}
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
	clusterSelector []metav1.LabelSelector, objectTest []ksapi.DownsyncObjectTest) {
	bindingPolicy := ksapi.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ksapi.BindingPolicySpec{
			ClusterSelectors: clusterSelector,
			Downsync:         objectTest,
		},
	}
	_, err := wds.ControlV1alpha1().BindingPolicies().Create(ctx, &bindingPolicy, metav1.CreateOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(func() *ksapi.BindingPolicy {
		p, _ := wds.ControlV1alpha1().BindingPolicies().Get(ctx, name, metav1.GetOptions{})
		return p
	}, timeout).Should(gomega.Not(gomega.BeNil()))
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

func ValidateNumDeployments(ctx context.Context, wec *kubernetes.Clientset, ns string, num int) {
	gomega.Eventually(func() int {
		deployments, err := wec.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(deployments.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateNumServices(ctx context.Context, wec *kubernetes.Clientset, ns string, num int) {
	gomega.Eventually(func() int {
		services, err := wec.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(services.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateNumJobs(ctx context.Context, wec *kubernetes.Clientset, ns string, num int) {
	gomega.Eventually(func() int {
		jobs, err := wec.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(jobs.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateSingletonStatus(ctx context.Context, wds *kubernetes.Clientset, ns string, name string) {
	gomega.Eventually(func() int {
		deployment, err := wds.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return int(deployment.Status.AvailableReplicas)
	}, timeout).Should(gomega.Equal(1))
}

func ValidateNumManifestworks(ctx context.Context, ocmWorkImbs *ocmWorkClient.Clientset, ns string, num int) {
	gomega.Eventually(func() int {
		list, err := ocmWorkImbs.WorkV1().ManifestWorks(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(list.Items)
	}, timeout).Should(gomega.Equal(num))
}

func ValidateNumDeploymentReplicas(ctx context.Context, wec *kubernetes.Clientset, ns string, numReplicas int) {
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
	deployments, err := wds.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	for _, deployment := range deployments.Items {
		wds.AppsV1().Deployments(ns).Delete(ctx, deployment.GetName(), metav1.DeleteOptions{})
	}
	gomega.Eventually(func() int {
		deployments, err = wds.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(deployments.Items)
	}).Should(gomega.Equal(0))

	services, err := wds.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	for _, service := range services.Items {
		wds.CoreV1().Services(ns).Delete(ctx, service.GetName(), metav1.DeleteOptions{})
	}
	gomega.Eventually(func() int {
		services, err = wds.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(services.Items)
	}).Should(gomega.Equal(0))

	jobs, err := wds.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	for _, job := range jobs.Items {
		wds.BatchV1().Jobs(ns).Delete(ctx, job.GetName(), metav1.DeleteOptions{})
	}
	gomega.Eventually(func() int {
		jobs, err = wds.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(jobs.Items)
	}).Should(gomega.Equal(0))

	bindingPolicies, err := ksWds.ControlV1alpha1().BindingPolicies().List(ctx, metav1.ListOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	for _, bp := range bindingPolicies.Items {
		ksWds.ControlV1alpha1().BindingPolicies().Delete(ctx, bp.GetName(), metav1.DeleteOptions{})
	}
	gomega.Eventually(func() int {
		bindingPolicies, err = ksWds.ControlV1alpha1().BindingPolicies().List(ctx, metav1.ListOptions{})
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		return len(bindingPolicies.Items)
	}, timeout).Should(gomega.Equal(0))
}

func DeletePod(ctx context.Context, client *kubernetes.Clientset, ns string, name string) {
	ginkgo.By("DeletePod")
	pods, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	for _, pod := range pods.Items {
		podName := pod.GetName()
		if strings.HasPrefix(podName, name) {
			err = client.CoreV1().Pods(ns).Delete(ctx, podName, metav1.DeleteOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
		}
	}
}
