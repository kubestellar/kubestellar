"use client";

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ChevronRight, ChevronDown, ExternalLink } from 'lucide-react';
import { useSharedConfig } from '@/hooks/useSharedConfig';

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

      {/* Project Links */}
      <div
        className={`
          overflow-hidden transition-all duration-300 ease-in-out
          ${isExpanded ? 'max-h-96 opacity-100' : 'max-h-0 opacity-0'}
        `}
      >
        <ul className="space-y-1 py-1">
          {relatedProjects.map((project) => {
            const isCurrentProject = project.title === currentProject;
            const isExternal = project.href.startsWith('http');

            return (
              <li key={project.title}>
                {isExternal ? (
                  <a
                    href={project.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={`
                      flex items-center gap-2 px-2 py-1.5 text-sm rounded-md transition-colors
                      ${isCurrentProject
                        ? 'bg-blue-50 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 font-medium'
                        : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}
                    `}
                  >
                    <span className="flex-1 truncate">{project.title}</span>
                    <ExternalLink className="w-3 h-3 shrink-0 opacity-50" />
                  </a>
                ) : (
                  <Link
                    href={project.href}
                    className={`
                      flex items-center gap-2 px-2 py-1.5 text-sm rounded-md transition-colors
                      ${isCurrentProject
                        ? 'bg-blue-50 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 font-medium'
                        : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}
                    `}
                  >
                    <span className="flex-1 truncate">{project.title}</span>
                  </Link>
                )}
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}

export default RelatedProjects;
