# Resource Dependency Operator - Complete Algorithm & Implementation logic

## Problem Statement

KubeStellar faces complex dependency management issues:
- Multiple initContainers with security vulnerabilities
- Complex RBAC requirements for kubectl containers
- No native "wait for resource" capability in Helm
- Need for priority-based resource creation
- Conditional resource creation based on states

## Solution: Dependency Management Operator

A Kubernetes-native operator that handles resource dependencies, priority-based creation, and state waiting - eliminating the need for initContainers and kubectl images.

---

## Core Algorithm

### 1. **Resource Definition & Priority Algorithm**

```yaml
apiVersion: tools.kubestellar.io/v1
kind: ResourceDependency
metadata:
  name: kubestellar-bootstrap
spec:
  # Priority-based execution (lower number = higher priority)
  priority: 1
  
  # Dependencies to wait for
  dependencies:
    - resource: "namespace/kflex-system"
      timeout: "2m"
      priority: 1
    - resource: "deployment/kflex-controller"
      namespace: "kflex-system"
      state: "Ready"
      minReplicas: 1
      timeout: "5m"
      priority: 2
    - resource: "crd/controlplanes.kubeflex.kubestellar.io"
      timeout: "1m"
      priority: 1

  # Actions to execute when dependencies are met
  actions:
    - priority: 1
      type: "create"
      resource:
        apiVersion: "v1"
        kind: "ConfigMap"
        metadata:
          name: "bootstrap-config"
          namespace: "kflex-system"
        data:
          phase: "dependencies-ready"
          timestamp: "{{ .Now }}"

    - priority: 2
      type: "template"
      template: |
        apiVersion: v1
        kind: Secret
        metadata:
          name: kubestellar-bootstrap
          namespace: {{ .Dependencies.namespace_kflex_system.Name }}
        data:
          config: {{ .Values.bootstrapConfig | b64enc }}

  # Failure handling
  onFailure:
    - type: "event"
      message: "KubeStellar bootstrap dependencies failed"
    - type: "status"
      phase: "Failed"
```

### 2. **Core Reconciliation Algorithm**

```
┌─────────────────────────────────────────┐
│          RECONCILIATION LOOP            │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  1. FETCH ResourceDependency Object     │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  2. SORT Dependencies by Priority       │
│     (Lower number = Higher priority)    │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  3. CHECK Dependencies in Order         │
│     For each dependency:                │
│     ├─ Check if resource exists         │
│     ├─ Validate required state         │
│     ├─ Check timeout constraints        │
│     └─ Mark as Ready/NotReady/Failed    │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  4. EVALUATE Overall Status             │
│     ├─ All Ready → Execute Actions      │
│     ├─ Some Pending → Wait & Requeue    │
│     └─ Any Timeout → Execute onFailure  │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  5. EXECUTE Actions (if ready)          │
│     Sort actions by priority            │
│     Execute in priority order           │
│     Update status for each action       │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  6. UPDATE Status & Requeue             │
│     ├─ Update ResourceDependency status │
│     ├─ Emit Kubernetes events           │
│     └─ Schedule next reconciliation     │
└─────────────────────────────────────────┘
```

### 3. **Dependency Checking Algorithm**

```go
func (r *ResourceDependencyReconciler) checkDependency(ctx context.Context, dep Dependency) DependencyStatus {
    // Parse resource reference
    gvk, name, namespace := parseResourceRef(dep.Resource)
    
    // Apply timeout constraint
    if time.Since(dep.StartTime) > dep.Timeout {
        return DependencyStatus{
            State: "TimedOut",
            Message: fmt.Sprintf("Dependency %s timed out after %v", dep.Resource, dep.Timeout),
            LastChecked: time.Now(),
        }
    }
    
    // Check if resource exists
    obj := &unstructured.Unstructured{}
    obj.SetGroupVersionKind(gvk)
    
    err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, obj)
    if err != nil {
        if apierrors.IsNotFound(err) {
            return DependencyStatus{
                State: "NotFound",
                Message: fmt.Sprintf("Resource %s not found", dep.Resource),
                LastChecked: time.Now(),
            }
        }
        return DependencyStatus{State: "Error", Message: err.Error()}
    }
    
    // Check specific state requirements
    return r.validateResourceState(obj, dep)
}

func (r *ResourceDependencyReconciler) validateResourceState(obj *unstructured.Unstructured, dep Dependency) DependencyStatus {
    switch obj.GetKind() {
    case "Deployment":
        return r.checkDeploymentReady(obj, dep)
    case "StatefulSet":
        return r.checkStatefulSetReady(obj, dep)
    case "Pod":
        return r.checkPodReady(obj, dep)
    case "Service":
        return r.checkServiceReady(obj, dep)
    case "CustomResourceDefinition":
        return r.checkCRDReady(obj, dep)
    case "Namespace":
        return r.checkNamespaceReady(obj, dep)
    default:
        // For unknown resources, just check existence
        return DependencyStatus{
            State: "Ready",
            Message: fmt.Sprintf("Resource %s exists", dep.Resource),
            LastChecked: time.Now(),
        }
    }
}
```

