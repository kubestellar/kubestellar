"use client";

import { useSharedConfig } from '@/hooks/useSharedConfig';
import type { ProjectId } from '@/config/versions';

interface EditPageLinkProps {
  filePath: string;
  projectId: ProjectId;
  variant?: 'full' | 'icon';
}

// Static edit base URLs (fallback if shared config unavailable)
const STATIC_EDIT_BASE_URLS: Record<string, string> = {
  kubestellar: 'https://github.com/kubestellar/docs/edit/main/docs/content',
  a2a: 'https://github.com/kubestellar/a2a/edit/main/docs',
  kubeflex: 'https://github.com/kubestellar/kubeflex/edit/main/docs',
  'multi-plugin': 'https://github.com/kubestellar/kubectl-multi-plugin/edit/main/docs',
  'kubectl-claude': 'https://github.com/kubestellar/kubectl-claude/edit/main/docs',
};

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

  // Get edit base URL from config or fallback
  const editBaseUrl = config?.editBaseUrls?.[projectId] ?? STATIC_EDIT_BASE_URLS[projectId];

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
