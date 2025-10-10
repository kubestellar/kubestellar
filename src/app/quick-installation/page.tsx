"use client";

import React, { useState } from "react";
import { motion } from "framer-motion";
import { 
  Terminal, 
  Server, 
  CheckCircle2, 
  Copy, 
  ChevronRight, 
  ChevronDown,
  ExternalLink,
  Zap,
  Info
} from "lucide-react";
import Navigation from "@/components/Navigation";
import Footer from "@/components/Footer";
import { GridLines, StarField } from "@/components/index";

// Define platform type for installation
type Platform = 'kind' | 'k3d';

// Define prerequisite data structure
interface Prerequisite {
  name: string;
  displayName: string;
  description: string;
  minVersion?: string;
  installCommand: string;
  installUrl: string;
  versionCommand: string;
}

// Define prerequisite category structure
interface PrerequisiteCategory {
  title: string;
  description: string;
  icon: React.ReactNode;
  prerequisites: Prerequisite[];
}

// Core prerequisites for using KubeStellar
const corePrerequisites: Prerequisite[] = [
  {
    name: 'docker',
    displayName: 'Docker',
    description: 'Container runtime platform',
    minVersion: '20.0.0',
    installCommand: 'curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh',
    installUrl: 'https://docs.docker.com/engine/install/',
    versionCommand: 'docker version --format "{{.Client.Version}}"',
  },
  {
    name: 'kubectl',
    displayName: 'kubectl',
    description: 'Kubernetes command-line tool',
    minVersion: '1.27.0',
    installCommand: 'curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && sudo mv kubectl /usr/local/bin/',
    installUrl: 'https://kubernetes.io/docs/tasks/tools/',
    versionCommand: 'kubectl version --client',
  },
  {
    name: 'kubeflex',
    displayName: 'KubeFlex',
    description: 'Core component for multi-cluster management',
    minVersion: '0.8.0',
    installCommand: 'bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubeflex/main/scripts/install-kubeflex.sh) --ensure-folder /usr/local/bin --strip-bin',
    installUrl: 'https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation',
    versionCommand: 'kflex version',
  },
  {
    name: 'clusteradm',
    displayName: 'OCM CLI',
    description: 'Open Cluster Management command line interface',
    minVersion: '0.7.0',
    installCommand: 'bash <(curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh) 0.10.1',
    installUrl: 'https://docs.kubestellar.io/latest/direct/pre-reqs/',
    versionCommand: 'clusteradm version',
  },
  {
    name: 'helm',
    displayName: 'Helm',
    description: 'Package manager for Kubernetes',
    minVersion: '3.0.0',
    installCommand: 'curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash',
    installUrl: 'https://helm.sh/docs/intro/install/',
    versionCommand: 'helm version',
  },
];

// Additional prerequisites for running examples
const examplePrerequisites: Prerequisite[] = [
  {
    name: 'kind',
    displayName: 'kind',
    description: 'Kubernetes IN Docker - local Kubernetes cluster',
    minVersion: '0.20.0',
    installCommand: 'curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind',
    installUrl: 'https://kind.sigs.k8s.io/docs/user/quick-start/#installation',
    versionCommand: 'kind version',
  },
  {
    name: 'k3d',
    displayName: 'k3d',
    description: 'Lightweight wrapper to run k3s in Docker',
    minVersion: '5.0.0',
    installCommand: 'curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash',
    installUrl: 'https://k3d.io/v5.4.6/#installation',
    versionCommand: 'k3d version',
  },
  {
    name: 'argo',
    displayName: 'Argo CD CLI',
    description: 'GitOps continuous delivery tool for Kubernetes',
    minVersion: '2.8.0',
    installCommand: 'curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64 && sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd && rm argocd-linux-amd64',
    installUrl: 'https://argo-cd.readthedocs.io/en/stable/cli_installation/',
    versionCommand: 'argocd version --client',
  },
];

// Prerequisites for building KubeStellar
const buildPrerequisites: Prerequisite[] = [
  {
    name: 'make',
    displayName: 'Make',
    description: 'Build automation tool',
    installCommand: 'sudo apt-get update && sudo apt-get install -y build-essential',
    installUrl: 'https://www.gnu.org/software/make/',
    versionCommand: 'make --version',
  },
  {
    name: 'go',
    displayName: 'Go',
    description: 'Programming language for building KubeStellar',
    minVersion: '1.21.0',
    installCommand: 'curl -fsSL https://golang.org/dl/go1.21.6.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf - && echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc',
    installUrl: 'https://golang.org/doc/install',
    versionCommand: 'go version',
  },
  {
    name: 'ko',
    displayName: 'ko',
    description: 'Container image builder for Go applications',
    minVersion: '0.14.0',
    installCommand: 'go install github.com/google/ko@latest',
    installUrl: 'https://ko.build/install/',
    versionCommand: 'ko version',
  },
];

