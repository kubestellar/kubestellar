export interface Plugin {
  id: string;
  name: string;
  slug: string;
  tagline: string;
  description: string;
  longDescription: string;
  icon: string;
  category: string;
  pricing: {
    type: "free" | "monthly" | "one-time";
    amount?: number;
  };
  author: string;
  downloads: number;
  rating: number;
  version: string;
  features: string[];
  requirements: string[];
  compatibility: string[];
  screenshots: string[];
  documentation: string;
  github?: string;
  website?: string;
  tags: string[];
}

export const plugins: Plugin[] = [
  {
    id: "1",
    name: "KubeCtl Multi",
    slug: "kubectl-multi",
    tagline: "Execute kubectl commands across multiple clusters simultaneously",
    description:
      "Streamline your multi-cluster operations with KubeCtl Multi. Execute commands across all your KubeStellar-managed clusters with a single command.",
    longDescription: `KubeCtl Multi is the ultimate productivity tool for managing multiple Kubernetes clusters through KubeStellar. Built specifically for the KubeStellar ecosystem, it allows you to execute kubectl commands across all your managed clusters simultaneously, saving hours of manual work.

With intelligent context switching and parallel execution, KubeCtl Multi ensures your operations are fast, reliable, and consistent across your entire infrastructure. Whether you're deploying applications, updating configurations, or troubleshooting issues, KubeCtl Multi makes it seamless.

Perfect for DevOps teams managing edge deployments, hybrid clouds, or geographically distributed clusters.`,
    icon: "🎯",
    category: "CLI Tools",
    pricing: {
      type: "free",
    },
    author: "KubeStellar Core Team",
    downloads: 15420,
    rating: 4.8,
    version: "2.1.0",
    features: [
      "Execute commands across all clusters simultaneously",
      "Intelligent context switching and management",
      "Parallel execution with customizable concurrency",
      "Built-in safety checks and rollback mechanisms",
      "Interactive mode for cluster selection",
      "Export results to JSON, YAML, or CSV formats",
      "Integration with KubeStellar's cluster inventory",
      "Real-time progress tracking and logging",
    ],
    requirements: [
      "KubeStellar v0.21.0 or higher",
      "kubectl v1.26+",
      "Go 1.20+ (for building from source)",
    ],
    compatibility: ["Linux", "macOS", "Windows"],
    screenshots: [],
    documentation: "https://docs.kubestellar.io/plugins/kubectl-multi",
    github: "https://github.com/kubestellar/kubectl-multi",
    tags: ["kubectl", "multi-cluster", "cli", "automation", "productivity"],
  },
  {
    id: "2",
    name: "Galaxy Sync Pro",
    slug: "galaxy-sync-pro",
    tagline: "Advanced synchronization engine for KubeStellar clusters",
    description:
      "Galaxy Sync Pro provides enterprise-grade synchronization capabilities with advanced conflict resolution, real-time monitoring, and automated healing.",
    longDescription: `Galaxy Sync Pro takes KubeStellar's synchronization capabilities to the next level with advanced features designed for enterprise workloads. Get real-time visibility into sync operations, automatic conflict resolution, and intelligent retry mechanisms.

Monitor sync health across your entire cluster fleet with beautiful dashboards, set up custom alerts for sync failures, and leverage machine learning-powered optimization to improve sync performance over time.

Trusted by Fortune 500 companies to keep their distributed applications in perfect harmony.`,
    icon: "🔄",
    category: "Synchronization",
    pricing: {
      type: "monthly",
      amount: 99,
    },
    author: "StellarSync Inc.",
    downloads: 8932,
    rating: 4.9,
    version: "3.5.2",
    features: [
      "Real-time sync monitoring dashboard",
      "Advanced conflict resolution algorithms",
      "Automated healing and retry mechanisms",
      "Custom sync policies and rules",
      "Performance analytics and optimization",
      "Multi-region replication",
      "Compliance and audit logging",
      "24/7 enterprise support",
    ],
    requirements: [
      "KubeStellar v0.20.0 or higher",
      "Minimum 3 managed clusters",
      "Valid license key",
    ],
    compatibility: ["Linux", "macOS"],
    screenshots: [],
    documentation: "https://docs.galaxysync.io",
    website: "https://galaxysync.io",
    tags: [
      "sync",
      "enterprise",
      "monitoring",
      "automation",
      "premium",
    ],
  },
  {
    id: "3",
    name: "Edge Guardian",
    slug: "edge-guardian",
    tagline: "Security and compliance for edge deployments",
    description:
      "Ensure your edge clusters meet security and compliance requirements with automated scanning, policy enforcement, and real-time threat detection.",
    longDescription: `Edge Guardian is your security command center for KubeStellar-managed edge deployments. With the explosion of edge computing, security can't be an afterthought. Edge Guardian provides comprehensive security scanning, policy enforcement, and threat detection specifically designed for distributed edge environments.

Get instant visibility into security postures across all edge locations, automatically enforce compliance policies, and detect anomalies before they become incidents. Built-in integrations with popular SIEM tools ensure your security team has all the information they need.

SOC 2, ISO 27001, and HIPAA compliant out of the box.`,
    icon: "🛡️",
    category: "Security",
    pricing: {
      type: "monthly",
      amount: 149,
    },
    author: "EdgeSecure Labs",
    downloads: 5621,
    rating: 4.7,
    version: "1.8.0",
    features: [
      "Continuous security scanning",
      "Policy-as-code enforcement",
      "Real-time threat detection",
      "Compliance reporting (SOC 2, ISO 27001, HIPAA)",
      "Vulnerability management",
      "Network policy validation",
      "Secret scanning and rotation",
      "Integration with popular SIEM tools",
    ],
    requirements: [
      "KubeStellar v0.21.0 or higher",
      "Kubernetes 1.24+",
      "Valid subscription",
    ],
    compatibility: ["Linux", "macOS"],
    screenshots: [],
    documentation: "https://docs.edgeguardian.io",
    website: "https://edgeguardian.io",
    tags: ["security", "compliance", "edge", "scanning", "premium"],
  },
  {
    id: "4",
    name: "Stellar Insights",
    slug: "stellar-insights",
    tagline: "Advanced analytics and observability for multi-cluster deployments",
    description:
      "Gain deep insights into your KubeStellar deployments with advanced metrics, tracing, and AI-powered anomaly detection.",
    longDescription: `Stellar Insights transforms your KubeStellar deployment into a fully observable system. Get deep visibility into every aspect of your multi-cluster infrastructure with advanced metrics collection, distributed tracing, and AI-powered analytics.

Our machine learning algorithms detect anomalies before they impact your applications, predict resource needs, and suggest optimizations. Beautiful, customizable dashboards give your team the insights they need at a glance.

Reduce MTTR by 70% and proactively prevent 90% of incidents with Stellar Insights.`,
    icon: "📊",
    category: "Observability",
    pricing: {
      type: "one-time",
      amount: 299,
    },
    author: "ObservaStar",
    downloads: 12453,
    rating: 4.9,
    version: "2.3.1",
    features: [
      "Real-time metrics and dashboards",
      "Distributed tracing across clusters",
      "AI-powered anomaly detection",
      "Predictive analytics and forecasting",
      "Custom alert rules and notifications",
      "Log aggregation and analysis",
      "Cost optimization recommendations",
      "API for custom integrations",
    ],
    requirements: [
      "KubeStellar v0.19.0 or higher",
      "Prometheus operator",
      "Grafana (optional)",
    ],
    compatibility: ["Linux", "macOS", "Windows"],
    screenshots: [],
    documentation: "https://docs.stellarinsights.io",
    website: "https://stellarinsights.io",
    tags: [
      "observability",
      "metrics",
      "analytics",
      "monitoring",
      "ai",
    ],
  },
  {
    id: "5",
    name: "Cluster Navigator",
    slug: "cluster-navigator",
    tagline: "Visual cluster topology and resource mapping",
    description:
      "Navigate your multi-cluster topology with an interactive visual interface. Understand relationships, dependencies, and resource distribution at a glance.",
    longDescription: `Cluster Navigator brings your KubeStellar infrastructure to life with stunning visual representations. See your entire cluster topology, understand service dependencies, track resource distribution, and identify bottlenecks with an intuitive, interactive interface.

Built for both developers and operators, Cluster Navigator makes complex multi-cluster architectures easy to understand. Zoom in on specific namespaces, filter by labels, and trace request paths across clusters with just a few clicks.

Free and open source, trusted by over 15,000 KubeStellar users worldwide.`,
    icon: "🗺️",
    category: "Visualization",
    pricing: {
      type: "free",
    },
    author: "KubeStellar Community",
    downloads: 18765,
    rating: 4.6,
    version: "1.5.3",
    features: [
      "Interactive cluster topology visualization",
      "Service dependency mapping",
      "Resource distribution heatmaps",
      "Real-time status indicators",
      "Search and filter capabilities",
      "Export to PNG, SVG, or PDF",
      "Dark and light themes",
      "Responsive web interface",
    ],
    requirements: [
      "KubeStellar v0.20.0 or higher",
      "Modern web browser",
      "Network access to clusters",
    ],
    compatibility: ["Web-based"],
    screenshots: [],
    documentation: "https://docs.kubestellar.io/plugins/cluster-navigator",
    github: "https://github.com/kubestellar/cluster-navigator",
    tags: [
      "visualization",
      "topology",
      "ui",
      "free",
      "open-source",
    ],
  },
  {
    id: "6",
    name: "Config Validator",
    slug: "config-validator",
    tagline: "Validate KubeStellar configurations before deployment",
    description:
      "Catch configuration errors before they cause problems. Validate your BindingPolicies, Workspaces, and custom resources with comprehensive rule sets.",
    longDescription: `Config Validator is your safety net for KubeStellar configurations. Before deploying changes to your production clusters, validate them against comprehensive rule sets, best practices, and your own custom policies.

Catch typos, misconfigurations, security issues, and compliance violations in seconds. Integrates seamlessly into your CI/CD pipelines and provides clear, actionable error messages that help you fix issues fast.

Prevent 99% of configuration-related incidents with Config Validator.`,
    icon: "✅",
    category: "Development Tools",
    pricing: {
      type: "free",
    },
    author: "KubeStellar Core Team",
    downloads: 11234,
    rating: 4.7,
    version: "1.2.0",
    features: [
      "Comprehensive validation rules",
      "Custom policy support",
      "CI/CD pipeline integration",
      "Detailed error reporting",
      "Best practice recommendations",
      "Schema validation",
      "Dry-run simulation",
      "VS Code extension available",
    ],
    requirements: [
      "KubeStellar v0.18.0 or higher",
      "kubectl v1.25+",
    ],
    compatibility: ["Linux", "macOS", "Windows"],
    screenshots: [],
    documentation: "https://docs.kubestellar.io/plugins/config-validator",
    github: "https://github.com/kubestellar/config-validator",
    tags: [
      "validation",
      "ci-cd",
      "configuration",
      "development",
      "free",
    ],
  },
  {
    id: "7",
    name: "Disaster Recovery Suite",
    slug: "disaster-recovery-suite",
    tagline: "Complete backup and disaster recovery solution",
    description:
      "Protect your KubeStellar deployments with automated backups, point-in-time recovery, and multi-region disaster recovery capabilities.",
    longDescription: `When disaster strikes, Disaster Recovery Suite ensures your KubeStellar infrastructure can be restored quickly and completely. Automated backups of your cluster state, configurations, and data happen continuously in the background.

Restore to any point in time, replicate backups across regions for true disaster recovery, and test your recovery procedures without impacting production. Meet your RTO and RPO requirements with confidence.

Enterprise-grade disaster recovery that just works.`,
    icon: "💾",
    category: "Backup & Recovery",
    pricing: {
      type: "monthly",
      amount: 199,
    },
    author: "ResilienStar Inc.",
    downloads: 6847,
    rating: 4.8,
    version: "2.0.5",
    features: [
      "Automated continuous backups",
      "Point-in-time recovery",
      "Multi-region replication",
      "Backup encryption at rest and in transit",
      "Recovery testing environment",
      "RPO under 5 minutes",
      "RTO under 30 minutes",
      "Compliance reporting",
    ],
    requirements: [
      "KubeStellar v0.20.0 or higher",
      "S3-compatible storage",
      "Valid license",
    ],
    compatibility: ["Linux", "macOS"],
    screenshots: [],
    documentation: "https://docs.resilienstar.io",
    website: "https://resilienstar.io",
    tags: [
      "backup",
      "disaster-recovery",
      "business-continuity",
      "enterprise",
      "premium",
    ],
  },
  {
    id: "8",
    name: "Auto Scaler Pro",
    slug: "auto-scaler-pro",
    tagline: "Intelligent autoscaling for distributed workloads",
    description:
      "Optimize resource utilization across your cluster fleet with AI-powered autoscaling that understands your workload patterns and business requirements.",
    longDescription: `Auto Scaler Pro uses machine learning to understand your workload patterns and automatically scale resources across your KubeStellar-managed clusters. Unlike traditional autoscalers, it considers cross-cluster dependencies, geographic distribution, and your specific business requirements.

Save up to 60% on infrastructure costs while maintaining performance and reliability. Predictive scaling ensures resources are ready before demand spikes, and intelligent placement ensures workloads run in the most cost-effective locations.

Smart scaling for the distributed cloud era.`,
    icon: "📈",
    category: "Resource Management",
    pricing: {
      type: "monthly",
      amount: 79,
    },
    author: "ScaleStar Technologies",
    downloads: 9521,
    rating: 4.6,
    version: "1.9.2",
    features: [
      "AI-powered predictive scaling",
      "Cross-cluster resource optimization",
      "Cost-aware workload placement",
      "Custom scaling policies",
      "Integration with cluster autoscalers",
      "Real-time cost analytics",
      "Schedule-based scaling",
      "Multi-metric decision making",
    ],
    requirements: [
      "KubeStellar v0.21.0 or higher",
      "Metrics server on all clusters",
      "Valid subscription",
    ],
    compatibility: ["Linux", "macOS"],
    screenshots: [],
    documentation: "https://docs.scalestar.io",
    website: "https://scalestar.io",
    tags: [
      "autoscaling",
      "optimization",
      "cost-management",
      "ai",
      "premium",
    ],
  },
  {
    id: "9",
    name: "Policy Engine",
    slug: "policy-engine",
    tagline: "Advanced policy management and enforcement",
    description:
      "Define, manage, and enforce sophisticated policies across your KubeStellar infrastructure with a powerful policy engine supporting OPA and custom rules.",
    longDescription: `Policy Engine brings enterprise-grade policy management to KubeStellar. Define policies as code, version them in Git, and enforce them consistently across your entire cluster fleet. Support for Open Policy Agent (OPA), custom admission webhooks, and built-in rule sets.

Ensure compliance, security, and best practices are maintained automatically. Audit policy violations, get detailed reports, and integrate with your existing governance frameworks.

One-time purchase, lifetime updates. Perfect for organizations with strict compliance requirements.`,
    icon: "📋",
    category: "Governance",
    pricing: {
      type: "one-time",
      amount: 499,
    },
    author: "GovStar Solutions",
    downloads: 4312,
    rating: 4.9,
    version: "2.1.0",
    features: [
      "Policy-as-code with Git integration",
      "OPA and custom webhook support",
      "Real-time policy enforcement",
      "Detailed audit logging",
      "Compliance framework templates",
      "Policy testing and simulation",
      "Role-based policy management",
      "API for custom integrations",
    ],
    requirements: [
      "KubeStellar v0.20.0 or higher",
      "Open Policy Agent (optional)",
    ],
    compatibility: ["Linux", "macOS", "Windows"],
    screenshots: [],
    documentation: "https://docs.govstar.io",
    website: "https://govstar.io",
    tags: [
      "policy",
      "governance",
      "compliance",
      "security",
      "opa",
    ],
  },
  {
    id: "10",
    name: "Network Mesh Optimizer",
    slug: "network-mesh-optimizer",
    tagline: "Optimize network performance across distributed clusters",
    description:
      "Free tool to analyze and optimize network mesh configurations for better performance, lower latency, and reduced costs in your KubeStellar deployments.",
    longDescription: `Network Mesh Optimizer helps you get the most out of your service mesh in a KubeStellar environment. Analyze traffic patterns, identify inefficiencies, and get actionable recommendations to improve performance and reduce costs.

Visualize cross-cluster traffic flows, detect misconfigurations, and optimize routing policies. Compatible with Istio, Linkerd, and other popular service meshes.

Completely free and open source, developed by the KubeStellar community.`,
    icon: "🌐",
    category: "Networking",
    pricing: {
      type: "free",
    },
    author: "KubeStellar Community",
    downloads: 13892,
    rating: 4.5,
    version: "1.3.0",
    features: [
      "Network traffic analysis",
      "Service mesh health checks",
      "Performance optimization recommendations",
      "Traffic visualization",
      "Latency analysis",
      "Cost optimization suggestions",
      "Multi-mesh support",
      "CLI and web interface",
    ],
    requirements: [
      "KubeStellar v0.19.0 or higher",
      "Service mesh (Istio, Linkerd, etc.)",
    ],
    compatibility: ["Linux", "macOS", "Windows"],
    screenshots: [],
    documentation:
      "https://docs.kubestellar.io/plugins/network-mesh-optimizer",
    github: "https://github.com/kubestellar/network-mesh-optimizer",
    tags: [
      "networking",
      "service-mesh",
      "optimization",
      "performance",
      "free",
    ],
  },
  {
    id: "11",
    name: "GitOps Bridge",
    slug: "gitops-bridge",
    tagline: "Seamless GitOps integration for KubeStellar",
    description:
      "Bridge KubeStellar with your GitOps workflow. Manage cluster configurations through Git with full support for ArgoCD and Flux.",
    longDescription: `GitOps Bridge makes it effortless to manage your KubeStellar infrastructure using GitOps principles. Declare your desired cluster state in Git, and let GitOps Bridge handle the synchronization across your entire cluster fleet.

Full support for ArgoCD, Flux CD, and other popular GitOps tools. Automatic drift detection, reconciliation, and rollback capabilities ensure your infrastructure always matches your Git repository.

Free forever, because GitOps should be accessible to everyone.`,
    icon: "🔀",
    category: "GitOps",
    pricing: {
      type: "free",
    },
    author: "KubeStellar Core Team",
    downloads: 16543,
    rating: 4.8,
    version: "2.0.0",
    features: [
      "ArgoCD and Flux CD integration",
      "Automated drift detection",
      "Multi-cluster GitOps workflows",
      "Progressive delivery support",
      "Rollback capabilities",
      "Git-based RBAC",
      "Webhook support for CI/CD",
      "Detailed sync status reporting",
    ],
    requirements: [
      "KubeStellar v0.20.0 or higher",
      "Git repository",
      "ArgoCD or Flux CD (optional)",
    ],
    compatibility: ["Linux", "macOS", "Windows"],
    screenshots: [],
    documentation: "https://docs.kubestellar.io/plugins/gitops-bridge",
    github: "https://github.com/kubestellar/gitops-bridge",
    tags: ["gitops", "argocd", "flux", "automation", "free"],
  },
  {
    id: "12",
    name: "Cost Analyzer Premium",
    slug: "cost-analyzer-premium",
    tagline: "Advanced cost tracking and optimization for multi-cluster environments",
    description:
      "Track, analyze, and optimize costs across your entire KubeStellar infrastructure with granular insights and AI-powered recommendations.",
    longDescription: `Cost Analyzer Premium gives you complete visibility into your multi-cluster spending. Track costs by cluster, namespace, team, or any custom label. Understand exactly where your cloud budget is going and get AI-powered recommendations to reduce waste.

Set budgets, get alerts before you overspend, and forecast future costs with machine learning models. Generate beautiful reports for stakeholders and integrate with your existing FinOps tools.

Pay once, use forever. Includes all future updates and features.`,
    icon: "💰",
    category: "Cost Management",
    pricing: {
      type: "one-time",
      amount: 399,
    },
    author: "FinOps Labs",
    downloads: 7234,
    rating: 4.7,
    version: "3.2.1",
    features: [
      "Granular cost tracking and attribution",
      "AI-powered optimization recommendations",
      "Budget management and alerts",
      "Cost forecasting and trending",
      "Chargeback and showback reports",
      "Multi-cloud cost aggregation",
      "Custom dashboards and reports",
      "Integration with cloud billing APIs",
    ],
    requirements: [
      "KubeStellar v0.20.0 or higher",
      "Access to cloud provider billing APIs",
    ],
    compatibility: ["Linux", "macOS", "Windows"],
    screenshots: [],
    documentation: "https://docs.finopslabs.io",
    website: "https://finopslabs.io",
    tags: [
      "cost-management",
      "finops",
      "analytics",
      "optimization",
      "premium",
    ],
  },
];
