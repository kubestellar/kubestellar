package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/apimachinery/pkg/runtime"
)

type ContextInfo struct {
	Name    string `json:"name"`
	Cluster string `json:"cluster"`
}

type ManagedClusterInfo struct {
	Name         string            `json:"name"`
	Labels       map[string]string `json:"labels"`
	CreationTime string            `json:"creationTime"`
}

type WorkloadInfo struct {
	Name         string `json:"name"`
	Kind         string `json:"kind"` // 'Deployment' or 'Service'
	Namespace    string `json:"namespace"`
	CreationTime string `json:"creationTime"`
}

type BindingPolicy struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	ClusterName string `json:"clusterName"`
}

func main() {
	router := gin.Default()

	// CORS Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	router.GET("/api/clusters", func(c *gin.Context) {
		contexts, clusters, currentContext, err, itsData := getKubeInfo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"contexts":       contexts,
			"clusters":       clusters,
			"currentContext": currentContext,
			"itsData":        itsData,
		})
	})

	router.GET("/api/wds/workloads", func(c *gin.Context) {
		workloads, err := getWDSWorkloads()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, workloads)
	})

	// New Log Endpoint
	router.GET("/api/log", func(c *gin.Context) {
		// Fetch Kubernetes Information
		contexts, clusters, currentContext, _, itsData := getKubeInfo()

		// Fetch WDS Workloads
		workloads, err := getWDSWorkloads()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var logBuilder strings.Builder

		logBuilder.WriteString("==== KubestellarUI Log ====\n\n")

		logBuilder.WriteString(fmt.Sprintf("Current Context: %s\n\n", currentContext))

		logBuilder.WriteString("=== Clusters ===\n")
		for _, cluster := range clusters {
			logBuilder.WriteString(fmt.Sprintf("- %s\n", cluster))
		}

		logBuilder.WriteString("\n=== Contexts ===\n")
		for _, ctx := range contexts {
			logBuilder.WriteString(fmt.Sprintf("- %s (Cluster: %s)\n", ctx.Name, ctx.Cluster))
		}

		logBuilder.WriteString("\n=== ITS Data ===\n")
		for _, cluster := range itsData {
			logBuilder.WriteString(fmt.Sprintf("- Name: %s\n", cluster.Name))
			logBuilder.WriteString("  Labels:\n")
			for key, value := range cluster.Labels {
				logBuilder.WriteString(fmt.Sprintf("    %s=%s\n", key, value))
			}
			logBuilder.WriteString(fmt.Sprintf("  Creation Time: %s\n", cluster.CreationTime))
		}

		logBuilder.WriteString("\n=== WDS Workloads ===\n")
		for _, workload := range workloads {
			logBuilder.WriteString(fmt.Sprintf("- Name: %s\n", workload.Name))
			logBuilder.WriteString(fmt.Sprintf("  Kind: %s\n", workload.Kind))
			logBuilder.WriteString(fmt.Sprintf("  Namespace: %s\n", workload.Namespace))
			logBuilder.WriteString(fmt.Sprintf("  Creation Time: %s\n\n", workload.CreationTime))
		}

		// Set Headers for File Download
		c.Header("Content-Type", "text/plain")
		c.Header("Content-Disposition", "attachment; filename=kubestellarui.log")

		// Send the Log String
		c.String(http.StatusOK, logBuilder.String())
	})

	router.GET("/api/binding-policies", func(c *gin.Context) {
		policies, err := getBindingPolicies()
		if err != nil {
			fmt.Printf("Error getting binding policies: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, policies)
	})

	router.Run(":4000")
}

func getKubeInfo() ([]ContextInfo, []string, string, error, []ManagedClusterInfo) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := homeDir(); home != "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		}
	}

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, nil, "", err, nil
	}

	var contexts []ContextInfo
	clusterSet := make(map[string]bool) // Use map to track unique clusters

	// Get contexts and their associated clusters
	for contextName, context := range config.Contexts {
		contexts = append(contexts, ContextInfo{
			Name:    contextName,
			Cluster: context.Cluster,
		})
		clusterSet[context.Cluster] = true
	}

	// Convert unique clusters to slice
	var clusters []string
	for clusterName := range clusterSet {
		clusters = append(clusters, clusterName)
	}

	itsData, err := getITSInfo()
	if err != nil {
		fmt.Printf("ITS error: %v\n", err) // Debug print
		// Don't return error, continue with other data
	}

	return contexts, clusters, config.CurrentContext, nil, itsData
}

func getITSInfo() ([]ManagedClusterInfo, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := homeDir(); home != "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	clustersBytes, err := clientset.RESTClient().Get().
		AbsPath("/apis/cluster.open-cluster-management.io/v1").
		Resource("managedclusters").
		DoRaw(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get managed clusters: %w", err)
	}

	var clusterList struct {
		Items []struct {
			Metadata struct {
				Name              string            `json:"name"`
				Labels            map[string]string `json:"labels"`
				CreationTimestamp string            `json:"creationTimestamp"`
			} `json:"metadata"`
		} `json:"items"`
	}

	if err := json.Unmarshal(clustersBytes, &clusterList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clusters: %w", err)
	}

	var managedClusters []ManagedClusterInfo
	for _, item := range clusterList.Items {
		managedClusters = append(managedClusters, ManagedClusterInfo{
			Name:         item.Metadata.Name,
			Labels:       item.Metadata.Labels,
			CreationTime: item.Metadata.CreationTimestamp,
		})
	}

	return managedClusters, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getWDSWorkloads() ([]WorkloadInfo, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := homeDir(); home != "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		}
	}

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Use WDS1 context specifically
	ctxContext := config.Contexts["wds1"]
	if ctxContext == nil {
		return nil, fmt.Errorf("WDS1 context not found in kubeconfig")
	}

	// Create config for WDS cluster
	clientConfig := clientcmd.NewDefaultClientConfig(
		*config,
		&clientcmd.ConfigOverrides{
			CurrentContext: "wds1",
		},
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	// Get Deployments
	deployments, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}

	// Get Services
	services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	var workloads []WorkloadInfo

	// Add deployments to workloads
	for _, deployment := range deployments.Items {
		workloads = append(workloads, WorkloadInfo{
			Name:         deployment.Name,
			Kind:         "Deployment",
			Namespace:    deployment.Namespace,
			CreationTime: deployment.CreationTimestamp.String(),
		})
	}

	// Add services to workloads
	for _, service := range services.Items {
		workloads = append(workloads, WorkloadInfo{
			Name:         service.Name,
			Kind:         "Service",
			Namespace:    service.Namespace,
			CreationTime: service.CreationTimestamp.String(),
		})
	}

	return workloads, nil
}

func getBindingPolicies() ([]BindingPolicy, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := homeDir(); home != "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// Get binding policies using dynamic client since it's a CRD
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	bindingPoliciesGVR := schema.GroupVersionResource{
		Group:    "policy.kubestellar.io",
		Version:  "v1alpha1",
		Resource: "bindingpolicies",
	}

	policies, err := dynamicClient.Resource(bindingPoliciesGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get binding policies: %w", err)
	}

	var bindingPolicies []BindingPolicy
	for _, policy := range policies.Items {
		var bp BindingPolicy
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(policy.Object, &bp); err != nil {
			return nil, fmt.Errorf("failed to convert binding policy: %w", err)
		}
		bindingPolicies = append(bindingPolicies, bp)
	}

	return bindingPolicies, nil
}