// Define prerequisite categories
const prerequisiteCategories: PrerequisiteCategory[] = [
  {
    title: "Core Prerequisites",
    description: "Essential tools for using KubeStellar",
    icon: <Server size={24} />,
    prerequisites: corePrerequisites,
  },
  {
    title: "Additional Prerequisites",
    description: "Additional tools for running KubeStellar examples",
    icon: <Terminal size={24} />,
    prerequisites: examplePrerequisites,
  },
  {
    title: "Build Prerequisites",
    description: "Tools required for building KubeStellar from source",
    icon: <Zap size={24} />,
    prerequisites: buildPrerequisites,
  },
];

// Platform installation scripts
const platformInstallationScripts = {
  kind: 'bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/refs/tags/v0.27.2/scripts/create-kubestellar-demo-env.sh) --platform kind',
  k3d: 'bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/refs/tags/v0.27.2/scripts/create-kubestellar-demo-env.sh) --platform k3d'
};

// FAQ data
const faqData = [
  {
    id: 1,
    question: "What's the difference between Core, Additional, and Build prerequisites?",
    answer: "Core prerequisites are essential for using KubeStellar. Additional prerequisites are needed for running examples and demos. Build prerequisites are only required if you plan to build KubeStellar from source code."
  },
  {
    id: 2,
    question: "Can I automatically check if I have all prerequisites installed?",
    answer: "Yes! Use the automated prerequisite check script: 'curl -fsSL https://raw.githubusercontent.com/kubestellar/kubestellar/refs/heads/main/scripts/check_pre_req.sh | bash'. This script will verify all prerequisites and provide installation guidance for missing tools."
  },
  {
    id: 3,
    question: "Do I need to install all prerequisites?",
    answer: "For basic KubeStellar usage, you only need the Core prerequisites. Install Additional prerequisites if you want to run examples. Build prerequisites are only needed for development and building from source."
  },
  {
    id: 4,
    question: "Can I use KubeStellar with existing Kubernetes clusters?",
    answer: "Yes! KubeStellar can manage existing Kubernetes clusters. You can connect your production clusters alongside local development clusters for unified multi-cluster management."
  },
  {
    id: 5,
    question: "What are the minimum system requirements?",
    answer: "KubeStellar requires at least 4GB RAM and 2 CPU cores. You'll need Docker (v20.0+), kubectl (v1.27+), and either kind (v0.20+) or k3d for local clusters."
  }
];

// Animated card component
const AnimatedCard = ({
  children,
  className = '',
  id,
}: {
  children: React.ReactNode;
  className?: string;
  id?: string;
}) => {
  return (
    <div
      id={id}
      className={`relative bg-gray-800/50 backdrop-blur-md rounded-xl shadow-lg border border-gray-700/50 transition-all duration-300 ${className}`}
    >
      <div className="absolute -inset-0.5 bg-gradient-to-r from-blue-600/20 to-purple-600/20 rounded-xl blur opacity-30"></div>
      <div className="relative">
        {children}
      </div>
    </div>
  );
};

// Code block component with copy button
const CodeBlock = ({
  code,
  language = 'bash',
}: {
  code: string;
  language?: string;
}) => {
  const [copied, setCopied] = useState(false);

  const copyToClipboard = () => {
    navigator.clipboard.writeText(code).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <div className="relative mb-4 overflow-hidden rounded-lg border border-gray-700/50 bg-gray-900/50">
      <div className="flex items-center justify-between border-b border-gray-700/50 bg-gray-800/50 px-4 py-2">
        <span className="font-mono text-xs text-gray-400">{language}</span>
        <button
          onClick={copyToClipboard}
          className="rounded bg-gray-700/50 p-1.5 transition-all hover:bg-gray-600/50"
          aria-label="Copy code"
        >
          {copied ? (
            <CheckCircle2 size={14} className="text-emerald-400" />
          ) : (
            <Copy size={14} className="text-gray-400" />
          )}
        </button>
      </div>
      <div className="overflow-x-auto p-4">
        <pre className="whitespace-pre-wrap break-all font-mono text-sm text-emerald-300">
          {code}
        </pre>
      </div>
    </div>
  );
};

