"use client";

import React from "react";
import { useTheme } from "next-themes";
import { Eye, Pencil } from "lucide-react";

interface EditViewSourceButtonsProps {
  filePath: string;
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

  React.useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return null;
  }

  const isDark = resolvedTheme === "dark";

  const fullPath = `${docsPath}${filePath}`;
  const viewUrl = `https://github.com/${user}/${repo}/blob/${branch}/${fullPath}`;
  const editUrl = `https://github.com/${user}/${repo}/edit/${branch}/${fullPath}`;

  const buttonBaseClasses = `inline-flex items-center gap-1.5 px-2 py-1 text-xs rounded-md transition-all duration-200 ${
    isDark
      ? "text-gray-400 hover:text-gray-200 hover:bg-neutral-800/50"
      : "text-gray-600 hover:text-gray-900 hover:bg-gray-100/70"
  }`;

  return (
    <div className="flex items-center gap-2 mb-4 not-prose">
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
    </div>
  );
}