### 4. **Priority-Based Action Execution Algorithm**

```go
func (r *ResourceDependencyReconciler) executeActions(ctx context.Context, actions []Action, depStatus map[string]DependencyStatus) error {
    // Sort actions by priority (ascending - lower numbers first)
    sort.Slice(actions, func(i, j int) bool {
        return actions[i].Priority < actions[j].Priority
    })
    
    // Execute actions in priority order
    for _, action := range actions {
        if err := r.executeAction(ctx, action, depStatus); err != nil {
            // Log error but continue with next action
            log.Error(err, "Failed to execute action", "priority", action.Priority, "type", action.Type)
            
            // Decide whether to fail fast or continue
            if action.FailFast {
                return fmt.Errorf("action failed with fail-fast enabled: %w", err)
            }
        }
        
        // Optional: Add delay between priority groups
        if action.DelayAfter != nil {
            time.Sleep(action.DelayAfter.Duration)
        }
    }
    
    return nil
}

func (r *ResourceDependencyReconciler) executeAction(ctx context.Context, action Action, depStatus map[string]DependencyStatus) error {
    switch action.Type {
    case "create":
        return r.createResource(ctx, action.Resource, depStatus)
    case "template":
        return r.createFromTemplate(ctx, action.Template, depStatus)
    case "patch":
        return r.patchResource(ctx, action.Patch, depStatus)
    case "delete":
        return r.deleteResource(ctx, action.Delete, depStatus)
    case "wait":
        return r.waitForDuration(ctx, action.Wait)
    case "exec":
        return r.executeCommand(ctx, action.Exec, depStatus)
    default:
        return fmt.Errorf("unknown action type: %s", action.Type)
    }
}
```

---

## Implementation Strategy

### Phase 1: Core Operator Structure

```bash
# Initialize project
kubebuilder init --domain kubestellar.io --repo github.com/yourusername/resource-dependency-operator

# Create main API
kubebuilder create api --group tools --version v1 --kind ResourceDependency --resource --controller

# Create supplementary APIs for advanced features
kubebuilder create api --group tools --version v1 --kind DependencyGroup --resource --controller
```

### Phase 2: Custom Resources Definition

#### **Main Resource: ResourceDependency**