// Section header component
const SectionHeader = ({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) => {
  return (
    <div className="mb-8">
      <div className="mb-2 flex items-center">
        <div className="mr-3 text-blue-400">{icon}</div>
        <h2 className="text-2xl font-bold text-white">{title}</h2>
      </div>
      <p className="ml-9 text-gray-300">{description}</p>
    </div>
  );
};

// FAQ Item component with dropdown functionality
const FAQItem = ({ faq }: { faq: { id: number; question: string; answer: string } }) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="border border-gray-700/50 rounded-lg bg-gray-800/50 backdrop-blur-md">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full p-6 text-left flex items-center justify-between hover:bg-gray-700/30 transition-colors duration-200"
      >
        <h3 className="text-lg font-medium text-white pr-4">{faq.question}</h3>
        <div className="text-gray-400 flex-shrink-0">
          {isOpen ? (
            <ChevronDown size={20} className="transition-transform duration-200" />
          ) : (
            <ChevronRight size={20} className="transition-transform duration-200" />
          )}
        </div>
      </button>
      
      {isOpen && (
        <motion.div
          initial={{ height: 0, opacity: 0 }}
          animate={{ height: "auto", opacity: 1 }}
          exit={{ height: 0, opacity: 0 }}
          transition={{ duration: 0.3, ease: "easeInOut" }}
          className="overflow-hidden"
        >
          <div className="px-6 pb-6 border-t border-gray-700/30">
            <p className="text-gray-300 leading-relaxed pt-4">{faq.answer}</p>
          </div>
        </motion.div>
      )}
    </div>
  );
};

// Prerequisite card component
const PrerequisiteCard = ({ prerequisite }: { prerequisite: Prerequisite }) => {
  return (
    <div className="relative group h-full">
      <div className="absolute -inset-0.5 bg-gradient-to-r from-blue-600/20 to-purple-600/20 rounded-lg blur opacity-30 group-hover:opacity-60 transition duration-300"></div>
      <div className="relative bg-gray-800/50 backdrop-blur-md rounded-lg shadow-lg p-6 border border-gray-700/50 transition-all duration-300 hover:border-gray-600/70 h-full flex flex-col">
        {/* Header */}
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-white">{prerequisite.displayName}</h3>
          <span className="rounded-full bg-blue-500/20 border border-blue-500/30 px-2 py-1 font-mono text-xs text-blue-400">
            {prerequisite.minVersion ? `v${prerequisite.minVersion}+` : 'Latest'}
          </span>
        </div>
        
        <p className="mb-4 text-gray-300 text-sm flex-grow">{prerequisite.description}</p>
        
        <div className="space-y-3 mt-auto">
          <div>
            <h4 className="mb-2 flex items-center text-xs font-medium text-emerald-400">
              <Terminal size={14} className="mr-1" />
              Install:
            </h4>
            <CodeBlock code={prerequisite.installCommand} />
          </div>
          
          <div>
            <h4 className="mb-2 flex items-center text-xs font-medium text-blue-400">
              <CheckCircle2 size={14} className="mr-1" />
              Verify:
            </h4>
            <CodeBlock code={prerequisite.versionCommand} />
          </div>
        </div>
      </div>
    </div>
  );
};

// Prerequisite category section component
const PrerequisiteCategorySection = ({ category }: { category: PrerequisiteCategory }) => {
  return (
    <div className="mb-12">
      <div className="mb-6">
        <div className="mb-2 flex items-center">
          <div className="mr-3 text-blue-400">{category.icon}</div>
          <h3 className="text-xl font-bold text-white">{category.title}</h3>
        </div>
        <p className="ml-9 text-gray-400 text-sm">{category.description}</p>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {category.prerequisites.map((prerequisite) => (
          <PrerequisiteCard key={prerequisite.name} prerequisite={prerequisite} />
        ))}
      </div>
    </div>
  );
};

