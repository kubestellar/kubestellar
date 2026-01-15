"use client";

import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useTheme } from 'next-themes';
import { ChevronRight, ChevronDown, FileText } from 'lucide-react';
import { RelatedProjects } from './RelatedProjects';
import { useDocsMenu } from './DocsProvider';

interface MenuItem {
  name: string;
  route?: string;
  title?: string;
  children?: MenuItem[];
  frontMatter?: Record<string, unknown>;
  kind?: string;
  theme?: { collapsed?: boolean };
}

interface DocsSidebarProps {
  pageMap: MenuItem[];
  className?: string;
}

export function DocsSidebar({ pageMap, className }: DocsSidebarProps) {
  const pathname = usePathname();
  const sidebarRef = useRef<HTMLElement>(null);
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const {
    sidebarCollapsed,
    toggleSidebar,
    menuOpen,
    bannerDismissed,
    navCollapsed: collapsed,
    setNavCollapsed: setCollapsed,
    toggleNavCollapsed,
    navInitialized
  } = useDocsMenu();

  useEffect(() => {
    setMounted(true);
  }, []);

  const isDark = mounted && resolvedTheme === 'dark';
  // Text colors based on theme
  const textColor = isDark ? '#e5e7eb' : '#374151'; // gray-200 : gray-700
  // Stable layout values - only recalculate on resize or banner change
  const [layoutValues, setLayoutValues] = useState({ top: '4rem', height: 'calc(100vh - 4rem)' });
  const layoutCalculatedRef = useRef(false);

  useEffect(() => {
    const calculateOffsets = () => {
      const navbar = document.querySelector('.nextra-nav-container');
      if (navbar) {
        const navbarHeight = (navbar as HTMLElement).offsetHeight;
        const newTop = `${navbarHeight}px`;
        const newHeight = `calc(100vh - ${navbarHeight}px)`;
        setLayoutValues({ top: newTop, height: newHeight });
        layoutCalculatedRef.current = true;
      }
    };

    // Calculate on mount and when banner state changes
    // Use setTimeout to allow DOM to update after banner dismiss
    const timeoutId = setTimeout(calculateOffsets, 50);

    window.addEventListener('resize', calculateOffsets);
    return () => {
      clearTimeout(timeoutId);
      window.removeEventListener('resize', calculateOffsets);
    };
  }, [bannerDismissed]);

  // Store initial pathname for initialization
  const initialPathnameRef = useRef(pathname);

  // Initialize collapsed state once on mount - collapse folders not in path to active
  // After initialization, user controls expand/collapse manually
  useEffect(() => {
    if (navInitialized.current) return;
    navInitialized.current = true;

    const initialCollapsed = new Set<string>();
    const pathToActive = new Set<string>();
    const currentPath = initialPathnameRef.current;

    // Find the path to the active item
    function findActivePath(items: MenuItem[], parentKey: string = ''): boolean {
      for (const item of items) {
        const itemKey = parentKey ? `${parentKey}-${item.name}` : item.name;
        const isActive = item.route && currentPath === item.route;

        if (isActive) {
          return true;
        }

        if (item.children) {
          const childActive = findActivePath(item.children, itemKey);
          if (childActive) {
            pathToActive.add(itemKey);
            return true;
          }
        }
      }
      return false;
    }

    // Collapse all folders except those in the active path or with theme.collapsed: false
    function collapseAll(items: MenuItem[], parentKey: string = '') {
      for (const item of items) {
        const itemKey = parentKey ? `${parentKey}-${item.name}` : item.name;
        const hasChildren = item.children && item.children.length > 0;

        if (hasChildren) {
          const shouldStayExpanded = item.theme?.collapsed === false;
          if (!pathToActive.has(itemKey) && !shouldStayExpanded) {
            initialCollapsed.add(itemKey);
          }
          if (item.children) {
            collapseAll(item.children, itemKey);
          }
        }
      }
    }

    findActivePath(pageMap);
    collapseAll(pageMap);
    setCollapsed(initialCollapsed);
  }, [pageMap]);

  const toggleCollapse = (itemKey: string) => {
    toggleNavCollapsed(itemKey);
  };

  const renderMenuItem = (item: MenuItem, depth: number = 0, parentKey: string = '') => {
    const hasChildren = item.children && item.children.length > 0;
    const itemKey = parentKey ? `${parentKey}-${item.name}` : item.name;
    const isCollapsed = collapsed.has(itemKey);
    const isActive = item.route && pathname === item.route;
    const displayTitle = item.title || item.name;

    // Skip separator, meta items, and items without title/name
    if (item.kind === 'Separator' || item.kind === 'Meta' || !displayTitle || displayTitle.trim() === '') {
      return null;
    }
    
    // Skip index files and hidden items
    if (item.name === 'index' || item.name === '_meta' || item.route === '#') {
      return null;
    }

    return (
      <div key={itemKey} className="relative space-y-1">
        <div className="flex items-center group relative">
          {/* Vertical line for nested items */}
          {depth > 0 && (
            <div 
              className="absolute left-0 top-0 bottom-0 w-px bg-gray-200 dark:bg-gray-700"
              style={{ left: `${(depth - 1) * 16 + 20}px` }}
            />
          )}
          
          {/* Folder or Page */}
          {hasChildren ? (
            // Folder - clickable to toggle
            <button
              onClick={() => toggleCollapse(itemKey)}
              className="flex-1 flex items-start gap-2 px-3 py-2 text-sm font-thin hover:font-semibold rounded-lg transition-all text-left w-full relative z-10"
              style={{ paddingLeft: `${depth * 16 + 12}px`, color: textColor }}
            >
              <span className="flex-1 wrap-break-word">{displayTitle}</span>
              <span className="ml-auto shrink-0 mt-0.5">
                {isCollapsed ? (
                  <ChevronRight className="w-4 h-4 transition-all duration-200" />
                ) : (
                  <ChevronDown className="w-4 h-4 transition-all duration-200" />
                )}
              </span>
            </button>
          ) : (
            // Page - clickable link with icon
            <Link
              href={item.route || '#'}
              className={`
                flex-1 flex items-start gap-2 px-3 py-2 text-sm rounded-lg transition-all relative z-10
                ${
                  isActive
                    ? 'font-thin text-blue-500 bg-blue-500/10'
                    : 'hover:font-semibold'
                }
              `}
              style={{
                paddingLeft: `${depth * 16 + 12}px`,
                color: isActive ? undefined : textColor
              }}
            >
              <FileText
                className="w-4 h-4 shrink-0 mt-0.5"
                style={{ color: isActive ? '#3b82f6' : textColor }}
              />
              <span className="flex-1 wrap-break-word">{displayTitle}</span>
            </Link>
          )}
        </div>

        {/* Render children */}
        {hasChildren && (
          <div 
            className={`
              relative space-y-1 overflow-hidden transition-all duration-300 ease-in-out
              ${isCollapsed ? 'max-h-0 opacity-0' : 'max-h-500 opacity-100'}
            `}
          >
            {/* Vertical line connecting children */}
            <div 
              className="absolute left-0 top-0 bottom-0 w-px bg-gray-200 dark:bg-gray-700"
              style={{ left: `${depth * 16 + 20}px` }}
            />
            {item.children!.map(child => renderMenuItem(child, depth + 1, itemKey))}
          </div>
        )}
      </div>
    );
  };

  // Render full sidebar (expanded state)
  const renderFullSidebar = () => (
    <>
      {/* Scrollable navigation area */}
      <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden">
        <nav className="p-4 pb-6 w-full space-y-2">
          {pageMap.map(item => renderMenuItem(item))}
        </nav>
      </div>

      {/* Related Projects - fixed at bottom, shrink-0 prevents shrinking */}
      <div className="shrink-0">
        <RelatedProjects onCollapse={toggleSidebar} isMobile={menuOpen} bannerActive={!bannerDismissed} />
      </div>
    </>
  );

  // Render slim sidebar (collapsed state) - Desktop only
  const renderSlimSidebar = () => (
    <div className="flex flex-col h-full">
      {/* Spacer */}
      <div className="flex-1"></div>

      {/* Footer with icon buttons */}
      <RelatedProjects onCollapse={toggleSidebar} variant="slim" />
    </div>
  );

  return (
    <aside
      ref={sidebarRef}
      className={`
        fixed lg:sticky left-0
        shadow-sm dark:shadow-none
        flex flex-col
        overflow-hidden
        transition-all duration-300 ease-in-out
        bg-white dark:bg-black
        border-r border-gray-200 dark:border-gray-800
        ${menuOpen ? 'translate-x-0 w-60 z-30' : '-translate-x-full w-0 lg:translate-x-0 z-20'}
        ${sidebarCollapsed ? 'lg:w-16' : 'lg:w-60'}
        ${className || ''}
      `}
      style={{
        top: layoutValues.top,
        height: layoutValues.height,
        maxHeight: layoutValues.height,
        boxShadow: '0 1px 6px 0 rgba(0,0,0,0.07)',
        backgroundColor: 'var(--background)',
      }}
      suppressHydrationWarning
    >
      {/* Full Sidebar Content */}
      <div className={`
        flex flex-col h-full
        transition-all duration-300 ease-in-out
        ${sidebarCollapsed ? 'opacity-0 pointer-events-none' : 'opacity-100'}
      `}>
        {renderFullSidebar()}
      </div>

      {/* Slim Sidebar Content - Desktop only */}
      <div className={`
        hidden lg:flex flex-col h-full absolute inset-0
        transition-all duration-300 ease-in-out
        ${sidebarCollapsed ? 'opacity-100' : 'opacity-0 pointer-events-none'}
      `}>
        {renderSlimSidebar()}
      </div>
    </aside>
  );
}
