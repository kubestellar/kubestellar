/*
Copyright 2022 The KCP Authors.

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

package plugin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"text/template"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/uuid"
	"github.com/martinlindhe/base36"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/edge-mc/pkg/cliplugins/base"
	kcpclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	"github.com/kcp-dev/kcp/pkg/cliplugins/helpers"
	"github.com/kcp-dev/logicalcluster/v3"
)

//go:embed *.yaml
var embeddedResources embed.FS

const (
	SyncerSecretConfigKey = "kubeconfig"
)

// EdgeSyncOptions contains options for configuring a SyncTarget and its corresponding syncer.
type EdgeSyncOptions struct {
	*base.Options

	// SyncerImage is the container image that should be used for the syncer.
	SyncerImage string
	// Replicas is the number of replicas to configure in the syncer's deployment.
	Replicas int
	// OutputFile is the path to a file where the YAML for the syncer should be written.
	OutputFile string
	// DownstreamNamespace is the name of the namespace in the physical cluster where the syncer deployment is created.
	DownstreamNamespace string
	// KCPNamespace is the name of the namespace in the kcp workspace where the service account is created for the
	// syncer.
	KCPNamespace string
	// QPS is the refill rate for the syncer client's rate limiter bucket (steady state requests per second).
	QPS float32
	// Burst is the maximum size for the syncer client's rate limiter bucket when idle.
	Burst int
	// SyncTargetName is the name of the SyncTarget in the kcp workspace.
	SyncTargetName string
	// SyncTargetLabels are the labels to be applied to the SyncTarget in the kcp workspace.
	SyncTargetLabels []string
}

// NewSyncOptions returns a new EdgeSyncOptions.
func NewEdgeSyncOptions(streams genericclioptions.IOStreams) *EdgeSyncOptions {
	return &EdgeSyncOptions{
		Options: base.NewOptions(streams),

		Replicas:     1,
		KCPNamespace: "default",
		QPS:          20,
		Burst:        30,
	}
}

// BindFlags binds fields EdgeSyncOptions as command line flags to cmd's flagset.
func (o *EdgeSyncOptions) BindFlags(cmd *cobra.Command) {
	o.Options.BindFlags(cmd)

	cmd.Flags().StringVar(&o.SyncerImage, "syncer-image", o.SyncerImage, "The edge-syncer image to use in the syncer's deployment YAML. Images are published at https://quay.io/repository/kcpedge/syncer")
	cmd.Flags().IntVar(&o.Replicas, "replicas", o.Replicas, "Number of replicas of the syncer deployment.")
	cmd.Flags().StringVar(&o.KCPNamespace, "kcp-namespace", o.KCPNamespace, "The name of the kcp namespace to create a service account in.")
	cmd.Flags().StringVarP(&o.OutputFile, "output-file", "o", o.OutputFile, "The manifest file to be created and applied to the physical cluster. Use - for stdout.")
	cmd.Flags().StringVarP(&o.DownstreamNamespace, "namespace", "n", o.DownstreamNamespace, "The namespace to create the syncer in the physical cluster. By default this is \"kcp-edge-syncer-<synctarget-name>-<uid>\".")
	cmd.Flags().Float32Var(&o.QPS, "qps", o.QPS, "QPS to use when talking to API servers.")
	cmd.Flags().IntVar(&o.Burst, "burst", o.Burst, "Burst to use when talking to API servers.")
	cmd.Flags().StringSliceVar(&o.SyncTargetLabels, "labels", o.SyncTargetLabels, "Labels to apply on the SyncTarget created in kcp, each label should be in the format of key=value.")
}

// Complete ensures all dynamically populated fields are initialized.
func (o *EdgeSyncOptions) Complete(args []string) error {
	if err := o.Options.Complete(); err != nil {
		return err
	}

	o.SyncTargetName = args[0]

	return nil
}

// Validate validates the EdgeSyncOptions are complete and usable.
func (o *EdgeSyncOptions) Validate() error {
	var errs []error

	if err := o.Options.Validate(); err != nil {
		errs = append(errs, err)
	}

	if o.SyncerImage == "" {
		errs = append(errs, errors.New("--syncer-image is required"))
	}

	if o.KCPNamespace == "" {
		errs = append(errs, errors.New("--kcp-namespace is required"))
	}

	if o.Replicas < 0 {
		errs = append(errs, errors.New("--replicas cannot be negative"))
	}
	if o.Replicas > 1 {
		// TODO: relax when we have leader-election in the syncer
		errs = append(errs, errors.New("only 0 and 1 are valid values for --replicas"))
	}

	if o.OutputFile == "" {
		errs = append(errs, errors.New("--output-file is required"))
	}

	for _, l := range o.SyncTargetLabels {
		if len(strings.Split(l, "=")) != 2 {
			errs = append(errs, fmt.Errorf("label '%s' is not in the format of key=value", l))
		}
	}

	return utilerrors.NewAggregate(errs)
}

// Run prepares a kcp workspace for use with a syncer and outputs the
// configuration required to deploy a syncer to the pcluster to stdout.
func (o *EdgeSyncOptions) Run(ctx context.Context) error {
	config, err := o.ClientConfig.ClientConfig()
	if err != nil {
		return err
	}

	var outputFile *os.File
	if o.OutputFile == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(o.OutputFile)
		if err != nil {
			return err
		}
		defer outputFile.Close()
	}

	labels := map[string]string{}
	for _, l := range o.SyncTargetLabels {
		parts := strings.Split(l, "=")
		if len(parts) != 2 {
			continue
		}
		labels[parts[0]] = parts[1]
	}

	token, syncerID, edgeSyncTarget, err := o.enableSyncerForWorkspace(ctx, config, o.SyncTargetName, o.KCPNamespace, labels)
	if err != nil {
		return err
	}

	configURL, _, err := helpers.ParseClusterURL(config.Host)
	if err != nil {
		return fmt.Errorf("current URL %q does not point to workspace", config.Host)
	}

	// Make sure the generated URL has the port specified correctly.
	if _, _, err = net.SplitHostPort(configURL.Host); err != nil {
		var addrErr *net.AddrError
		const missingPort = "missing port in address"
		if errors.As(err, &addrErr) && addrErr.Err == missingPort {
			if configURL.Scheme == "https" {
				configURL.Host = net.JoinHostPort(configURL.Host, "443")
			} else {
				configURL.Host = net.JoinHostPort(configURL.Host, "80")
			}
		} else {
			return fmt.Errorf("failed to parse host %q: %w", configURL.Host, err)
		}
	}

	if o.DownstreamNamespace == "" {
		o.DownstreamNamespace = syncerID
	}

	// Compose the syncer's upstream configuration server URL without any path. This is
	// required so long as the API importer and syncer expect to require cluster clients.
	//
	// TODO(marun) It's probably preferable that the syncer and importer are provided a
	// cluster configuration since they only operate against a single workspace.
	serverURL := configURL.Scheme + "://" + configURL.Host
	input := templateInputForEdge{
		ServerURL:    serverURL,
		CAData:       base64.StdEncoding.EncodeToString(config.CAData),
		Token:        token,
		KCPNamespace: o.KCPNamespace,
		Namespace:    o.DownstreamNamespace,

		SyncTargetPath: logicalcluster.From(edgeSyncTarget).Path().String(),
		SyncTarget:     o.SyncTargetName,
		SyncTargetUID:  string(edgeSyncTarget.UID),

		Image:    o.SyncerImage,
		Replicas: o.Replicas,
		QPS:      o.QPS,
		Burst:    o.Burst,
	}

	resources, err := renderEdgeSyncerResources(input, syncerID)
	if err != nil {
		return err
	}

	_, err = outputFile.Write(resources)
	if o.OutputFile != "-" {
		fmt.Fprintf(o.ErrOut, "\nWrote physical cluster manifest to %s for namespace %q. Use\n\n  KUBECONFIG=<pcluster-config> kubectl apply -f %q\n\nto apply it. "+
			"Use\n\n  KUBECONFIG=<pcluster-config> kubectl get deployment -n %q %s\n\nto verify the syncer pod is running.\n", o.OutputFile, o.DownstreamNamespace, o.OutputFile, o.DownstreamNamespace, syncerID)
	}
	return err
}

// getEdgeSyncerID returns a unique ID for a syncer derived from the name and its UID. It's
// a valid DNS segment and can be used as namespace or object names.
func getEdgeSyncerID(edgeSyncTarget *typeEdgeSyncTarget) string {
	syncerHash := sha256.Sum224([]byte(edgeSyncTarget.UID))
	base36hash := strings.ToLower(base36.EncodeBytes(syncerHash[:]))
	return fmt.Sprintf("kcp-edge-syncer-%s-%s", edgeSyncTarget.Name, base36hash[:8])
}

type typeEdgeSyncTarget struct {
	UID         types.UID
	Name        string
	Annotations map[string]string
}

func (o *typeEdgeSyncTarget) GetAnnotations() map[string]string {
	return o.Annotations
}

func (o *EdgeSyncOptions) applyEdgeSyncTarget(ctx context.Context, kcpClient kcpclient.Interface, edgeSyncTargetName string, labels map[string]string) (*typeEdgeSyncTarget, error) {
	logicalCluster, err := kcpClient.CoreV1alpha1().LogicalClusters().Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get default logical cluster %q: %w", edgeSyncTargetName, err)
	}
	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID %q: %w", edgeSyncTargetName, err)
	}
	edgeSyncTarget := typeEdgeSyncTarget{
		UID:         types.UID(uuid.String()),
		Name:        edgeSyncTargetName,
		Annotations: logicalCluster.Annotations,
	}
	return &edgeSyncTarget, nil
}

// enableSyncerForWorkspace creates a sync target with the given name and creates a service
// account for the syncer in the given namespace. The expectation is that the provided config is
// for a logical cluster (workspace). Returns the token the syncer will use to connect to kcp.
func (o *EdgeSyncOptions) enableSyncerForWorkspace(ctx context.Context, config *rest.Config, edgeSyncTargetName, namespace string, labels map[string]string) (saToken string, syncerID string, edgeSyncTarget *typeEdgeSyncTarget, err error) {
	kcpClient, err := kcpclient.NewForConfig(config)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create kcp client: %w", err)
	}

	edgeSyncTarget, err = o.applyEdgeSyncTarget(ctx, kcpClient, edgeSyncTargetName, labels)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to apply synctarget %q: %w", edgeSyncTargetName, err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	var syncConfig *unstructured.Unstructured
	if err := wait.PollImmediateInfiniteWithContext(ctx, time.Second*1, func(ctx context.Context) (bool, error) {
		syncConfig, err = createEdgeSyncConfig(ctx, config, edgeSyncTargetName)
		return err == nil, nil
	}); err != nil {
		return "", "", nil, fmt.Errorf("failed to get or create EdgeSyncConfig resource: %w", err)
	}
	syncerID = getEdgeSyncerID(edgeSyncTarget)
	syncTargetOwnerReferences := []metav1.OwnerReference{{
		APIVersion: syncConfig.GetAPIVersion(),
		Kind:       syncConfig.GetKind(),
		Name:       syncConfig.GetName(),
		UID:        syncConfig.GetUID(),
	}}
	sa, err := kubeClient.CoreV1().ServiceAccounts(namespace).Get(ctx, syncerID, metav1.GetOptions{})

	switch {
	case apierrors.IsNotFound(err):
		fmt.Fprintf(o.ErrOut, "Creating service account %q\n", syncerID)
		if sa, err = kubeClient.CoreV1().ServiceAccounts(namespace).Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:            syncerID,
				OwnerReferences: syncTargetOwnerReferences,
			},
		}, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return "", "", nil, fmt.Errorf("failed to create ServiceAccount %s|%s/%s: %w", edgeSyncTargetName, namespace, syncerID, err)
		}
	case err == nil:
		oldData, err := json.Marshal(corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: sa.OwnerReferences,
			},
		})
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to marshal old data for ServiceAccount %s|%s/%s: %w", edgeSyncTargetName, namespace, syncerID, err)
		}

		newData, err := json.Marshal(corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				UID:             sa.UID,
				ResourceVersion: sa.ResourceVersion,
				OwnerReferences: mergeOwnerReferenceForEdge(sa.ObjectMeta.OwnerReferences, syncTargetOwnerReferences),
			},
		})
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to marshal new data for ServiceAccount %s|%s/%s: %w", edgeSyncTargetName, namespace, syncerID, err)
		}

		patchBytes, err := jsonpatch.CreateMergePatch(oldData, newData)
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to create patch for ServiceAccount %s|%s/%s: %w", edgeSyncTargetName, namespace, syncerID, err)
		}

		fmt.Fprintf(o.ErrOut, "Updating service account %q.\n", syncerID)
		if sa, err = kubeClient.CoreV1().ServiceAccounts(namespace).Patch(ctx, sa.Name, types.MergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
			return "", "", nil, fmt.Errorf("failed to patch ServiceAccount %s|%s/%s: %w", edgeSyncTargetName, syncerID, namespace, err)
		}
	default:
		return "", "", nil, fmt.Errorf("failed to get the ServiceAccount %s|%s/%s: %w", edgeSyncTargetName, syncerID, namespace, err)
	}

	// Create a cluster role that provides the syncer the minimal permissions
	// required by KCP to manage the sync target, and by the syncer virtual
	// workspace to sync.
	rules := []rbacv1.PolicyRule{
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"*"},
			Resources: []string{"*"},
		},
		{
			Verbs:           []string{"access"},
			NonResourceURLs: []string{"/"},
		},
	}

	cr, err := kubeClient.RbacV1().ClusterRoles().Get(ctx,
		syncerID,
		metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		fmt.Fprintf(o.ErrOut, "Creating cluster role %q to give service account %q\n\n 1. write and sync access to the synctarget %q\n 2. write access to apiresourceimports.\n\n", syncerID, syncerID, syncerID)
		if _, err = kubeClient.RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:            syncerID,
				OwnerReferences: syncTargetOwnerReferences,
			},
			Rules: rules,
		}, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return "", "", nil, err
		}
	case err == nil:
		oldData, err := json.Marshal(rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: cr.OwnerReferences,
			},
			Rules: cr.Rules,
		})
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to marshal old data for ClusterRole %s|%s: %w", edgeSyncTargetName, syncerID, err)
		}

		newData, err := json.Marshal(rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				UID:             cr.UID,
				ResourceVersion: cr.ResourceVersion,
				OwnerReferences: mergeOwnerReferenceForEdge(cr.OwnerReferences, syncTargetOwnerReferences),
			},
			Rules: rules,
		})
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to marshal new data for ClusterRole %s|%s: %w", edgeSyncTargetName, syncerID, err)
		}

		patchBytes, err := jsonpatch.CreateMergePatch(oldData, newData)
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to create patch for ClusterRole %s|%s: %w", edgeSyncTargetName, syncerID, err)
		}

		fmt.Fprintf(o.ErrOut, "Updating cluster role %q with\n\n 1. write and sync access to the synctarget %q\n 2. write access to apiresourceimports.\n\n", syncerID, syncerID)
		if _, err = kubeClient.RbacV1().ClusterRoles().Patch(ctx, cr.Name, types.MergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
			return "", "", nil, fmt.Errorf("failed to patch ClusterRole %s|%s/%s: %w", edgeSyncTargetName, syncerID, namespace, err)
		}
	default:
		return "", "", nil, err
	}

	// Grant the service account the role created just above in the workspace
	subjects := []rbacv1.Subject{{
		Kind:      "ServiceAccount",
		Name:      syncerID,
		Namespace: namespace,
	}}
	roleRef := rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     syncerID,
		APIGroup: "rbac.authorization.k8s.io",
	}

	_, err = kubeClient.RbacV1().ClusterRoleBindings().Get(ctx,
		syncerID,
		metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return "", "", nil, err
	}
	if err == nil {
		if err := kubeClient.RbacV1().ClusterRoleBindings().Delete(ctx, syncerID, metav1.DeleteOptions{}); err != nil {
			return "", "", nil, err
		}
	}

	fmt.Fprintf(o.ErrOut, "Creating or updating cluster role binding %q to bind service account %q to cluster role %q.\n", syncerID, syncerID, syncerID)
	if _, err = kubeClient.RbacV1().ClusterRoleBindings().Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            syncerID,
			OwnerReferences: syncTargetOwnerReferences,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return "", "", nil, err
	}

	// Wait for the service account to be updated with the name of the token secret
	tokenSecretName := ""
	err = wait.PollImmediateWithContext(ctx, 100*time.Millisecond, 20*time.Second, func(ctx context.Context) (bool, error) {
		serviceAccount, err := kubeClient.CoreV1().ServiceAccounts(namespace).Get(ctx, sa.Name, metav1.GetOptions{})
		if err != nil {
			klog.FromContext(ctx).V(5).WithValues("err", err).Info("failed to retrieve ServiceAccount")
			return false, nil
		}
		if len(serviceAccount.Secrets) == 0 {
			return false, nil
		}
		tokenSecretName = serviceAccount.Secrets[0].Name
		return true, nil
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("timed out waiting for token secret name to be set on ServiceAccount %s/%s", namespace, sa.Name)
	}

	// Retrieve the token that the syncer will use to authenticate to kcp
	tokenSecret, err := kubeClient.CoreV1().Secrets(namespace).Get(ctx, tokenSecretName, metav1.GetOptions{})
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to retrieve Secret: %w", err)
	}
	saTokenBytes := tokenSecret.Data["token"]
	if len(saTokenBytes) == 0 {
		return "", "", nil, fmt.Errorf("token secret %s/%s is missing a value for `token`", namespace, tokenSecretName)
	}

	return string(saTokenBytes), syncerID, edgeSyncTarget, nil
}

// mergeOwnerReferenceForEdge: merge a slice of ownerReference with a given ownerReferences.
func mergeOwnerReferenceForEdge(ownerReferences, newOwnerReferences []metav1.OwnerReference) []metav1.OwnerReference {
	var merged []metav1.OwnerReference

	merged = append(merged, ownerReferences...)

	for _, ownerReference := range newOwnerReferences {
		found := false
		for _, mergedOwnerReference := range merged {
			if mergedOwnerReference.UID == ownerReference.UID {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, ownerReference)
		}
	}

	return merged
}

// templateInputForEdge represents the external input required to render the resources to
// deploy the syncer to a pcluster.
type templateInputForEdge struct {
	// ServerURL is the logical cluster url the syncer configuration will use
	ServerURL string
	// CAData holds the PEM-encoded bytes of the ca certificate(s) a syncer will use to validate
	// kcp's serving certificate
	CAData string
	// Token is the service account token used to authenticate a syncer for access to a workspace
	Token string
	// KCPNamespace is the name of the kcp namespace of the syncer's service account
	KCPNamespace string
	// Namespace is the name of the syncer namespace on the pcluster
	Namespace string
	// SyncTargetPath is the qualified kcp logical cluster name the syncer will sync from
	SyncTargetPath string
	// SyncTarget is the name of the sync target the syncer will use to
	// communicate its status and read configuration from
	SyncTarget string
	// SyncTargetUID is the UID of the sync target the syncer will use to
	// communicate its status and read configuration from. This information is used by the
	// Syncer in order to avoid a conflict when a synctarget gets deleted and another one is
	// created with the same name.
	SyncTargetUID string
	// Image is the name of the container image that the syncer deployment will use
	Image string
	// Replicas is the number of syncer pods to run (should be 0 or 1).
	Replicas int
	// QPS is the qps the syncer uses when talking to an apiserver.
	QPS float32
	// Burst is the burst the syncer uses when talking to an apiserver.
	Burst int
}

// templateArgsForEdge represents the full set of arguments required to render the resources
// required to deploy the syncer.
type templateArgsForEdge struct {
	templateInputForEdge
	// ServiceAccount is the name of the service account to create in the syncer
	// namespace on the pcluster.
	ServiceAccount string
	// ClusterRole is the name of the cluster role to create for the syncer on the
	// pcluster.
	ClusterRole string
	// ClusterRoleBinding is the name of the cluster role binding to create for the
	// syncer on the pcluster.
	ClusterRoleBinding string
	// GroupMappings is the mapping of api group to resources that will be used to
	// define the cluster role rules for the syncer in the pcluster. The syncer will be
	// granted full permissions for the resources it will synchronize.
	GroupMappings []groupMappingForEdge
	// Secret is the name of the secret that will contain the kubeconfig the syncer
	// will use to connect to the kcp logical cluster (workspace) that it will
	// synchronize from.
	Secret string
	// Key in the syncer secret for the kcp logical cluster kubconfig.
	SecretConfigKey string
	// Deployment is the name of the deployment that will run the syncer in the
	// pcluster.
	Deployment string
	// DeploymentApp is the label value that the syncer's deployment will select its
	// pods with.
	DeploymentApp string
}

// renderEdgeSyncerResources renders the resources required to deploy a syncer to a pcluster.
func renderEdgeSyncerResources(input templateInputForEdge, syncerID string) ([]byte, error) {

	tmplArgs := templateArgsForEdge{
		templateInputForEdge: input,
		ServiceAccount:       syncerID,
		ClusterRole:          syncerID,
		ClusterRoleBinding:   syncerID,
		GroupMappings:        []groupMappingForEdge{},
		Secret:               syncerID,
		SecretConfigKey:      SyncerSecretConfigKey,
		Deployment:           syncerID,
		DeploymentApp:        syncerID,
	}

	syncerTemplate, err := embeddedResources.ReadFile("edge-syncer.yaml")
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("syncerTemplate").Parse(string(syncerTemplate))
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, tmplArgs)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// groupMappingForEdge associates an api group to the resources in that group.
type groupMappingForEdge struct {
	APIGroup  string
	Resources []string
}

func createEdgeSyncConfig(ctx context.Context, cfg *rest.Config, edgeSyncTargetName string) (*unstructured.Unstructured, error) {
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client :%w", err)
	}
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(cfg)
	gk := schema.GroupKind{
		Group: "edge.kcp.io",
		Kind:  "EdgeSyncConfig",
	}
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get API group resources :%w", err)
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := restMapper.RESTMapping(gk, "v1alpha1")
	if err != nil || mapping == nil {
		return nil, fmt.Errorf("failed to get resource mapping :%w", err)
	}
	cr, err := dynamicClient.Resource(mapping.Resource).Get(ctx, edgeSyncTargetName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		cr = &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": gk.Group + "/v1alpha1",
			"kind":       gk.Kind,
			"metadata": map[string]interface{}{
				"name": edgeSyncTargetName,
			},
			"spec": map[string]interface{}{},
		}}
		cr, err := dynamicClient.Resource(mapping.Resource).Create(ctx, cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create EdgeSyncConfig :%w", err)
		}
		return cr, nil
	} else {
		return cr, nil
	}
}
