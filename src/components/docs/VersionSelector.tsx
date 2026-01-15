"use client";

import { useState, useRef, useEffect } from 'react';
import { usePathname } from 'next/navigation';
import { useTheme } from 'next-themes';
import {
  getProjectFromPath,
  getProjectVersions as getStaticProjectVersions,
  getVersionUrl,
} from '@/config/versions';
import { useSharedConfig, getVersionsForProject } from '@/hooks/useSharedConfig';

interface VersionSelectorProps {
  className?: string;
  isMobile?: boolean;
}

// Detect current version from hostname
function detectCurrentVersionKey(versions: Array<{ key: string; branch: string; isDev?: boolean }>): string {
  if (typeof window === 'undefined') return 'latest';

  const hostname = window.location.hostname;

  // Production site = latest
  if (hostname === 'kubestellar.io' || hostname === 'www.kubestellar.io') {
    return 'latest';
  }

  // Netlify branch deploys: {branch-slug}--{site-name}.netlify.app
  const branchDeployMatch = hostname.match(/^(.+)--[\w-]+\.netlify\.app$/);
  if (branchDeployMatch) {
    const branchSlug = branchDeployMatch[1];

    // Check for "main" branch deploy
    if (branchSlug === 'main') {
      return 'main';
    }

    // Check for deploy previews (deploy-preview-XXX)
    if (branchSlug.startsWith('deploy-preview-')) {
      return 'latest'; // Preview shows as latest
    }

    // Match branch slug to version (e.g., docs-0-28-0 -> 0.28.0)
    for (const version of versions) {
      const versionBranchSlug = version.branch.replace(/\//g, '-').replace(/\./g, '-');
      if (branchSlug === versionBranchSlug) {
        return version.key;
      }
    }
  }

  // Default to latest
  return 'latest';
}

export function VersionSelector({ className = '', isMobile = false }: VersionSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [currentVersionKey, setCurrentVersionKey] = useState('latest');
  const dropdownRef = useRef<HTMLDivElement>(null);
  const pathname = usePathname();
  const { resolvedTheme } = useTheme();
  const isDark = resolvedTheme === 'dark';

  // Fetch dynamic config from production (always shows latest versions)
  const { config: sharedConfig } = useSharedConfig();

  // Detect current project from URL
  const currentProject = getProjectFromPath(pathname);
  const projectId = currentProject.id;

  // Get versions - prefer dynamic config, fall back to static
  const rawVersions = sharedConfig
    ? getVersionsForProject(sharedConfig, projectId)
    : getStaticProjectVersions(projectId);

  // Sort versions: default first, then dev, then rest by version number descending
  const versions = rawVersions.sort((a, b) => {
    if (a.isDefault && !b.isDefault) return -1;
    if (!a.isDefault && b.isDefault) return 1;
    if (a.isDev && !b.isDev) return -1;
    if (!a.isDev && b.isDev) return 1;
    // Sort by version number descending (legacy goes last)
    if (a.key === 'legacy') return 1;
    if (b.key === 'legacy') return -1;
    return b.key.localeCompare(a.key, undefined, { numeric: true });
  });

  // Detect current version from hostname on mount
  useEffect(() => {
    setCurrentVersionKey(detectCurrentVersionKey(versions));
  }, [versions]);

  // Get the label for the current version
  const currentVersion = versions.find(v => v.key === currentVersionKey);
  const currentVersionLabel = currentVersion?.label || `v${currentProject.currentVersion}`;

  // Show project name for non-KubeStellar projects
  const showProjectName = projectId !== 'kubestellar';

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Close dropdown on escape key
  useEffect(() => {
    function handleEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    }

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, []);

  const handleVersionChange = (versionKey: string) => {
    setIsOpen(false);

    // Get the URL for the selected version (project-aware)
    const url = getVersionUrl(versionKey, pathname, projectId);

    // Navigate to the new version
    window.location.href = url;
  };

  if (isMobile) {
    // Mobile version - expandable list
    return (
      <div className={className}>
        <button
          onClick={() => setIsOpen(!isOpen)}
          className={`flex items-center justify-between w-full px-3 py-2 text-sm rounded-md transition-colors ${
            isDark
              ? 'text-gray-300 hover:bg-neutral-800'
              : 'text-gray-700 hover:bg-gray-100'
          }`}
        >
          <span className="flex items-center">
            <svg className="w-4 h-4 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
            </svg>
            {showProjectName && <span className="font-medium mr-1">{currentProject.name}:</span>}
            Version: {currentVersionLabel}
          </span>
          <svg
            className={`w-4 h-4 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        {isOpen && (
          <div className="mt-1 ml-7 space-y-1">
            {versions.map(({ key, label, externalUrl }) => {
              const isCurrentVersion = key === currentVersionKey;
              const isExternal = !!externalUrl;
              return (
                <button
                  key={key}
                  onClick={() => handleVersionChange(key)}
                  className={`block w-full text-left px-3 py-2 text-sm rounded-md transition-colors ${
                    isCurrentVersion
                      ? isDark
                        ? 'bg-neutral-800 text-white font-medium'
                        : 'bg-gray-100 text-gray-900 font-medium'
                      : isDark
                        ? 'text-gray-400 hover:bg-neutral-800 hover:text-gray-200'
                        : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                  }`}
                >
                  {label}
                  {isExternal && (
                    <svg className="w-3 h-3 ml-1 inline-block" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                    </svg>
                  )}
                </button>
              );
            })}
          </div>
        )}
      </div>
    );
  }

  // Desktop version - dropdown
  return (
    <div className={`relative ${className}`} ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-1 text-xs font-mono px-2 py-1.5 rounded-md transition-colors cursor-pointer ${
          isDark
            ? 'text-gray-400 bg-neutral-800/50 hover:bg-neutral-700 hover:text-gray-200'
            : 'text-gray-600 bg-gray-100 hover:bg-gray-200 hover:text-gray-900'
        }`}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        aria-label={`Select ${currentProject.name} documentation version`}
      >
        {showProjectName && <span className="font-semibold">{currentProject.name}</span>}
        <span>{currentVersionLabel}</span>
        <svg
          className={`w-3 h-3 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div
          className={`absolute top-full right-0 mt-1 w-44 max-h-80 overflow-y-auto rounded-md shadow-lg border z-50 ${
            isDark
              ? 'bg-neutral-900 border-neutral-700'
              : 'bg-white border-gray-200'
          }`}
          role="listbox"
          aria-label={`${currentProject.name} documentation versions`}
        >
          <div className="py-1">
            {versions.map(({ key, label, externalUrl }) => {
              const isCurrentVersion = key === currentVersionKey;
              const isExternal = !!externalUrl;

              return (
                <button
                  key={key}
                  onClick={() => handleVersionChange(key)}
                  className={`flex items-center justify-between w-full text-left px-3 py-2 text-sm transition-colors ${
                    isCurrentVersion
                      ? isDark
                        ? 'bg-neutral-800 text-white font-medium'
                        : 'bg-gray-100 text-gray-900 font-medium'
                      : isDark
                        ? 'text-gray-300 hover:bg-neutral-800'
                        : 'text-gray-700 hover:bg-gray-50'
                  }`}
                  role="option"
                  aria-selected={isCurrentVersion}
                >
                  <span>{label}</span>
                  {isExternal && (
                    <svg className="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                    </svg>
                  )}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

export default VersionSelector;