```go
type ResourceDependencySpec struct {
    // Priority for this dependency group (lower = higher priority)
    Priority int `json:"priority,omitempty"`
    
    // Dependencies to wait for
    Dependencies []Dependency `json:"dependencies"`
    
    // Actions to execute when dependencies are met
    Actions []Action `json:"actions,omitempty"`
    
    // Failure handling
    OnFailure []FailureAction `json:"onFailure,omitempty"`
    
    // Global timeout for all dependencies
    GlobalTimeout *metav1.Duration `json:"globalTimeout,omitempty"`
    
    // Retry configuration
    RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
}

type Dependency struct {
    // Resource reference (kind/name or kind/name/namespace)
    Resource string `json:"resource"`
    
    // Namespace (optional)
    Namespace string `json:"namespace,omitempty"`
    
    // State to wait for
    State string `json:"state,omitempty"`
    
    // Priority within this dependency group
    Priority int `json:"priority,omitempty"`
    
    // Timeout for this specific dependency
    Timeout *metav1.Duration `json:"timeout,omitempty"`
    
    // Custom conditions
    Conditions []DependencyCondition `json:"conditions,omitempty"`
    
    // Minimum replicas (for Deployments, StatefulSets)
    MinReplicas *int32 `json:"minReplicas,omitempty"`
    
    // Labels selector for additional filtering
    LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

type DependencyCondition struct {
    Type     string `json:"type"`     // "Ready", "Available", "Custom"
    Status   string `json:"status"`   // "True", "False", "Unknown"
    Reason   string `json:"reason,omitempty"`
    JSONPath string `json:"jsonPath,omitempty"` // For custom conditions
    Value    string `json:"value,omitempty"`    // Expected value
}

type Action struct {
    Priority int    `json:"priority,omitempty"`
    Type     string `json:"type"` // "create", "template", "patch", "delete", "wait", "exec"
    
    // Different action types
    Resource *unstructured.Unstructured `json:"resource,omitempty"`
    Template string                     `json:"template,omitempty"`
    Patch    *PatchAction              `json:"patch,omitempty"`
    Delete   *DeleteAction             `json:"delete,omitempty"`
    Wait     *metav1.Duration          `json:"wait,omitempty"`
    Exec     *ExecAction               `json:"exec,omitempty"`
    
    // Action modifiers
    FailFast     bool                `json:"failFast,omitempty"`
    DelayAfter   *metav1.Duration    `json:"delayAfter,omitempty"`
    Condition    *ActionCondition    `json:"condition,omitempty"`
}
```

### Phase 3: Controller Implementation

#### **Main Reconciliation Logic**

```go
func (r *ResourceDependencyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    
    // 1. Fetch ResourceDependency
    var resDep toolsv1.ResourceDependency
    if err := r.Get(ctx, req.NamespacedName, &resDep); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    // 2. Check if already completed
    if resDep.Status.Phase == "Completed" {
        return ctrl.Result{}, nil
    }
    
    // 3. Sort dependencies by priority
    deps := make([]toolsv1.Dependency, len(resDep.Spec.Dependencies))
    copy(deps, resDep.Spec.Dependencies)
    sort.Slice(deps, func(i, j int) bool {
        return deps[i].Priority < deps[j].Priority
    })
    
    // 4. Check dependencies in priority order
    depStatus := make(map[string]DependencyStatus)
    allReady := true
    anyFailed := false
    
    for _, dep := range deps {
        status := r.checkDependency(ctx, dep)
        depStatus[dep.Resource] = status
        
        switch status.State {
        case "Ready":
            continue
        case "TimedOut", "Failed":
            anyFailed = true
            allReady = false
        default:
            allReady = false
        }
        
        // If high priority dependency isn't ready, wait before checking lower priority ones
        if dep.Priority < 5 && status.State != "Ready" {
            break
        }
    }
    
    // 5. Handle different states
    if anyFailed {
        return r.handleFailure(ctx, &resDep, depStatus)
    }
    
    if !allReady {
        r.updateStatus(ctx, &resDep, "Waiting", "Dependencies not yet ready", depStatus)
        return ctrl.Result{RequeueAfter: r.calculateBackoff(resDep.Status.RetryCount)}, nil
    }
    
    // 6. Execute actions
    if err := r.executeActions(ctx, resDep.Spec.Actions, depStatus); err != nil {
        r.updateStatus(ctx, &resDep, "ActionFailed", err.Error(), depStatus)
        return ctrl.Result{RequeueAfter: time.Minute}, err
    }
    
    // 7. Mark as completed
    r.updateStatus(ctx, &resDep, "Completed", "All dependencies met and actions executed", depStatus)
    
    return ctrl.Result{}, nil
}
```

### Phase 4: KubeStellar Integration Examples

#### **Example 1: KubeFlex Bootstrap**

