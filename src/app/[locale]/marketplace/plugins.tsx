"use client";

import { useTranslations } from "next-intl";

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

export function usePlugins(): Plugin[] {
  const t = useTranslations("marketplace");

  return [
    {
      id: "1",
      name: "KubeCtl Multi",
      slug: "kubectl-multi",
      tagline:
        "Execute kubectl commands across multiple clusters simultaneously",
      description:
        "Streamline your multi-cluster operations with KubeCtl Multi. Execute commands across all your KubeStellar-managed clusters with a single command.",
      longDescription: `KubeCtl Multi is the ultimate productivity tool for managing multiple Kubernetes clusters through KubeStellar. Built specifically for the KubeStellar ecosystem, it allows you to execute kubectl commands across all your managed clusters simultaneously, saving hours of manual work.

With intelligent context switching and parallel execution, KubeCtl Multi ensures your operations are fast, reliable, and consistent across your entire infrastructure. Whether you're deploying applications, updating configurations, or troubleshooting issues, KubeCtl Multi makes it seamless.

Perfect for DevOps teams managing edge deployments, hybrid clouds, or geographically distributed clusters.`,
      icon: "🎯",
      category: t("categories.cliTools"),
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
      category: t("categories.synchronization"),
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
      tags: ["sync", "enterprise", "monitoring", "automation", "premium"],
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
      category: t("categories.security"),
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
      tagline:
        "Advanced analytics and observability for multi-cluster deployments",
      description:
        "Gain deep insights into your KubeStellar deployments with advanced metrics, tracing, and AI-powered anomaly detection.",
      longDescription: `Stellar Insights transforms your KubeStellar deployment into a fully observable system. Get deep visibility into every aspect of your multi-cluster infrastructure with advanced metrics collection, distributed tracing, and AI-powered analytics.

Our machine learning algorithms detect anomalies before they impact your applications, predict resource needs, and suggest optimizations. Beautiful, customizable dashboards give your team the insights they need at a glance.

Reduce MTTR by 70% and proactively prevent 90% of incidents with Stellar Insights.`,
      icon: "📊",
      category: t("categories.observability"),
      pricing: {
        type: "one-time",
        amount: 299,
      },
      author: "ObservaStar Technologies",
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
      tags: ["observability", "metrics", "analytics", "monitoring", "ai"],
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
      category: t("categories.visualization"),
      pricing: {
        type: "free",
      },
      author: "KubeStellar Core Team",
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
      tags: ["visualization", "topology", "ui", "free", "open-source"],
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
      requirements: ["KubeStellar v0.18.0 or higher", "kubectl v1.25+"],
      compatibility: ["Linux", "macOS", "Windows"],
      screenshots: [],
      documentation: "https://docs.kubestellar.io/plugins/config-validator",
      github: "https://github.com/kubestellar/config-validator",
      tags: ["validation", "ci-cd", "configuration", "development", "free"],
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
      tags: ["autoscaling", "optimization", "cost-management", "ai", "premium"],
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
      tags: ["policy", "governance", "compliance", "security", "opa"],
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
      tagline:
        "Advanced cost tracking and optimization for multi-cluster environments",
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
    {
      id: "13",
      name: "Secrets Manager",
      slug: "secrets-manager",
      tagline: "Centralized secrets management for distributed clusters",
      description:
        "Securely manage and distribute secrets across all your KubeStellar clusters with automatic rotation, encryption, and audit logging.",
      longDescription: `Secrets Manager provides enterprise-grade secrets management for your KubeStellar infrastructure. Centralize secret storage, automate rotation, and ensure secrets are encrypted at rest and in transit. Built-in integration with HashiCorp Vault, AWS Secrets Manager, and Azure Key Vault.

Automatic secret rotation prevents credential compromise, while detailed audit logs ensure compliance. RBAC controls determine who can access which secrets, and automatic sync ensures secrets are always up-to-date across all clusters.`,
      icon: "🔐",
      category: "Security",
      pricing: {
        type: "monthly",
        amount: 129,
      },
      author: "VaultStar Security",
      downloads: 8456,
      rating: 4.9,
      version: "1.6.0",
      features: [
        "Centralized secret storage",
        "Automatic secret rotation",
        "Integration with Vault, AWS, Azure",
        "End-to-end encryption",
        "Detailed audit logging",
        "RBAC and access policies",
        "Cross-cluster secret sync",
        "Emergency secret revocation",
      ],
      requirements: [
        "KubeStellar v0.21.0 or higher",
        "External secrets backend (optional)",
      ],
      compatibility: ["Linux", "macOS"],
      screenshots: [],
      documentation: "https://docs.vaultstar.io",
      website: "https://vaultstar.io",
      tags: ["security", "secrets", "vault", "encryption", "premium"],
    },
    {
      id: "14",
      name: "Load Balancer Controller",
      slug: "load-balancer-controller",
      tagline: "Intelligent load balancing across cluster boundaries",
      description:
        "Free load balancer controller optimized for KubeStellar deployments. Distribute traffic intelligently across clusters based on latency, capacity, and custom rules.",
      longDescription: `Load Balancer Controller extends Kubernetes load balancing capabilities to work seamlessly across KubeStellar-managed clusters. Route traffic based on geographic proximity, cluster capacity, custom weights, or health scores.

Automatic failover ensures high availability, while built-in circuit breakers prevent cascading failures. Compatible with major cloud load balancers and on-premises solutions.

Open source and community-driven, perfect for hybrid and multi-cloud deployments.`,
      icon: "⚖️",
      category: "Networking",
      pricing: {
        type: "free",
      },
      author: "KubeStellar Community",
      downloads: 14782,
      rating: 4.6,
      version: "2.2.0",
      features: [
        "Cross-cluster load balancing",
        "Geographic traffic routing",
        "Health-based routing",
        "Automatic failover",
        "Circuit breaker support",
        "Custom routing rules",
        "Integration with cloud LBs",
        "Real-time metrics",
      ],
      requirements: ["KubeStellar v0.20.0 or higher", "Kubernetes 1.24+"],
      compatibility: ["Linux", "macOS"],
      screenshots: [],
      documentation: "https://docs.kubestellar.io/plugins/lb-controller",
      github: "https://github.com/kubestellar/lb-controller",
      tags: ["networking", "load-balancing", "traffic", "free", "open-source"],
    },
    {
      id: "15",
      name: "Workload Migrator",
      slug: "workload-migrator",
      tagline: "Seamlessly migrate workloads between clusters",
      description:
        "Migrate applications and data between KubeStellar clusters with zero downtime. Perfect for cluster upgrades, disaster recovery, and rebalancing.",
      longDescription: `Workload Migrator makes cluster migrations painless. Whether you're upgrading clusters, rebalancing workloads, or responding to disasters, Workload Migrator ensures smooth transitions with zero downtime.

Intelligent pre-flight checks catch potential issues, automated data migration handles persistent volumes, and gradual traffic shifting ensures reliability. Rollback capabilities provide a safety net if anything goes wrong.

Essential tool for production KubeStellar deployments.`,
      icon: "🚚",
      category: "CLI Tools",
      pricing: {
        type: "one-time",
        amount: 199,
      },
      author: "MigrateStar",
      downloads: 5934,
      rating: 4.8,
      version: "1.4.2",
      features: [
        "Zero-downtime migrations",
        "Pre-flight validation checks",
        "Persistent volume migration",
        "Gradual traffic shifting",
        "Automated rollback",
        "Migration scheduling",
        "Progress tracking",
        "Dry-run mode",
      ],
      requirements: ["KubeStellar v0.21.0 or higher", "kubectl v1.26+"],
      compatibility: ["Linux", "macOS", "Windows"],
      screenshots: [],
      documentation: "https://docs.migratestar.io",
      website: "https://migratestar.io",
      tags: ["migration", "workload", "cli", "disaster-recovery", "premium"],
    },
    {
      id: "16",
      name: "Metrics Aggregator",
      slug: "metrics-aggregator",
      tagline: "Unified metrics collection from all clusters",
      description:
        "Free tool to aggregate metrics from all your KubeStellar clusters into a single Prometheus instance. Simplify monitoring and alerting.",
      longDescription: `Metrics Aggregator solves the challenge of monitoring distributed clusters by aggregating metrics into a centralized location. Compatible with Prometheus, it provides a unified view of your entire infrastructure without the complexity of federated setups.

Automatic cluster discovery, intelligent metric filtering, and built-in dashboards get you up and running in minutes. Reduce monitoring overhead and costs while maintaining complete visibility.

Free forever, built by the community for the community.`,
      icon: "📉",
      category: "Observability",
      pricing: {
        type: "free",
      },
      author: "KubeStellar Community",
      downloads: 12876,
      rating: 4.5,
      version: "1.7.1",
      features: [
        "Multi-cluster metrics aggregation",
        "Prometheus compatibility",
        "Automatic cluster discovery",
        "Metric filtering and sampling",
        "Built-in Grafana dashboards",
        "Low resource overhead",
        "High availability support",
        "Custom metric pipelines",
      ],
      requirements: ["KubeStellar v0.19.0 or higher", "Prometheus operator"],
      compatibility: ["Linux", "macOS"],
      screenshots: [],
      documentation: "https://docs.kubestellar.io/plugins/metrics-aggregator",
      github: "https://github.com/kubestellar/metrics-aggregator",
      tags: ["observability", "metrics", "prometheus", "monitoring", "free"],
    },
    {
      id: "17",
      name: "Cluster Provisioner",
      slug: "cluster-provisioner",
      tagline: "Automate cluster creation and onboarding",
      description:
        "Streamline cluster provisioning with automated creation, configuration, and onboarding to KubeStellar. Support for major cloud providers and on-prem.",
      longDescription: `Cluster Provisioner eliminates the manual work of creating and onboarding new clusters. Define your cluster specifications in code, and let Cluster Provisioner handle creation across AWS, Azure, GCP, or on-premises infrastructure.

Automatic KubeStellar onboarding ensures new clusters are ready to use immediately. Built-in compliance templates ensure clusters meet your security and governance requirements from day one.

One-time purchase includes lifetime updates.`,
      icon: "🏗️",
      category: "CLI Tools",
      pricing: {
        type: "one-time",
        amount: 349,
      },
      author: "ProvisionStar",
      downloads: 6123,
      rating: 4.7,
      version: "2.1.3",
      features: [
        "Multi-cloud cluster creation",
        "Infrastructure-as-code support",
        "Automatic KubeStellar onboarding",
        "Compliance template library",
        "Cluster templating",
        "Batch provisioning",
        "Cost estimation",
        "Terraform and Pulumi integration",
      ],
      requirements: [
        "KubeStellar v0.20.0 or higher",
        "Cloud provider credentials",
      ],
      compatibility: ["Linux", "macOS", "Windows"],
      screenshots: [],
      documentation: "https://docs.provisionstar.io",
      website: "https://provisionstar.io",
      tags: ["provisioning", "automation", "infrastructure", "cli", "premium"],
    },
    {
      id: "18",
      name: "Alert Manager Plus",
      slug: "alert-manager-plus",
      tagline: "Advanced alerting and incident management",
      description:
        "Sophisticated alerting system for KubeStellar with smart routing, deduplication, and integration with popular incident management platforms.",
      longDescription: `Alert Manager Plus transforms Kubernetes alerts into actionable insights. Smart alert routing ensures the right people get notified, while intelligent deduplication reduces noise. Integration with PagerDuty, OpsGenie, and Slack ensures alerts reach your team wherever they are.

Alert correlation detects patterns and prevents alert storms. Runbook automation can auto-remediate common issues, reducing manual intervention and MTTR.

Affordable monthly pricing with enterprise support.`,
      icon: "🚨",
      category: "Observability",
      pricing: {
        type: "monthly",
        amount: 89,
      },
      author: "AlertStar Inc.",
      downloads: 9234,
      rating: 4.8,
      version: "1.5.0",
      features: [
        "Smart alert routing",
        "Intelligent deduplication",
        "PagerDuty/OpsGenie integration",
        "Alert correlation",
        "Runbook automation",
        "Escalation policies",
        "Custom alert templates",
        "Detailed analytics",
      ],
      requirements: [
        "KubeStellar v0.20.0 or higher",
        "Prometheus or compatible metrics source",
      ],
      compatibility: ["Linux", "macOS"],
      screenshots: [],
      documentation: "https://docs.alertstar.io",
      website: "https://alertstar.io",
      tags: [
        "alerting",
        "observability",
        "incident-management",
        "automation",
        "premium",
      ],
    },
    {
      id: "19",
      name: "Resource Optimizer",
      slug: "resource-optimizer",
      tagline: "AI-powered resource recommendations",
      description:
        "Free tool that analyzes workload patterns and provides AI-powered recommendations for right-sizing resources across your clusters.",
      longDescription: `Resource Optimizer uses machine learning to analyze your workload behavior and recommend optimal resource allocations. Stop wasting money on over-provisioned pods and prevent performance issues from under-provisioning.

Historical analysis identifies trends, while predictive modeling forecasts future needs. One-click apply makes optimization effortless. Typical users save 30-40% on infrastructure costs.

Completely free and open source.`,
      icon: "🎛️",
      category: "Resource Management",
      pricing: {
        type: "free",
      },
      author: "KubeStellar Community",
      downloads: 11567,
      rating: 4.6,
      version: "1.3.2",
      features: [
        "AI-powered recommendations",
        "Historical usage analysis",
        "Predictive resource modeling",
        "One-click optimization",
        "Cost savings estimates",
        "Cluster-wide analysis",
        "Custom recommendation policies",
        "Integration with VPA",
      ],
      requirements: ["KubeStellar v0.20.0 or higher", "Metrics server"],
      compatibility: ["Linux", "macOS", "Windows"],
      screenshots: [],
      documentation: "https://docs.kubestellar.io/plugins/resource-optimizer",
      github: "https://github.com/kubestellar/resource-optimizer",
      tags: ["optimization", "resources", "ai", "cost-savings", "free"],
    },
    {
      id: "20",
      name: "Compliance Reporter",
      slug: "compliance-reporter",
      tagline: "Automated compliance reporting and auditing",
      description:
        "Generate compliance reports for SOC 2, ISO 27001, HIPAA, and custom frameworks. Continuous monitoring ensures you stay compliant.",
      longDescription: `Compliance Reporter automates the tedious work of compliance auditing and reporting. Continuous monitoring checks your KubeStellar infrastructure against compliance frameworks, automatically generating reports and alerting you to violations.

Built-in templates for SOC 2, ISO 27001, HIPAA, PCI-DSS, and GDPR. Custom framework support lets you define your own requirements. Evidence collection happens automatically, making audits stress-free.

Essential for regulated industries. One-time purchase, lifetime value.`,
      icon: "📝",
      category: "Governance",
      pricing: {
        type: "one-time",
        amount: 449,
      },
      author: "ComplianceStar",
      downloads: 4567,
      rating: 4.9,
      version: "2.0.1",
      features: [
        "Automated compliance monitoring",
        "SOC 2, ISO 27001, HIPAA templates",
        "Custom framework support",
        "Automated evidence collection",
        "Scheduled report generation",
        "Violation alerts",
        "Audit trail management",
        "Export to PDF, CSV, JSON",
      ],
      requirements: ["KubeStellar v0.21.0 or higher", "Audit logging enabled"],
      compatibility: ["Linux", "macOS", "Windows"],
      screenshots: [],
      documentation: "https://docs.compliancestar.io",
      website: "https://compliancestar.io",
      tags: ["compliance", "governance", "auditing", "reporting", "premium"],
    },
    {
      id: "21",
      name: "Service Mesh Bridge",
      slug: "service-mesh-bridge",
      tagline: "Connect and manage multiple service meshes",
      description:
        "Unified management for Istio, Linkerd, and Consul service meshes across your KubeStellar clusters. Centralized observability and control.",
      longDescription: `Service Mesh Bridge provides a unified control plane for managing multiple service mesh implementations across your distributed clusters. Whether you're running Istio, Linkerd, Consul, or a mix of different meshes, Service Mesh Bridge gives you a single pane of glass for configuration, monitoring, and troubleshooting.

Cross-mesh service discovery enables services in different meshes to communicate seamlessly. Advanced traffic management features allow you to implement sophisticated routing, load balancing, and failover strategies across mesh boundaries.

Free forever, built by mesh enthusiasts for mesh enthusiasts.`,
      icon: "🕸️",
      category: "Networking",
      pricing: {
        type: "free",
      },
      author: "KubeStellar Community",
      downloads: 9876,
      rating: 4.7,
      version: "1.2.4",
      features: [
        "Multi-mesh management",
        "Cross-mesh service discovery",
        "Unified observability dashboard",
        "Traffic mirroring and shadowing",
        "Centralized policy management",
        "Mesh migration tools",
        "Performance metrics aggregation",
        "Integration with popular meshes",
      ],
      requirements: [
        "KubeStellar v0.20.0 or higher",
        "Service mesh installed (Istio/Linkerd/Consul)",
      ],
      compatibility: ["Linux", "macOS"],
      screenshots: [],
      documentation: "https://docs.kubestellar.io/plugins/service-mesh-bridge",
      github: "https://github.com/kubestellar/service-mesh-bridge",
      tags: ["service-mesh", "networking", "istio", "linkerd", "free"],
    },
    {
      id: "22",
      name: "Edge Monitor Pro",
      slug: "edge-monitor-pro",
      tagline: "Specialized monitoring for edge computing deployments",
      description:
        "Monitor edge devices and clusters with limited connectivity. Offline-first design, bandwidth-efficient metrics collection, and intelligent data aggregation.",
      longDescription: `Edge Monitor Pro is built from the ground up for the unique challenges of edge computing. Handle intermittent connectivity, bandwidth constraints, and resource-limited devices with ease. Smart data aggregation ensures you get the insights you need without overwhelming your edge infrastructure.

Local data retention keeps critical metrics available even when connectivity is lost. When connection is restored, intelligent sync ensures central visibility without flooding your network. Edge-specific alerts handle scenarios like prolonged disconnection or unusual power consumption.

Purpose-built for IoT, retail, manufacturing, and remote deployments.`,
      icon: "📡",
      category: "Observability",
      pricing: {
        type: "monthly",
        amount: 119,
      },
      author: "EdgeTech Solutions",
      downloads: 6234,
      rating: 4.8,
      version: "2.4.0",
      features: [
        "Offline-first architecture",
        "Bandwidth-efficient metrics",
        "Local data retention",
        "Intelligent sync when online",
        "Edge-specific alerts",
        "Device health monitoring",
        "Power consumption tracking",
        "Mobile app for remote access",
      ],
      requirements: [
        "KubeStellar v0.21.0 or higher",
        "Edge clusters with limited resources",
      ],
      compatibility: ["Linux", "ARM"],
      screenshots: [],
      documentation: "https://docs.edgetechsolutions.io",
      website: "https://edgetechsolutions.io",
      tags: ["edge", "monitoring", "iot", "observability", "premium"],
    },
    {
      id: "23",
      name: "Multi-Tenancy Manager",
      slug: "multi-tenancy-manager",
      tagline: "Enterprise-grade multi-tenancy for KubeStellar",
      description:
        "Implement secure multi-tenancy with isolated namespaces, RBAC templates, resource quotas, and tenant-specific policies across your cluster fleet.",
      longDescription: `Multi-Tenancy Manager transforms your KubeStellar deployment into a secure, multi-tenant platform. Onboard new tenants in seconds with pre-configured isolation, security policies, and resource limits. Each tenant gets their own isolated environment while you maintain centralized control and visibility.

Hierarchical namespace support enables complex organizational structures. Per-tenant billing and showback reports make it easy to track and allocate costs. Advanced RBAC templates ensure security best practices are enforced automatically.

Perfect for SaaS platforms, shared services, and organizations with multiple teams.`,
      icon: "🏢",
      category: "Governance",
      pricing: {
        type: "one-time",
        amount: 599,
      },
      author: "TenantStar Enterprise",
      downloads: 3892,
      rating: 4.9,
      version: "3.0.1",
      features: [
        "Automated tenant onboarding",
        "Hierarchical namespaces",
        "RBAC template library",
        "Resource quota management",
        "Network isolation policies",
        "Per-tenant billing reports",
        "Tenant self-service portal",
        "Audit logging per tenant",
      ],
      requirements: ["KubeStellar v0.21.0 or higher", "Kubernetes 1.25+"],
      compatibility: ["Linux", "macOS", "Windows"],
      screenshots: [],
      documentation: "https://docs.tenantstar.io",
      website: "https://tenantstar.io",
      tags: ["multi-tenancy", "governance", "rbac", "enterprise", "premium"],
    },
    {
      id: "24",
      name: "Chaos Engineering Toolkit",
      slug: "chaos-engineering-toolkit",
      tagline: "Test resilience with controlled chaos experiments",
      description:
        "Free chaos engineering toolkit for KubeStellar. Inject failures, test failover mechanisms, and validate resilience across your distributed infrastructure.",
      longDescription: `Chaos Engineering Toolkit brings the power of chaos engineering to KubeStellar deployments. Test your system's resilience by injecting controlled failures, simulating network issues, and validating failover mechanisms. Learn how your system behaves under stress before real failures occur.

Pre-built experiment templates cover common scenarios like pod failures, network latency, resource exhaustion, and cluster partitions. Safety mechanisms ensure experiments can be aborted and rolled back if needed. Detailed reports show exactly how your system responded.

Free and open source, inspired by Chaos Mesh and powered by the community.`,
      icon: "🌪️",
      category: "Development Tools",
      pricing: {
        type: "free",
      },
      author: "KubeStellar Community",
      downloads: 7654,
      rating: 4.6,
      version: "1.6.0",
      features: [
        "Pod failure injection",
        "Network chaos (latency, packet loss)",
        "Resource stress testing",
        "Time chaos experiments",
        "Scheduled experiments",
        "Safety mechanisms and rollback",
        "Experiment templates library",
        "Integration with CI/CD",
      ],
      requirements: ["KubeStellar v0.20.0 or higher", "Kubernetes 1.24+"],
      compatibility: ["Linux", "macOS"],
      screenshots: [],
      documentation: "https://docs.kubestellar.io/plugins/chaos-toolkit",
      github: "https://github.com/kubestellar/chaos-toolkit",
      tags: ["chaos-engineering", "testing", "resilience", "sre", "free"],
    },
  ];
}

// For backward compatibility, export a static version
export const plugins: Plugin[] = [];
