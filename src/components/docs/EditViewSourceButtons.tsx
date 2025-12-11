"use client";

import React from "react";
import { useTheme } from "next-themes";
import { Eye, Pencil, Bug } from "lucide-react";
import { usePathname } from "next/navigation";

interface EditViewSourceButtonsProps {
  filePath?: string;
  user?: string;
  repo?: string;
  branch?: string;
  docsPath?: string;
}

export default function EditViewSourceButtons({
  filePath,
  user = "kubestellar",
  repo = "kubestellar",
  branch = "main",
  docsPath = "docs/content/",
}: EditViewSourceButtonsProps) {
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = React.useState(false);
  const pathname = usePathname();

  React.useEffect(() => {
    setMounted(true);
  }, []);

  const isDark = resolvedTheme === "dark";

  // Determine the issue repository
  const issueRepo = filePath ? repo : "docs";
  
  // Generate URLs
  const fullPath = filePath ? `${docsPath}${filePath}` : null;
  const viewUrl = fullPath ? `https://github.com/${user}/${repo}/blob/${branch}/${fullPath}` : null;
  const editUrl = fullPath ? `https://github.com/${user}/${repo}/edit/${branch}/${fullPath}` : null;
  
  // Generate issue URL with pre-filled title and body
  const pageTitle = pathname ? pathname.split('/').filter(Boolean).pop()?.replace(/-/g, ' ') : 'page';
  const issueTitle = encodeURIComponent(`Issue with page: ${pageTitle}`);
  const issueBody = encodeURIComponent(`Page URL: ${typeof window !== 'undefined' ? window.location.href : pathname}\n\n**Describe the issue:**\n\n`);
  const issueUrl = `https://github.com/${user}/${issueRepo}/issues/new?title=${issueTitle}&body=${issueBody}`;

  // Separate conditional classes for better maintainability
  const textColorClasses = isDark
    ? "text-gray-400 hover:text-gray-200"
    : "text-gray-600 hover:text-gray-900";
  const bgColorClasses = isDark
    ? "hover:bg-neutral-800/50"
    : "hover:bg-gray-100/70";
  const buttonBaseClasses = `inline-flex items-center gap-1.5 px-2 py-1 text-xs rounded-md transition-all duration-200 ${textColorClasses} ${bgColorClasses}`;

  // Show a minimal skeleton during hydration to prevent content flash
  if (!mounted) {
    return (
      <div className="flex items-center gap-2 mb-4 not-prose">
        {fullPath && (
          <>
            <div className="inline-flex items-center gap-1.5 px-2 py-1 h-6 w-24 rounded-md bg-gray-200/50 dark:bg-neutral-800/50 animate-pulse" />
            <div className="inline-flex items-center gap-1.5 px-2 py-1 h-6 w-20 rounded-md bg-gray-200/50 dark:bg-neutral-800/50 animate-pulse" />
          </>
        )}
        <div className="inline-flex items-center gap-1.5 px-2 py-1 h-6 w-24 rounded-md bg-gray-200/50 dark:bg-neutral-800/50 animate-pulse" />
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2 mb-4 not-prose">
      {viewUrl && (
        <a
          href={viewUrl}
          target="_blank"
          rel="noopener noreferrer"
          className={buttonBaseClasses}
          title="View source on GitHub"
          aria-label="View source on GitHub"
        >
          <Eye className="w-3.5 h-3.5" />
          <span className="font-medium">View source</span>
        </a>
      )}
      {editUrl && (
        <a
          href={editUrl}
          target="_blank"
          rel="noopener noreferrer"
          className={buttonBaseClasses}
          title="Edit this page on GitHub"
          aria-label="Edit this page on GitHub"
        >
          <Pencil className="w-3.5 h-3.5" />
          <span className="font-medium">Edit page</span>
        </a>
      )}
      <a
        href={issueUrl}
        target="_blank"
        rel="noopener noreferrer"
        className={buttonBaseClasses}
        title="Report an issue with this page"
        aria-label="Report an issue with this page"
      >
        <Bug className="w-3.5 h-3.5" />
        <span className="font-medium">Open issue</span>
      </a>
    </div>
  );
}