```yaml
apiVersion: tools.kubestellar.io/v1
kind: ResourceDependency
metadata:
  name: kubeflex-bootstrap
  namespace: kubeflex-system
spec:
  priority: 1
  globalTimeout: "10m"
  
  dependencies:
    # Priority 1: Core infrastructure
    - resource: "namespace/kubeflex-system"
      priority: 1
      timeout: "2m"
    
    - resource: "crd/controlplanes.kubeflex.kubestellar.io"
      priority: 1
      timeout: "2m"
    
    # Priority 2: Controllers
    - resource: "deployment/kubeflex-controller-manager"
      namespace: "kubeflex-system"
      state: "Ready"
      minReplicas: 1
      priority: 2
      timeout: "5m"
      conditions:
        - type: "Available"
          status: "True"
  
  actions:
    - priority: 1
      type: "create"
      resource:
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: kubeflex-ready
          namespace: kubeflex-system
        data:
          status: "ready"
          phase: "bootstrap-complete"
    
    - priority: 2
      type: "template"
      template: |
        apiVersion: kubeflex.kubestellar.io/v1alpha1
        kind: ControlPlane
        metadata:
          name: {{ .Values.itsName | default "its1" }}
        spec:
          type: vcluster
```

#### **Example 2: WDS Setup with ITS Dependency**

```yaml
apiVersion: tools.kubestellar.io/v1
kind: ResourceDependency
metadata:
  name: wds-setup
spec:
  priority: 2
  
  dependencies:
    # Wait for ITS to be ready
    - resource: "controlplane/its1"
      state: "Ready"
      priority: 1
      timeout: "8m"
      conditions:
        - type: "Ready"
          status: "True"
        - jsonPath: ".status.phase"
          value: "Running"
    
    # Wait for OCM in ITS
    - resource: "deployment/cluster-manager"
      namespace: "its1"
      state: "Ready"
      priority: 2
      timeout: "5m"
  
  actions:
    - priority: 1
      type: "template"
      template: |
        apiVersion: kubeflex.kubestellar.io/v1alpha1
        kind: ControlPlane
        metadata:
          name: {{ .Values.wdsName | default "wds1" }}
        spec:
          type: k8s
    
    - priority: 2
      type: "wait"
      wait: "30s"  # Wait for WDS to initialize
    
    - priority: 3
      type: "template"
      template: |
        apiVersion: v1
        kind: Secret
        metadata:
          name: wds-its-connection
          namespace: {{ .Values.wdsName | default "wds1" }}
        data:
          its-endpoint: {{ .Dependencies.controlplane_its1.Status.APIEndpoint | b64enc }}
```

### Phase 5: Advanced Features

#### **Conditional Actions**

```yaml
actions:
  - priority: 1
    type: "create"
    condition:
      type: "dependency"
      dependency: "deployment/postgres"
      state: "Ready"
    resource:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: app-config
      data:
        db_ready: "true"
```

#### **Dependency Groups**

```yaml
apiVersion: tools.kubestellar.io/v1
kind: DependencyGroup
metadata:
  name: kubestellar-full-stack
spec:
  groups:
    - name: "infrastructure"
      priority: 1
      dependencies:
        - "namespace/kubeflex-system"
        - "crd/controlplanes.kubeflex.kubestellar.io"
    
    - name: "control-planes"
      priority: 2
      dependsOn: ["infrastructure"]
      dependencies:
        - "controlplane/its1"
        - "controlplane/wds1"
    
    - name: "applications"
      priority: 3
      dependsOn: ["control-planes"]
      dependencies:
        - "deployment/kubestellar-controller"
```

---

## Benefits for KubeStellar

### 1. **Eliminates Current Pain Points**
- ❌ **No more initContainers**
- ❌ **No kubectl container images**
- ❌ **No complex RBAC for hooks**
- ❌ **No security vulnerabilities**

### 2. **Kubernetes-Native Solution**
- ✅ **Declarative resource management**
- ✅ **Built-in retry and backoff**
- ✅ **Proper status reporting**
- ✅ **Event-driven reconciliation**

### 3. **Advanced Capabilities**
- ✅ **Priority-based execution**
- ✅ **Complex dependency chains**
- ✅ **Template-based resource creation**
- ✅ **Conditional actions**
- ✅ **Timeout and failure handling**

### 4. **Reusability**
- ✅ **Works across all KubeStellar components**
- ✅ **Can be used by other projects**
- ✅ **Reduces code duplication**

This operator will transform KubeStellar's deployment model from script-based to fully declarative, Kubernetes-native dependency management.