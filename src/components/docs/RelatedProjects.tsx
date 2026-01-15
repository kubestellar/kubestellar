"use client";

import { useState } from 'react';
import { usePathname } from 'next/navigation';
import { ChevronRight, ChevronDown } from 'lucide-react';
import { useSharedConfig } from '@/hooks/useSharedConfig';

// Production URL - all cross-project links go here
const PRODUCTION_URL = 'https://kubestellar.io';

// Static fallback for related projects
const STATIC_RELATED_PROJECTS = [
  { title: 'KubeStellar', href: '/docs/what-is-kubestellar/overview', description: 'Multi-cluster configuration management' },
  { title: 'kubectl-claude', href: '/docs/kubectl-claude/overview/introduction', description: 'AI-powered kubectl plugin' },
  { title: 'KubeFlex', href: '/docs/kubeflex/overview/introduction', description: 'Lightweight Kubernetes control planes' },
  { title: 'A2A', href: '/docs/a2a/overview/introduction', description: 'Agent-to-Agent protocol' },
  { title: 'Multi Plugin', href: '/docs/multi-plugin/overview/introduction', description: 'Multi-cluster kubectl plugin' },
];

interface RelatedProjectsProps {
  variant?: 'full' | 'slim';
}

export function RelatedProjects({ variant = 'full' }: RelatedProjectsProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const pathname = usePathname();
  const { config } = useSharedConfig();

  // Get related projects from config or fallback
  const relatedProjects = config?.relatedProjects ?? STATIC_RELATED_PROJECTS;

  // Don't render in slim mode
  if (variant === 'slim') return null;

  // Determine current project from pathname
  const getCurrentProject = () => {
    if (pathname.startsWith('/docs/a2a')) return 'A2A';
    if (pathname.startsWith('/docs/kubeflex')) return 'KubeFlex';
    if (pathname.startsWith('/docs/multi-plugin')) return 'Multi Plugin';
    if (pathname.startsWith('/docs/kubectl-claude')) return 'kubectl-claude';
    return 'KubeStellar';
  };

  const currentProject = getCurrentProject();

  // Check if we're on production or a branch deploy
  const isProduction = typeof window !== 'undefined' && 
    (window.location.hostname === 'kubestellar.io' || 
     window.location.hostname === 'www.kubestellar.io' ||
     window.location.hostname === 'localhost');

  // Get the full URL for a project link
  // On branch deploys, use absolute URL to production for cross-project links
  const getProjectUrl = (href: string, projectTitle: string) => {
    // If we're on production or localhost, use relative links
    if (isProduction) {
      return href;
    }
    // On branch deploys, always link to production for all projects
    // This ensures users can navigate between projects even on old version branches
    return `${PRODUCTION_URL}${href}`;
  };

  return (
    <div className="shrink-0 py-2 px-4 border-t border-gray-200 dark:border-gray-700">
      {/* Header - clickable to toggle */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex items-center justify-between w-full py-2 text-xs font-semibold uppercase tracking-wider transition-colors text-gray-500 dark:text-gray-400"
      >
        <span>Related Projects</span>
        <span className="ml-auto">
          {isExpanded ? (
            <ChevronDown className="w-3 h-3" />
          ) : (
            <ChevronRight className="w-3 h-3" />
          )}
        </span>
      </button>

      {/* Project links */}
      <div
        className={`
          space-y-1 overflow-hidden transition-all duration-200 ease-in-out
          ${isExpanded ? 'max-h-96 opacity-100 pb-2' : 'max-h-0 opacity-0'}
        `}
      >
        {relatedProjects.map((project: { title: string; href: string; description?: string }) => {
          const isCurrentProject = project.title === currentProject;
          const projectUrl = getProjectUrl(project.href, project.title);
          const isExternal = projectUrl.startsWith('http');

          return (
            <a
              key={project.title}
              href={projectUrl}
              className={`
                block px-3 py-1.5 text-sm rounded-md transition-colors
                ${isCurrentProject
                  ? 'bg-blue-50 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 font-medium'
                  : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
                }
              `}
              {...(isExternal ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
            >
              {project.title}
            </a>
          );
        })}
      </div>
    </div>
  );
}