// Platform selection component
const PlatformSelector = ({
  platform,
  setPlatform,
}: {
  platform: Platform;
  setPlatform: (platform: Platform) => void;
}) => {
  return (
    <div className="mb-6">
      <h3 className="mb-4 text-lg font-medium text-white">Choose Your Platform:</h3>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <button
          onClick={() => setPlatform('kind')}
          className={`group relative rounded-lg border p-4 text-left transition-all duration-300 ${
            platform === 'kind'
              ? 'border-blue-500/50 bg-blue-500/10'
              : 'border-gray-700/50 bg-gray-800/30 hover:border-gray-600/70'
          }`}
        >
          <div className="absolute -inset-0.5 bg-gradient-to-r from-blue-600/20 to-purple-600/20 rounded-lg blur opacity-0 group-hover:opacity-30 transition duration-300"></div>
          <div className="relative flex items-center justify-between">
            <div>
              <h4 className="font-medium text-white">kind</h4>
              <p className="text-sm text-gray-400">Kubernetes in Docker</p>
            </div>
            {platform === 'kind' && <CheckCircle2 size={18} className="text-blue-400" />}
          </div>
        </button>
        
        <button
          onClick={() => setPlatform('k3d')}
          className={`group relative rounded-lg border p-4 text-left transition-all duration-300 ${
            platform === 'k3d'
              ? 'border-blue-500/50 bg-blue-500/10'
              : 'border-gray-700/50 bg-gray-800/30 hover:border-gray-600/70'
          }`}
        >
          <div className="absolute -inset-0.5 bg-gradient-to-r from-blue-600/20 to-purple-600/20 rounded-lg blur opacity-0 group-hover:opacity-30 transition duration-300"></div>
          <div className="relative flex items-center justify-between">
            <div>
              <h4 className="font-medium text-white">k3d</h4>
              <p className="text-sm text-gray-400">Lightweight Kubernetes</p>
            </div>
            {platform === 'k3d' && <CheckCircle2 size={18} className="text-blue-400" />}
          </div>
        </button>
      </div>
    </div>
  );
};

