"use client";

import { useState, useEffect } from 'react';
import { useTheme } from 'next-themes';
import { useDocsMenu } from './DocsProvider';

interface MobileHeaderProps {
  onToggleSidebar: () => void;
  pageTitle?: string;
}

export function MobileHeader({ onToggleSidebar, pageTitle }: MobileHeaderProps) {
  const { resolvedTheme } = useTheme();
  const { dismissBanner } = useDocsMenu();
  const [mounted, setMounted] = useState(false);
  const [isHovered, setIsHovered] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const handleToggle = () => {
    // Dismiss the banner when opening the sidebar on mobile
    dismissBanner();
    onToggleSidebar();
  };

  // Prevent hydration mismatch by not applying theme-specific styles until mounted
  const isDark = mounted ? resolvedTheme === 'dark' : false;

  return (
    <div className="lg:hidden">
      <button
        onClick={handleToggle}
        className="flex items-center px-4 sm:px-6 md:px-8 py-3 focus:outline-none transition-colors w-full"
        aria-label="Open sidebar"
        style={{
          color: isHovered
            ? (isDark ? '#f3f4f6' : '#111827')
            : (isDark ? '#9ca3af' : '#6b7280'),
        }}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
        suppressHydrationWarning
      >
        <span className="text-sm font-medium flex-1 text-left">
          {pageTitle || 'Menu'}
        </span>
        <svg
          className="w-5 h-5 rotate-90 ml-2"
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path
            fillRule="evenodd"
            d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
            clipRule="evenodd"
          />
        </svg>
      </button>
    </div>
  );
}
