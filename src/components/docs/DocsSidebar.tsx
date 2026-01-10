"use client";

import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ChevronRight, ChevronDown, FileText } from 'lucide-react';
import { SidebarFooter } from './SidebarFooter';
import { useDocsMenu } from './DocsProvider';
import { useTheme } from 'next-themes';

interface MenuItem {
  name: string;
  route?: string;
  title?: string;
  children?: MenuItem[];
  frontMatter?: Record<string, unknown>;
  kind?: string;
}

interface DocsSidebarProps {
  pageMap: MenuItem[];
  className?: string;
}

export function DocsSidebar({ pageMap, className }: DocsSidebarProps) {
  const pathname = usePathname();
  const navRef = useRef<HTMLElement>(null);
  const sidebarRef = useRef<HTMLElement>(null);
  const { sidebarCollapsed, toggleSidebar, menuOpen, bannerDismissed } = useDocsMenu();
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set());
  const [availableHeight, setAvailableHeight] = useState<string>('auto');
  const [topOffset, setTopOffset] = useState('4rem');
  const [sidebarHeight, setSidebarHeight] = useState('calc(100vh - 4rem)');

  useEffect(() => {
    setMounted(true);
  }, []);

  // Calculate dynamic top offset and height based on banner presence
  useEffect(() => {
    const calculateOffsets = () => {
      // Get the navbar element
      const navbar = document.querySelector('.nextra-nav-container');
      
      if (navbar) {
        // Get the actual bottom position of the navbar (which includes banner if present)
        const navbarRect = navbar.getBoundingClientRect();
        const navbarBottom = navbarRect.bottom;
        
        // The sidebar should start where the navbar ends
        setTopOffset(`${navbarBottom}px`);
        setSidebarHeight(`calc(100vh - ${navbarBottom}px)`);
      } else {
        // Fallback if navbar not found
        setTopOffset('4rem');
        setSidebarHeight('calc(100vh - 4rem)');
      }
    };

    calculateOffsets();
    
    // Recalculate on window resize, scroll (for sticky navbar), and banner changes
    window.addEventListener('resize', calculateOffsets);
    window.addEventListener('scroll', calculateOffsets);
    
    // Also recalculate after a short delay to ensure DOM is ready
    const timer = setTimeout(calculateOffsets, 100);
    
    // Recalculate periodically to catch any changes
    const interval = setInterval(calculateOffsets, 500);
    
    return () => {
      window.removeEventListener('resize', calculateOffsets);
      window.removeEventListener('scroll', calculateOffsets);
      clearTimeout(timer);
      clearInterval(interval);
    };
  }, [bannerDismissed]);

  // Calculate available height for navigation, accounting for footer and viewport changes
  useEffect(() => {
    const calculateHeight = () => {
      if (navRef.current) {
        const parent = navRef.current.parentElement;
        if (parent) {
          // Get the parent's actual height (which is constrained by viewport)
          const parentHeight = parent.clientHeight;
          
          // Find footer element and get its height
          const footer = parent.nextElementSibling;
          const footerHeight = footer ? footer.clientHeight : 0;
          
          // Calculate available height for navigation
          const navHeight = parentHeight - footerHeight;
          setAvailableHeight(`${navHeight}px`);
        }
      }
    };

    calculateHeight();
    
    // Recalculate on window resize
    window.addEventListener('resize', calculateHeight);
    
    // Recalculate on scroll (for when banner appears/disappears)
    window.addEventListener('scroll', calculateHeight);
    
    // Use MutationObserver to recalculate if DOM changes
    const observer = new MutationObserver(calculateHeight);
    if (navRef.current?.parentElement?.parentElement) {
      observer.observe(navRef.current.parentElement.parentElement, {
        childList: true,
        subtree: true,
        attributes: true,
        attributeFilter: ['class', 'style']
      });
    }
    
    return () => {
      window.removeEventListener('resize', calculateHeight);
      window.removeEventListener('scroll', calculateHeight);
      observer.disconnect();
    };
  }, []);

  // Auto-expand only the folders that contain the active page
  useEffect(() => {
    const newCollapsed = new Set<string>();
    const pathToActive = new Set<string>();
    
    // Find the path to the active item
    function findActivePath(items: MenuItem[], parentKey: string = ''): boolean {
      for (const item of items) {
        const itemKey = parentKey ? `${parentKey}-${item.name}` : item.name;
        const isActive = item.route && pathname === item.route;
        
        if (isActive) {
          // Found the active item, don't collapse any parent in the path
          return true;
        }
        
        if (item.children) {
          const childActive = findActivePath(item.children, itemKey);
          if (childActive) {
            // This folder contains the active item, keep it expanded
            pathToActive.add(itemKey);
            return true;
          }
        }
      }
      return false;
    }
    
    // Collapse all folders except those in the active path
    function collapseAll(items: MenuItem[], parentKey: string = '') {
      for (const item of items) {
        const itemKey = parentKey ? `${parentKey}-${item.name}` : item.name;
        const hasChildren = item.children && item.children.length > 0;
        
        if (hasChildren) {
          // Collapse this folder if it's not in the path to active item
          if (!pathToActive.has(itemKey)) {
            newCollapsed.add(itemKey);
          }
          // Recursively check children
          if (item.children) {
            collapseAll(item.children, itemKey);
          }
        }
      }
    }
    
    findActivePath(pageMap);
    collapseAll(pageMap);
    setCollapsed(newCollapsed);
  }, [pathname, pageMap]);

  const toggleCollapse = (itemKey: string) => {
    setCollapsed(prev => {
      const next = new Set(prev);
      if (next.has(itemKey)) {
        next.delete(itemKey);
      } else {
        next.add(itemKey);
      }
      return next;
    });
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
              style={{ 
                paddingLeft: `${depth * 16 + 12}px`,
                color: 'var(--foreground)'
              }}
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
                    ? 'font-thin'
                    : 'hover:font-semibold'
                }
              `}
              style={{ 
                paddingLeft: `${depth * 16 + 12}px`,
                color: isActive ? undefined : 'var(--foreground)',
                backgroundColor: isActive ? 'rgba(59, 130, 246, 0.1)' : undefined
              }}
            >
              <FileText className="w-4 h-4 shrink-0 mt-0.5" style={{ color: isActive ? '#3b82f6' : undefined }} />
              <span className="flex-1 wrap-break-word" style={{ color: isActive ? '#3b82f6' : undefined }}>{displayTitle}</span>
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
        <nav 
          ref={navRef}
          className="p-4 pb-6 w-full space-y-2" 
          style={{ maxHeight: availableHeight }}
        >
          {pageMap.map(item => renderMenuItem(item))}
        </nav>
      </div>
      
      {/* Footer at bottom */}
      <SidebarFooter onCollapse={toggleSidebar} isMobile={menuOpen} />
    </>
  );

  // Render slim sidebar (collapsed state) - Desktop only
  const renderSlimSidebar = () => (
    <div className="flex flex-col h-full">
      {/* Spacer */}
      <div className="flex-1"></div>
      
      {/* Footer with icon buttons */}
      <SidebarFooter onCollapse={toggleSidebar} variant="slim" />
    </div>
  );

  const isDark = mounted && resolvedTheme === 'dark';

  return (
    <aside
      ref={sidebarRef}
      className={`
        fixed lg:sticky left-0
        shadow-sm dark:shadow-none
        flex flex-col
        overflow-hidden
        transition-all duration-300 ease-in-out
        ${menuOpen ? 'translate-x-0 w-60 z-30' : '-translate-x-full w-0 lg:translate-x-0 z-20'}
        ${sidebarCollapsed ? 'lg:w-16' : 'lg:w-60'}
        ${className || ''}
      `}
      style={{
        top: topOffset,
        height: sidebarHeight,
        maxHeight: sidebarHeight,
        boxShadow: '0 1px 6px 0 rgba(0,0,0,0.07)',
        backgroundColor: isDark ? '#000000' : '#ffffff',
        borderRight: isDark ? '1px solid #1f2937' : '1px solid #e5e7eb',
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
