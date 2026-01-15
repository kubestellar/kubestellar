"use client";

import { useState, useEffect } from 'react';
import { usePathname } from 'next/navigation';
import { useTheme } from 'next-themes';
import { ChevronRight, ChevronDown, Moon, Sun, PanelRightOpenIcon, PanelLeftOpen } from 'lucide-react';
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
  onCollapse?: () => void;
  isMobile?: boolean;
  bannerActive?: boolean;
}

export function RelatedProjects({ variant = 'full', onCollapse, isMobile = false, bannerActive = false }: RelatedProjectsProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const [mounted, setMounted] = useState(false);
  const pathname = usePathname();
  const { config } = useSharedConfig();
  const { resolvedTheme, setTheme } = useTheme();

  useEffect(() => {
    setMounted(true);
  }, []);

  const isDark = mounted && resolvedTheme === 'dark';

  // Get related projects from config or fallback
  const relatedProjects = config?.relatedProjects ?? STATIC_RELATED_PROJECTS;

  // Slim variant - icon-only vertical layout
  if (variant === 'slim') {
    if (!mounted) {
      return (
        <div className="shrink-0 flex flex-col items-center gap-2 py-4 min-w-16">
          <div className="w-5 h-5" />
          <div className="w-5 h-5" />
        </div>
      );
    }

    return (
      <div
        className="shrink-0 sticky flex flex-col items-center gap-2 py-4 min-w-16 border-t border-gray-200 dark:border-gray-700"
        suppressHydrationWarning
      >
        {/* Theme Toggle Icon */}
        <button
          onClick={() => setTheme(isDark ? 'light' : 'dark')}
          title="Change theme"
          className="group p-2 rounded-md hover:font-bold transition-all"
          style={{
            color: isDark ? '#f3f4f6' : '#111827',
          }}
          suppressHydrationWarning
        >
          <div className="relative w-5 h-5">
            <Moon
              className={`absolute inset-0 w-5 h-5 transition-all duration-300 group-hover:rotate-45 ${
                isDark ? 'opacity-100 rotate-0 scale-100' : 'opacity-0 -rotate-90 scale-0'
              }`}
            />
            <Sun
              className={`absolute inset-0 w-5 h-5 transition-all duration-300 group-hover:rotate-45 ${
                !isDark ? 'opacity-100 rotate-0 scale-100' : 'opacity-0 rotate-90 scale-0'
              }`}
            />
          </div>
        </button>

        {/* Expand Sidebar Icon */}
        {onCollapse && (
          <button
            onClick={onCollapse}
            title="Expand sidebar"
            className="p-2 rounded-md hover:font-bold transition-all"
            style={{
              color: isDark ? '#f3f4f6' : '#111827',
            }}
            suppressHydrationWarning
          >
            <PanelLeftOpen className="w-5 h-5" />
          </button>
        )}
      </div>
    );
  }

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
  const getProjectUrl = (href: string) => {
    if (isProduction) {
      return href;
    }
    return `${PRODUCTION_URL}${href}`;
  };

  return (
    <div className={`shrink-0 px-4 border-t border-gray-200 dark:border-gray-700 ${bannerActive ? 'py-1' : 'py-2'}`}>
      {/* Header - clickable to toggle */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className={`flex items-center justify-between w-full text-xs font-semibold uppercase tracking-wider transition-colors text-gray-500 dark:text-gray-400 ${bannerActive ? 'py-1' : 'py-2'}`}
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
          overflow-hidden transition-all duration-200 ease-in-out
          ${isExpanded ? 'max-h-96 opacity-100' : 'max-h-0 opacity-0'}
          ${bannerActive ? 'space-y-0' : 'space-y-1 pb-2'}
        `}
      >
        {relatedProjects.map((project: { title: string; href: string; description?: string }) => {
          const isCurrentProject = project.title === currentProject;
          const projectUrl = getProjectUrl(project.href);

          return (
            <a
              key={project.title}
              href={projectUrl}
              className={`
                block px-3 text-sm rounded-md transition-colors
                ${bannerActive ? 'py-0.5' : 'py-1.5'}
                ${isCurrentProject
                  ? 'bg-blue-50 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 font-medium'
                  : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
                }
              `}
            >
              {project.title}
            </a>
          );
        })}
      </div>

      {/* Footer Controls */}
      {mounted && (
        <div
          className={`flex items-center gap-2 border-t border-gray-200 dark:border-gray-700 ${bannerActive ? 'pt-2 mt-1' : 'pt-3 mt-2'}`}
          suppressHydrationWarning
        >
          {/* Theme Toggle Button */}
          <button
            onClick={() => setTheme(isDark ? 'light' : 'dark')}
            title="Change theme"
            className="group cursor-pointer h-7 rounded-md px-2 text-sm font-thin transition-all hover:font-bold flex items-center gap-2 flex-1"
            style={{
              color: isDark ? '#f3f4f6' : '#111827',
            }}
            suppressHydrationWarning
          >
            <div className="relative w-5 h-5">
              <Moon
                className={`absolute inset-0 w-5 h-5 transition-all duration-300 group-hover:rotate-45 ${
                  isDark ? 'opacity-100 rotate-0 scale-100' : 'opacity-0 -rotate-90 scale-0'
                }`}
              />
              <Sun
                className={`absolute inset-0 w-5 h-5 transition-all duration-300 group-hover:rotate-45 ${
                  !isDark ? 'opacity-100 rotate-0 scale-100' : 'opacity-0 rotate-90 scale-0'
                }`}
              />
            </div>
            <span>{isDark ? 'Dark' : 'Light'}</span>
          </button>

          {/* Collapse Sidebar Button - Hidden on mobile */}
          {onCollapse && !isMobile && (
            <button
              onClick={onCollapse}
              className="transition-all cursor-pointer rounded-md p-2 hover:font-bold"
              style={{
                color: isDark ? '#f3f4f6' : '#111827',
              }}
              title="Collapse sidebar"
              type="button"
              suppressHydrationWarning
            >
              <PanelRightOpenIcon className="w-4 h-4" />
            </button>
          )}
        </div>
      )}
    </div>
  );
}