// Main Quick Installation Page component
const QuickInstallationPage = () => {
  const [platform, setPlatform] = useState<Platform>('kind');

  return (
    <main className="min-h-screen">
      <Navigation />
      
      <section className="relative py-24 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden">
        {/* Dark base background */}
        <div className="absolute inset-0 bg-[#0a0a0a]"></div>

        {/* Starfield background */}
        <StarField
          density="low"
          showComets={true}
          cometCount={2}
        />

        {/* Grid lines background */}
        <GridLines />

        {/* Content */}
        <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="text-center mb-16">
            <h1 className="text-5xl md:text-6xl font-bold mb-6">
              <span className="text-gradient">
                Quick Installation Guide
              </span>
            </h1>
            <p className="text-xl text-gray-300 max-w-3xl mx-auto leading-relaxed">
              Get KubeStellar up and running quickly with this streamlined installation guide. 
              Follow the prerequisites and installation steps below.
            </p>
          </div>

          {/* Prerequisites Section */}
          <AnimatedCard className="mb-12 p-8">
            <SectionHeader
              icon={<Server size={24} />}
              title="Prerequisites"
              description="Install the required tools based on your use case"
            />
            
            {prerequisiteCategories.map((category, index) => (
              <PrerequisiteCategorySection key={index} category={category} />
            ))}

            {/* Single Guide Button */}
            <div className="text-center">
              <a
                href="https://docs.kubestellar.io/latest/direct/pre-reqs/"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center rounded-lg bg-gradient-to-r from-purple-600/80 to-blue-600/80 px-6 py-3 text-sm font-medium text-white transition-all hover:from-purple-500/80 hover:to-blue-500/80 hover:scale-105 hover:shadow-lg"
              >
                <ExternalLink size={16} className="mr-2" />
                View Detailed Installation Guides
              </a>
            </div>
          </AnimatedCard>

          {/* Automated Prerequisites Check Section */}
          <AnimatedCard className="mb-12 p-8">
            <SectionHeader
              icon={<CheckCircle2 size={24} />}
              title="Automated Check of Prerequisites"
              description="Use this script to automatically verify your system has all required tools"
            />
            
            <div className="mb-6">
              <h3 className="mb-4 text-lg font-medium text-white">
                Run Prerequisites Check:
              </h3>
              <CodeBlock code="curl -fsSL https://raw.githubusercontent.com/kubestellar/kubestellar/refs/heads/main/scripts/check_pre_req.sh | bash" />
            </div>

            <div className="mb-6 rounded-lg border border-blue-500/30 bg-blue-500/10 p-6">
              <div className="flex items-start">
                <Info size={20} className="mr-3 mt-0.5 flex-shrink-0 text-blue-400" />
                <div>
                  <h4 className="mb-3 text-lg font-medium text-blue-300">
                    About the Prerequisites Check Script
                  </h4>
                  <ul className="space-y-2 text-blue-200">
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-blue-400" />
                      Self-contained script suitable for &quot;curl-to-bash&quot; usage
                    </li>
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-blue-400" />
                      Checks for prerequisite presence in your $PATH using the <code className="bg-gray-900/50 px-2 py-1 rounded text-blue-200 text-sm">which</code> command
                    </li>
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-blue-400" />
                      Provides version and path information for present prerequisites
                    </li>
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-blue-400" />
                      Shows installation information for missing prerequisites
                    </li>
                  </ul>
                </div>
              </div>
            </div>

            <div className="rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-6">
              <div className="flex items-start">
                <CheckCircle2 size={20} className="mr-3 mt-0.5 flex-shrink-0 text-emerald-400" />
                <div>
                  <h4 className="mb-3 text-lg font-medium text-emerald-300">
                    For Specific Releases
                  </h4>
                  <p className="text-emerald-200 mb-3">
                    To check prerequisites for a particular KubeStellar release, use the script from that specific release instead of the main branch.
                  </p>
                  <p className="text-emerald-200">
                    <strong>Tip:</strong> Run this check before proceeding with the installation to ensure your system is properly configured.
                  </p>
                </div>
              </div>
            </div>
          </AnimatedCard>

          {/* Platform Installation Section */}
          <AnimatedCard className="p-8">
            <SectionHeader
              icon={<Terminal size={24} />}
              title="KubeStellar Installation"
              description="Choose your platform and run the installation script"
            />
            
            <PlatformSelector platform={platform} setPlatform={setPlatform} />
            
            <div className="mb-6">
              <h3 className="mb-4 text-lg font-medium text-white">
                Installation Script for {platform}:
              </h3>
              <CodeBlock code={platformInstallationScripts[platform]} />
            </div>

            {/* Installation Process Info */}
            <div className="mb-8 rounded-lg border border-blue-500/30 bg-blue-500/10 p-6">
              <div className="flex items-start">
                <Info size={20} className="mr-3 mt-0.5 flex-shrink-0 text-blue-400" />
                <div>
                  <h4 className="mb-3 text-lg font-medium text-blue-300">
                    Installation Process
                  </h4>
                  <ul className="space-y-2 text-blue-200">
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-blue-400" />
                      Creates a local {platform} cluster
                    </li>
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-blue-400" />
                      Installs KubeStellar core components
                    </li>
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-blue-400" />
                      Sets up multi-cluster management capabilities
                    </li>
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-blue-400" />
                      Configures workload distribution
                    </li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Next Steps */}
            <div className="rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-6">
              <div className="flex items-start">
                <CheckCircle2 size={20} className="mr-3 mt-0.5 flex-shrink-0 text-emerald-400" />
                <div>
                  <h4 className="mb-3 text-lg font-medium text-emerald-300">
                    Next Steps
                  </h4>
                  <ul className="space-y-2 text-emerald-200">
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-emerald-400" />
                      Verify installation: <code className="ml-1 bg-gray-900/50 px-2 py-1 rounded text-emerald-200 text-sm">kubectl get namespaces</code>
                    </li>
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-emerald-400" />
                      Check KubeStellar status: <code className="ml-1 bg-gray-900/50 px-2 py-1 rounded text-emerald-200 text-sm">kflex ctx</code>
                    </li>
                    <li className="flex items-start">
                      <ChevronRight size={16} className="mr-2 mt-0.5 flex-shrink-0 text-emerald-400" />
                      Explore the{" "}
                      <a href="https://docs.kubestellar.io/latest/direct/get-started/" target="_blank" rel="noopener noreferrer" className="text-emerald-300 hover:text-emerald-200 underline">
                        documentation
                      </a>{" "}
                      for examples and advanced usage
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </AnimatedCard>

          {/* FAQ Section */}
          <AnimatedCard className="mt-12 p-8">
            <SectionHeader
              icon={<Info size={24} />}
              title="Frequently Asked Questions"
              description="Common questions about KubeStellar installation and setup"
            />
            
            <div className="space-y-4">
              {faqData.map((faq) => (
                <FAQItem key={faq.id} faq={faq} />
              ))}
            </div>
          </AnimatedCard>
        </div>
      </section>

      <Footer />
    </main>
  );
};

export default QuickInstallationPage;