"use client";

import React from "react";
import { useTheme } from "next-themes";
import { Eye, Pencil} from "lucide-react";

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
  repo = "docs",  
  branch = "main",
  docsPath = "docs/content/",
}: EditViewSourceButtonsProps) {
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = React.useState(false);

  React.useEffect(() => {
    setMounted(true);
  }, []);

  const isDark = resolvedTheme === "dark";

  
  // Generate URLs
  const fullPath = filePath ? `${docsPath}${filePath}` : null;
  const viewUrl = fullPath ? `https://github.com/${user}/${repo}/blob/${branch}/${fullPath}` : null;
  const editUrl = fullPath ? `https://github.com/${user}/${repo}/edit/${branch}/${fullPath}` : null;


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
    </div>
  );
}
