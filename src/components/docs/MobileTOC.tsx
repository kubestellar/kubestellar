"use client";

import { useState, useEffect } from 'react';
import { useTheme } from 'next-themes';
import Link from 'next/link';

interface TOCItem {
  id: string;
  value: string;
  depth: number;
}

interface MobileTOCProps {
  toc?: TOCItem[];
}

function TOCLink({ item, isDark, onClose }: { item: TOCItem; isDark: boolean; onClose: () => void }) {
  const [isHovered, setIsHovered] = useState(false);
  const indent = (item.depth - 2) * 16;

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    const element = document.getElementById(item.id);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'start' });
      // Update URL without jumping
      window.history.pushState(null, '', `#${item.id}`);
    }
    onClose();
  };

  return (
    <Link
      href={`#${item.id}`}
      className="block py-2 text-sm border-l-2 transition-colors"
      style={{
        paddingLeft: `${indent + 12}px`,
        borderColor: isHovered 
          ? (isDark ? '#374151' : '#2563eb')
          : 'transparent',
        color: isHovered
          ? (isDark ? '#f3f4f6' : '#111827')
          : '#6b7280',
      }}
      onClick={handleClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      suppressHydrationWarning
    >
      {item.value}
    </Link>
  );
}

export function MobileTOC({ toc }: MobileTOCProps) {
  const [isOpen, setIsOpen] = useState(false);
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const [headerHovered, setHeaderHovered] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!toc || toc.length === 0) {
    return null;
  }

  // Prevent hydration mismatch by using default light theme on server
  const isDark = mounted && resolvedTheme === 'dark';

  return (
    <div 
      className="xl:hidden mb-6 rounded-lg overflow-hidden border sticky top-16 z-10"
      style={{
        backgroundColor: isDark ? '#000000' : '#ffffff',
        borderColor: isDark ? '#1f2937' : '#e5e7eb',
      }}
      suppressHydrationWarning
    >
      {/* Accordion Header */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-4 py-3 flex items-center justify-between text-left transition-colors"
        style={{
          backgroundColor: headerHovered 
            ? (isDark ? '#111827' : '#f3f4f6')
            : (isDark ? '#000000' : '#ffffff'),
          color: headerHovered
            ? (isDark ? '#f3f4f6' : '#111827')
            : (isDark ? '#9ca3af' : '#6b7280'),
        }}
        suppressHydrationWarning
        onMouseEnter={() => setHeaderHovered(true)}
        onMouseLeave={() => setHeaderHovered(false)}
      >
        <div className="flex items-center gap-2">
          <svg 
            width="20" 
            height="20" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2"
            strokeLinecap="round" 
            strokeLinejoin="round"
          >
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
          <span className="font-semibold text-sm">On This Page</span>
        </div>
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          style={{
            transform: isOpen ? 'rotate(180deg)' : 'rotate(0deg)',
            transition: 'transform 0.2s ease',
          }}
        >
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>

      {/* Accordion Content */}
      <div
        style={{
          maxHeight: isOpen ? '400px' : '0',
          overflow: isOpen ? 'auto' : 'hidden',
          transition: 'max-height 0.3s ease-in-out',
        }}
        suppressHydrationWarning
      >
        <nav 
          className="px-4 py-3 space-y-1"
          style={{
            backgroundColor: isDark ? '#000000' : '#ffffff',
          }}
          suppressHydrationWarning
        >
          {toc.map((item) => (
            <TOCLink
              key={item.id}
              item={item}
              isDark={isDark}
              onClose={() => setIsOpen(false)}
            />
          ))}
        </nav>
      </div>
    </div>
  );
}
