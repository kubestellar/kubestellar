"use client";

import { useState, useEffect } from 'react';

// Production URL for fetching shared config
const PRODUCTION_CONFIG_URL = 'https://kubestellar.io/config/shared.json';

// Type definitions
export interface VersionInfo {
  label: string;
  branch: string;
  isDefault: boolean;
  externalUrl?: string;
  isDev?: boolean;
}

export interface ProjectInfo {
  name: string;
  basePath: string;
  currentVersion: string;
}

export interface RelatedProject {
  title: string;
  href: string;
  description?: string;
}

export interface SharedConfig {
  versions: Record<string, Record<string, VersionInfo>>;
  projects: Record<string, ProjectInfo>;
  relatedProjects: RelatedProject[];
  editBaseUrls: Record<string, string>;
  updatedAt: string;
}

// Cache for the config to avoid repeated fetches
let configCache: SharedConfig | null = null;
let fetchPromise: Promise<SharedConfig | null> | null = null;

async function fetchConfig(): Promise<SharedConfig | null> {
  // Return cached config if available
  if (configCache) {
    return configCache;
  }

  // Return existing fetch promise if in progress
  if (fetchPromise) {
    return fetchPromise;
  }

  fetchPromise = (async () => {
    try {
      // Try production URL first (works for all branch deploys)
      const res = await fetch(PRODUCTION_CONFIG_URL, {
        cache: 'no-cache', // Allow caching with revalidation to reduce network requests
        headers: {
          'Accept': 'application/json',
        },
      });
      if (res.ok) {
        configCache = await res.json();
        return configCache;
      }
    } catch (e) {
      console.warn('Failed to fetch config from production:', e);
    }

    try {
      // Fallback to local config (for local dev or if production unreachable)
      const res = await fetch('/config/shared.json');
      if (res.ok) {
        configCache = await res.json();
        return configCache;
      }
    } catch (e) {
      console.warn('Failed to fetch local config:', e);
    }

    return null;
  })();

  return fetchPromise;
}

export function useSharedConfig() {
  const [config, setConfig] = useState<SharedConfig | null>(configCache);
  const [loading, setLoading] = useState(!configCache);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    // If already have cached config, use it immediately
    if (configCache) {
      setConfig(configCache);
      setLoading(false);
      return;
    }

    let mounted = true;

    fetchConfig()
      .then((data) => {
        if (mounted) {
          setConfig(data);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (mounted) {
          setError(err);
          setLoading(false);
        }
      });

    return () => {
      mounted = false;
    };
  }, []);

  return { config, loading, error };
}

// Utility functions that work with the shared config
export function getVersionsForProject(
  config: SharedConfig | null,
  projectId: string
): Array<{ key: string } & VersionInfo> {
  if (!config || !config.versions[projectId]) {
    return [];
  }
  return Object.entries(config.versions[projectId]).map(([key, value]) => ({
    key,
    ...value,
  }));
}

export function getProjectInfo(
  config: SharedConfig | null,
  projectId: string
): ProjectInfo | null {
  if (!config || !config.projects[projectId]) {
    return null;
  }
  return config.projects[projectId];
}

export function getEditUrl(
  config: SharedConfig | null,
  projectId: string,
  filePath: string
): string | null {
  if (!config || !config.editBaseUrls[projectId]) {
    return null;
  }
  // Remove leading slash if present
  const cleanPath = filePath.startsWith('/') ? filePath.slice(1) : filePath;
  return `${config.editBaseUrls[projectId]}/${cleanPath}`;
}

// Export the fetch function for server-side usage
export { fetchConfig };
