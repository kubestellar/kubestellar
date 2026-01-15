"use client";

import { ReactNode } from 'react';
import { DocsSidebar } from './DocsSidebar';
import { TableOfContents } from './TableOfContents';
import { MobileTOC } from './MobileTOC';
import { MobileHeader } from './MobileSidebarToggle';
import { EditPageLink } from './EditPageLink';
import { useDocsMenu } from './DocsProvider';
import type { ProjectId } from '@/config/versions';

interface TOCItem {
  id: string;
  value: string;
  depth: number;
}

interface PageMapItem {
  name: string;
  route?: string;
  title?: string;
  children?: PageMapItem[];
  frontMatter?: Record<string, unknown>;
  kind?: string;
}

interface Metadata {
  title?: string;
  description?: string;
  [key: string]: unknown;
}

interface DocsLayoutProps {
  children: ReactNode;
  pageMap: PageMapItem[];
  toc?: TOCItem[];
  metadata?: Metadata;
  filePath?: string;
  projectId?: ProjectId;
}

export function DocsLayout({ children, pageMap, toc, metadata, filePath, projectId }: DocsLayoutProps) {
  const { menuOpen, toggleMenu } = useDocsMenu();

  return (
    <div className="flex flex-1 relative">
      {/* Sidebar - Self-contained with all logic */}
      <DocsSidebar pageMap={pageMap} />

      {/* Mobile overlay */}
      {menuOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-20 lg:hidden"
          onClick={toggleMenu}
        />
      )}

      {/* Main content area */}
      <main className="flex-1 min-w-0 lg:ml-0">
        <div className="mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Page header with edit button */}
          <div className="flex items-center justify-between mb-4">
            {/* Mobile Header with Sidebar Toggle - Only visible on mobile/tablet */}
            <div className="flex-1">
              <MobileHeader
                onToggleSidebar={toggleMenu}
                pageTitle={metadata?.title}
              />
            </div>

            {/* Edit page icon - top right */}
            {filePath && projectId && (
              <div className="shrink-0 ml-2">
                <EditPageLink filePath={filePath} projectId={projectId} variant="icon" />
              </div>
            )}
          </div>

          {/* Mobile TOC Accordion - Only visible on mobile/tablet */}
          <MobileTOC toc={toc} />

          {/* Article content */}
          <article className="prose prose-slate dark:prose-invert max-w-none">
            {children}
          </article>
        </div>
      </main>

      {/* Table of Contents - Right sidebar on desktop */}
      <TableOfContents toc={toc} />
    </div>
  );
}
