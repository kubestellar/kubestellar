"use client";

import { useEffect, useState } from 'react';
import { useTheme } from 'next-themes';
import Link from 'next/link';

interface TOCItem {
  id: string;
  value: string;
  depth: number;
}

interface TableOfContentsProps {
  toc?: TOCItem[];
}

function TOCLink({ item, isActive, isDark }: { item: TOCItem; isActive: boolean; isDark: boolean }) {
  const [isHovered, setIsHovered] = useState(false);
  const indent = (item.depth - 2) * 12;

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    const element = document.getElementById(item.id);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'start' });
      // Update URL without jumping
      window.history.pushState(null, '', `#${item.id}`);
    }
  };

  return (
    <Link
      href={`#${item.id}`}
      className="block py-1.5 text-sm transition-colors border-l-2"
      style={{
        paddingLeft: `${indent + 12}px`,
        borderColor: isActive
          ? (isDark ? '#60a5fa' : '#2563eb')
          : (isHovered ? (isDark ? '#374151' : '#d1d5db') : 'transparent'),
        color: isActive
          ? (isDark ? '#60a5fa' : '#2563eb')
          : (isHovered ? (isDark ? '#f3f4f6' : '#111827') : (isDark ? '#9ca3af' : '#374151')),
        fontWeight: isActive ? 500 : 400,
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

export function TableOfContents({ toc }: TableOfContentsProps) {
  const [activeId, setActiveId] = useState<string>('');
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!toc || toc.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
          }
        });
      },
      {
        rootMargin: '0px 0px -80% 0px',
        threshold: 1.0,
      }
    );

    // Observe all heading elements
    toc.forEach(({ id }) => {
      const element = document.getElementById(id);
      if (element) {
        observer.observe(element);
      }
    });

    return () => {
      observer.disconnect();
    };
  }, [toc]);

  if (!toc || toc.length === 0) {
    return null;
  }

  const isDark = mounted && resolvedTheme === 'dark';

  return (
    <aside 
      className="hidden xl:block w-64 overflow-y-auto"
      style={{
        position: 'sticky',
        top: 'calc(var(--nextra-navbar-height, 4rem) + var(--nextra-banner-height, 0px))',
        height: 'calc(100vh - var(--nextra-navbar-height, 4rem) - var(--nextra-banner-height, 0px))',
        borderLeft: isDark ? '1px solid #1f2937' : '1px solid #e5e7eb',
      }}
      suppressHydrationWarning
    >
      <div className="p-4">
        <h3 
          className="text-sm font-semibold mb-4"
          style={{
            color: isDark ? '#f3f4f6' : '#111827',
          }}
          suppressHydrationWarning
        >
          On This Page
        </h3>
        <nav className="space-y-2">
          {toc.map((item) => (
            <TOCLink
              key={item.id}
              item={item}
              isActive={activeId === item.id}
              isDark={isDark}
            />
          ))}
        </nav>

        {/* Back to top link */}
        <div 
          className="mt-8 pt-4"
          style={{
            borderTop: isDark ? '1px solid #1f2937' : '1px solid #e5e7eb',
          }}
          suppressHydrationWarning
        >
          <Link
            href="#"
            className="text-xs hover:underline"
            style={{
              color: isDark ? '#60a5fa' : '#2563eb',
            }}
            onClick={(e) => {
              e.preventDefault();
              window.scrollTo({ top: 0, behavior: 'smooth' });
            }}
            suppressHydrationWarning
          >
            ↑ Back to top
          </Link>
        </div>
      </div>
    </aside>
  );
}
