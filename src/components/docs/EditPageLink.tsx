"use client";

import { useSharedConfig, getVersionsForProject, VersionInfo } from '@/hooks/useSharedConfig';
import { getProjectVersions as getStaticProjectVersions } from '@/config/versions';
import type { ProjectId } from '@/config/versions';

interface EditPageLinkProps {
  filePath: string;
  projectId: ProjectId;
  variant?: 'full' | 'icon';
}

// Version entry with key from the versions config
type VersionEntry = { key: string } & VersionInfo;

// Convert branch name to Netlify slug format (e.g., docs/0.29.0 -> docs-0-29-0)
function branchToSlug(branch: string): string {
  return branch.replace(/\//g, '-').replace(/\./g, '-');
}

// Detect current branch from hostname (for kubestellar docs repo)
function detectCurrentBranch(versions: VersionEntry[]): string {
  if (typeof window === 'undefined') return 'main';

  const hostname = window.location.hostname;

  // Production site uses the "latest" version's branch
  if (hostname === 'kubestellar.io' || hostname === 'www.kubestellar.io') {
    const latestVersion = versions.find(v => v.key === 'latest');
    return latestVersion?.branch || 'main';
  }

  // Netlify branch deploys: {branch-slug}--{site-name}.netlify.app
  const branchDeployMatch = hostname.match(/^(.+)--[\w-]+\.netlify\.app$/);
  if (branchDeployMatch) {
    const branchSlug = branchDeployMatch[1];

    // Main branch deploy
    if (branchSlug === 'main') {
      return 'main';
    }

    // Deploy previews go to main
    if (branchSlug.startsWith('deploy-preview-')) {
      return 'main';
    }

    // Match branch slug to version branch (e.g., docs-0-29-0 -> docs/0.29.0)
    for (const version of versions) {
      if (branchSlug === branchToSlug(version.branch)) {
        return version.branch;
      }
    }
  }

  return 'main';
}

// Source repos for each project (used when on main branch)
const SOURCE_REPOS: Record<string, { repo: string; docsPath: string }> = {
  a2a: { repo: 'kubestellar/a2a', docsPath: 'docs' },
  kubeflex: { repo: 'kubestellar/kubeflex', docsPath: 'docs' },
  'multi-plugin': { repo: 'kubestellar/kubectl-multi-plugin', docsPath: 'docs' },
  'kubectl-claude': { repo: 'kubestellar/kubectl-claude', docsPath: 'docs' },
};

// Build edit URL for a project, using correct branch
function buildEditBaseUrl(projectId: ProjectId, branch: string): string {
  // KubeStellar docs always live in docs repo
  if (projectId === 'kubestellar') {
    return `https://github.com/kubestellar/docs/edit/${branch}/docs/content`;
  }

  // For other projects: version branches are in docs repo, main goes to source repo
  if (branch !== 'main' && branch.startsWith('docs/')) {
    // Version branch in docs repo (e.g., docs/kubectl-claude/0.4.6)
    return `https://github.com/kubestellar/docs/edit/${branch}/docs/content/${projectId}`;
  }

  // Main branch - link to source repo
  const source = SOURCE_REPOS[projectId];
  if (source) {
    return `https://github.com/${source.repo}/edit/main/${source.docsPath}`;
  }

  return '';
}

// Validate that URL is a safe GitHub edit URL to prevent XSS
function isValidGitHubEditUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    // Only allow https GitHub URLs with /edit/ path
    return (
      parsed.protocol === 'https:' &&
      parsed.hostname === 'github.com' &&
      parsed.pathname.includes('/edit/')
    );
  } catch {
    return false;
  }
}

export function EditPageLink({ filePath, projectId, variant = 'full' }: EditPageLinkProps) {
  const { config } = useSharedConfig();

  // Get versions to detect current branch
  const versions = config
    ? getVersionsForProject(config, projectId)
    : getStaticProjectVersions(projectId);

  // Detect current branch from hostname
  const currentBranch = detectCurrentBranch(versions);

  // Build edit URL with correct branch
  const editBaseUrl = buildEditBaseUrl(projectId, currentBranch);

  if (!editBaseUrl) return null;

  // Sanitize filePath to prevent path traversal
  const sanitizedFilePath = filePath.replace(/\.\./g, '').replace(/^\/+/, '');

  // Construct the full edit URL
  const editUrl = `${editBaseUrl}/${sanitizedFilePath}`;

  // Validate URL before rendering to prevent XSS
  if (!isValidGitHubEditUrl(editUrl)) return null;

  // Use validated URL object to construct safe href
  // CodeQL: URL is validated above to only allow https://github.com with /edit/ path
  const safeUrl = new URL(editUrl);

  // Pencil icon SVG
  const PencilIcon = () => (
    <svg
      className={variant === 'icon' ? 'w-5 h-5' : 'w-4 h-4'}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
      />
    </svg>
  );

  // Icon-only variant for top right placement
  if (variant === 'icon') {
    return (
      <a
        href={safeUrl.href}
        target="_blank"
        rel="noopener noreferrer"
        title="Edit this page on GitHub"
        className="p-2 rounded-md text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
      >
        <PencilIcon />
      </a>
    );
  }

  // Full variant (original) for bottom of page
  return (
    <div className="mt-12 pt-6 border-t border-gray-200 dark:border-neutral-700">
      <a
        href={safeUrl.href}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 transition-colors"
      >
        <PencilIcon />
        Edit this page on GitHub
      </a>
    </div>
  );
}

export default EditPageLink;
