"use client";

import React, { useState, useEffect, useRef } from "react";
import Link from "next/link";
import Image from "next/image";
import { useTheme } from "next-themes";

type DropdownType = "contribute" | "community" | "version" | "language" | "github" | null;

export default function DocsNavbar() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [openDropdown, setOpenDropdown] = useState<DropdownType>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [searchResults, setSearchResults] = useState<Array<{
    title: string;
    url: string;
    category: string;
    snippet: string;
    highlightedSnippet: string;
    matchType: string;
  }>>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const commandPaletteRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<NodeJS.Timeout | null>(null);
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const [githubStats, setGithubStats] = useState({
    stars: "0",
    forks: "0",
    watchers: "0",
  });

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    const fetchGithubStats = async () => {
      try {
        const response = await fetch(
          "https://api.github.com/repos/kubestellar/kubestellar"
        );
        if (!response.ok) {
          throw new Error("Network response was not okay");
        }
        const data = await response.json();
        const formatNumber = (num: number): string => {
          if (num >= 1000) {
            return (num / 1000).toFixed(1) + "K";
          }
          return num.toString();
        };
        setGithubStats({
          stars: formatNumber(data.stargazers_count),
          forks: formatNumber(data.forks_count),
          watchers: formatNumber(data.subscribers_count),
        });
      } catch (err) {
        console.error("Failed to fetch Github stats: ", err);
      }
    };
    fetchGithubStats();

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (isSearchOpen) {
          setIsSearchOpen(false);
          setSearchQuery("");
          setSearchResults([]);
          setSelectedIndex(0);
        } else {
          setOpenDropdown(null);
        }
      }
      if ((e.ctrlKey || e.metaKey) && e.key === "k") {
        e.preventDefault();
        setIsSearchOpen(!isSearchOpen);
        setTimeout(() => searchInputRef.current?.focus(), 100);
      }
      
      // Navigation in search results
      if (isSearchOpen && searchResults.length > 0) {
        if (e.key === "ArrowDown") {
          e.preventDefault();
          setSelectedIndex((prev) => (prev < searchResults.length - 1 ? prev + 1 : prev));
        } else if (e.key === "ArrowUp") {
          e.preventDefault();
          setSelectedIndex((prev) => (prev > 0 ? prev - 1 : 0));
        } else if (e.key === "Enter" && searchResults[selectedIndex]) {
          e.preventDefault();
          window.location.href = searchResults[selectedIndex].url;
        }
      }
    };
    
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isSearchOpen, searchResults, selectedIndex]);

  const handleMouseEnter = (dropdown: DropdownType) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    setOpenDropdown(dropdown);
  };

  const handleMouseLeave = () => {
    timeoutRef.current = setTimeout(() => {
      setOpenDropdown(null);
    }, 150);
  };

  const handleDropdownMouseEnter = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  };

  const isDark = resolvedTheme === 'dark';
  const [isSearching, setIsSearching] = useState(false);

  const performSearchAPI = async (query: string) => {
    try {
      // Call the search API
      const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
      
      if (!response.ok) {
        throw new Error('Search failed');
      }
      
      const data = await response.json();
      setSearchResults(data.results || []);
    } catch (error) {
      console.error('Search error:', error);
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  };

  const performSearch = (query: string) => {
    setSearchQuery(query);
    setSelectedIndex(0);
    
    // Clear existing debounce timer
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }
    
    if (!query.trim()) {
      setSearchResults([]);
      setIsSearching(false);
      return;
    }

    setIsSearching(true);

    // Debounce search API calls (300ms delay)
    debounceRef.current = setTimeout(() => {
      performSearchAPI(query);
    }, 300);
  };

  if (!mounted) {
    return null;
  }
  
  const buttonClasses = `text-sm transition-colors px-2 py-1.5 rounded-md flex items-center gap-1.5 ${
    isDark 
      ? 'text-gray-400 hover:text-gray-100 hover:bg-neutral-800'
      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
  }`;
  
  const dropdownClasses = `absolute left-0 mt-1 w-52 rounded-md shadow-xl py-1 border z-50 ${
    isDark 
      ? 'bg-neutral-900 border-neutral-800'
      : 'bg-white border-gray-200'
  }`;
  
  const dropdownItemClasses = `flex items-center px-3 py-2 text-sm transition-colors ${
    isDark
      ? 'text-gray-300 hover:bg-neutral-800'
      : 'text-gray-700 hover:bg-gray-100'
  }`;

  return (
    <div className="nextra-nav-container sticky top-0 z-20 w-full bg-transparent">
      <div className={`nextra-nav-container-blur pointer-events-none absolute z-[-1] h-full w-full shadow-sm border-b ${
        isDark 
          ? 'bg-[#111] border-neutral-800' 
          : 'bg-white border-gray-200'
      }`} />
      
      <div className="mx-auto flex items-center gap-2 h-16 px-4 max-w-[90rem]">
        <Link href="/" className="flex items-center hover:opacity-75 transition-opacity">
          <Image
            src="/KubeStellar-with-Logo-transparent.png"
            alt="KubeStellar"
            width={160}
            height={40}
            className="h-8 w-auto"
            priority
          />
        </Link>

        <div className="flex-1" />

        <div className="hidden md:flex items-center gap-1">
          
          <div 
            className="relative" 
            onMouseEnter={() => handleMouseEnter("contribute")}
            onMouseLeave={handleMouseLeave}
          >
            <button
              type="button"
              className={buttonClasses}
              aria-haspopup="true"
              aria-expanded={openDropdown === "contribute"}
            >
              <svg
                className="w-3.5 h-3.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                />
              </svg>
              <span>Contribute</span>
              <svg className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            {openDropdown === "contribute" && (
              <div
                className={dropdownClasses}
                onMouseEnter={handleDropdownMouseEnter}
                onMouseLeave={handleMouseLeave}
              >
              <Link
                href="/#join-in"
                className={dropdownItemClasses}
              >
                <svg className="w-4 h-4 mr-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
                </svg>
                Join In
              </Link>
              <Link
                href="/contribute-handbook"
                className={dropdownItemClasses}
              >
                <svg className="w-4 h-4 mr-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                </svg>
                Contribute Handbook
              </Link>
              <Link
                href="/#security"
                className={dropdownItemClasses}
              >
                <svg className="w-4 h-4 mr-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
                Security
              </Link>
              </div>
            )}
          </div>

          <div 
            className="relative"
            onMouseEnter={() => handleMouseEnter("community")}
            onMouseLeave={handleMouseLeave}
          >
            <button
              type="button"
              className={buttonClasses}
              aria-haspopup="true"
              aria-expanded={openDropdown === "community"}
            >
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
              <span>Community</span>
              <svg className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            {openDropdown === "community" && (
              <div
                className={dropdownClasses}
                onMouseEnter={handleDropdownMouseEnter}
                onMouseLeave={handleMouseLeave}
              >
              <Link
                href="/#get-involved"
                className={dropdownItemClasses}
              >
                <svg className="w-4 h-4 mr-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
                </svg>
                Get Involved
              </Link>
              <Link
                href="/programs"
                className={dropdownItemClasses}
              >
                <svg className="w-4 h-4 mr-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2H5a2 2 0 00-2 2v2M7 7h10" />
                </svg>
                Programs
              </Link>
              <Link
                href="/#ladder"
                className={dropdownItemClasses}
              >
                <svg className="w-4 h-4 mr-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                </svg>
                Ladder
              </Link>
              <Link
                href="/#contact-us"
                className={dropdownItemClasses}
              >
                <svg className="w-4 h-4 mr-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 8l7.89 4.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                Contact Us
              </Link>
              <Link
                href="/#partners"
                className={dropdownItemClasses}
              >
                <svg className="w-4 h-4 mr-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
                Partners
              </Link>
              </div>
            )}
          </div>

          <div className="w-px h-5 bg-gray-300 dark:bg-neutral-700 mx-1" />

          <div 
            className="relative"
            onMouseEnter={() => handleMouseEnter("version")}
            onMouseLeave={handleMouseLeave}
          >
            <button
              className={`text-xs font-mono transition-colors px-2 py-1.5 rounded-md flex items-center gap-1.5 ${
                isDark 
                  ? 'text-gray-400 hover:text-gray-100 hover:bg-neutral-800'
                  : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
              }`}
              aria-haspopup="true"
              aria-expanded={openDropdown === "version"}
            >
              <span>v3.8.1</span>
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            {openDropdown === "version" && (
              <div
                className={`absolute right-0 mt-1 w-44 rounded-md shadow-xl py-1 border z-50 ${
                  isDark 
                    ? 'bg-neutral-900 border-neutral-800'
                    : 'bg-white border-gray-200'
                }`}
                onMouseEnter={handleDropdownMouseEnter}
                onMouseLeave={handleMouseLeave}
              >
              <a href="#" className={`block px-3 py-2 text-sm transition-colors ${
                isDark
                  ? 'text-gray-300 hover:bg-neutral-800'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}>
                v3.8.1 <span className={isDark ? 'text-gray-400' : 'text-gray-500'}>(Current)</span>
              </a>
              <a href="#" className={`block px-3 py-2 text-sm transition-colors ${
                isDark
                  ? 'text-gray-300 hover:bg-neutral-800'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}>
                v3.8.0
              </a>
              <hr className={isDark ? 'my-1 border-neutral-800' : 'my-1 border-gray-200'} />
              <a href="#" className={`block px-3 py-2 text-sm transition-colors ${
                isDark
                  ? 'text-gray-300 hover:bg-neutral-800'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}>
                All versions →
              </a>
              </div>
            )}
          </div>

          <div 
            className="relative"
            onMouseEnter={() => handleMouseEnter("language")}
            onMouseLeave={handleMouseLeave}
          >
            <button
              className={`text-sm transition-colors p-1.5 rounded-md flex items-center ${
                isDark 
                  ? 'text-gray-400 hover:text-gray-100 hover:bg-neutral-800'
                  : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
              }`}
              aria-label="Change language"
              aria-haspopup="true"
              aria-expanded={openDropdown === "language"}
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
              </svg>
            </button>
            {openDropdown === "language" && (
              <div
                className={`absolute right-0 mt-1 w-32 rounded-md shadow-xl py-1 border z-50 ${
                  isDark 
                    ? 'bg-neutral-900 border-neutral-800'
                    : 'bg-white border-gray-200'
                }`}
                onMouseEnter={handleDropdownMouseEnter}
                onMouseLeave={handleMouseLeave}
              >
              <a href="#" className={`block px-3 py-2 text-sm transition-colors ${
                isDark
                  ? 'text-gray-300 hover:bg-neutral-800'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}>
                English
              </a>
              <a href="#" className={`block px-3 py-2 text-sm transition-colors ${
                isDark
                  ? 'text-gray-300 hover:bg-neutral-800'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}>
                日本語
              </a>
              <a href="#" className={`block px-3 py-2 text-sm transition-colors ${
                isDark
                  ? 'text-gray-300 hover:bg-neutral-800'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}>
                简体中文
              </a>
              </div>
            )}
          </div>

          <div 
            className="relative"
            onMouseEnter={() => handleMouseEnter("github")}
            onMouseLeave={handleMouseLeave}
          >
            <button
              className={`text-sm transition-colors p-1.5 rounded-md flex items-center ${
                isDark 
                  ? 'text-gray-400 hover:text-gray-100 hover:bg-neutral-800'
                  : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
              }`}
              aria-label="GitHub"
              aria-haspopup="true"
              aria-expanded={openDropdown === "github"}
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0C5.374 0 0 5.373 0 12 0 17.302 3.438 21.8 8.207 23.387c.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.300 24 12c0-6.627-5.373-12-12-12z" />
              </svg>
            </button>
            {openDropdown === "github" && (
              <div
                className={`absolute right-0 mt-1 w-44 rounded-md shadow-xl py-1 border z-50 ${
                  isDark 
                    ? 'bg-neutral-900 border-neutral-800'
                    : 'bg-white border-gray-200'
                }`}
                onMouseEnter={handleDropdownMouseEnter}
                onMouseLeave={handleMouseLeave}
              >
              <a
                href="https://github.com/kubestellar/kubestellar"
                target="_blank"
                rel="noopener noreferrer"
                className={`flex items-center justify-between px-3 py-2 text-sm transition-colors ${
                  isDark
                    ? 'text-gray-300 hover:bg-neutral-800'
                    : 'text-gray-700 hover:bg-gray-100'
                }`}
              >
                <span className="flex items-center gap-2">
                  <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 16 16">
                    <path d="M8 .25a.75.75 0 01.673.418l1.882 3.815 4.21.612a.75.75 0 01.416 1.279l-3.046 2.97.719 4.192a.75.75 0 01-1.088.791L8 12.347l-3.766 1.98a.75.75 0 01-1.088-.79l.72-4.194L.818 6.374a.75.75 0 01.416-1.28l4.21-.611L7.327.668A.75.75 0 018 .25z"/>
                  </svg>
                  Star
                </span>
                <span className={`text-xs px-1.5 py-0.5 rounded ${
                  isDark ? 'bg-neutral-800' : 'bg-gray-200'
                }`}>
                  {githubStats.stars}
                </span>
              </a>
              <a
                href="https://github.com/kubestellar/kubestellar/fork"
                target="_blank"
                rel="noopener noreferrer"
                className={`flex items-center justify-between px-3 py-2 text-sm transition-colors ${
                  isDark
                    ? 'text-gray-300 hover:bg-neutral-800'
                    : 'text-gray-700 hover:bg-gray-100'
                }`}
              >
                <span className="flex items-center gap-2">
                  <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 16 16">
                    <path d="M5 3.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm0 2.122a2.25 2.25 0 10-1.5 0v.878A2.25 2.25 0 005.75 8.5h1.5v2.128a2.251 2.251 0 101.5 0V8.5h1.5a2.25 2.25 0 002.25-2.25v-.878a2.25 2.25 0 10-1.5 0v.878a.75.75 0 01-.75.75h-4.5A.75.75 0 015 6.25v-.878zm3.75 7.378a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm3-8.75a.75.75 0 100-1.5.75.75 0 000 1.5z"/>
                  </svg>
                  Fork
                </span>
                <span className={`text-xs px-1.5 py-0.5 rounded ${
                  isDark ? 'bg-neutral-800' : 'bg-gray-200'
                }`}>
                  {githubStats.forks}
                </span>
              </a>
              </div>
            )}
          </div>
        </div>

        <button
          onClick={() => {
            setIsSearchOpen(true);
            setTimeout(() => searchInputRef.current?.focus(), 100);
          }}
          className={`hidden md:flex text-sm transition-colors px-3 py-1.5 rounded-md items-center gap-2 ml-2 ${
            isDark 
              ? 'text-gray-400 hover:text-gray-100 hover:bg-neutral-800 border border-neutral-800'
              : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100 border border-gray-200'
          }`}
          aria-label="Search documentation"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <span className="text-xs">Search docs...</span>
          <kbd className={`text-xs px-1.5 py-0.5 rounded ${
            isDark ? 'bg-neutral-800 text-gray-400' : 'bg-gray-100 text-gray-500'
          }`}>
            ⌘K
          </kbd>
        </button>

        <button
          onClick={() => {
            setIsSearchOpen(true);
            setTimeout(() => searchInputRef.current?.focus(), 100);
          }}
          className={`md:hidden p-1.5 rounded-md transition-colors ${
            isDark 
              ? 'text-gray-400 hover:text-gray-100 hover:bg-neutral-800'
              : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
          }`}
          aria-label="Search documentation"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </button>

        <button
          className={`md:hidden p-1.5 rounded-md transition-colors ${
            isDark 
              ? 'text-gray-400 hover:text-gray-100 hover:bg-neutral-800'
              : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
          }`}
          aria-label="Toggle menu"
          onClick={() => setIsMenuOpen(!isMenuOpen)}
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={isMenuOpen ? "M6 18L18 6M6 6l12 12" : "M4 6h16M4 12h16M4 18h16"} />
          </svg>
        </button>
      </div>

      {/* Command Palette Modal */}
      {isSearchOpen && (
        <>
          {/* Backdrop */}
          <div 
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50"
            onClick={() => {
              setIsSearchOpen(false);
              setSearchQuery("");
              setSearchResults([]);
            }}
          />
          
          {/* Command Palette */}
          <div className="fixed top-20 left-1/2 -translate-x-1/2 w-full max-w-2xl z-50 px-4">
            <div 
              ref={commandPaletteRef}
              className={`rounded-lg shadow-2xl border ${
                isDark 
                  ? 'bg-neutral-900 border-neutral-700' 
                  : 'bg-white border-gray-200'
              }`}
            >
              {/* Search Input */}
              <div className="flex items-center gap-3 px-4 py-3 border-b border-neutral-700">
                <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <input
                  ref={searchInputRef}
                  type="text"
                  value={searchQuery}
                  onChange={(e) => performSearch(e.target.value)}
                  placeholder="Search documentation..."
                  className={`flex-1 bg-transparent outline-none text-base ${
                    isDark ? 'text-gray-100 placeholder-gray-500' : 'text-gray-900 placeholder-gray-400'
                  }`}
                  autoFocus
                />
                <kbd className={`text-xs px-2 py-1 rounded ${
                  isDark ? 'bg-neutral-800 text-gray-400' : 'bg-gray-100 text-gray-500'
                }`}>
                  ESC
                </kbd>
              </div>

              {/* Search Results */}
              <div className="max-h-96 overflow-y-auto">
                {searchQuery.trim() === "" ? (
                  <div className="px-4 py-8 text-center">
                    <svg className="w-12 h-12 mx-auto mb-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                    <p className={`text-sm ${isDark ? 'text-gray-400' : 'text-gray-600'}`}>
                      Search for any word or phrase in the documentation...
                    </p>
                    <p className={`text-xs mt-2 ${isDark ? 'text-gray-500' : 'text-gray-500'}`}>
                      Try &quot;kubectl&quot;, &quot;cluster&quot;, &quot;workload&quot;, or &quot;installation&quot;
                    </p>
                  </div>
                ) : isSearching ? (
                  <div className="px-4 py-8 text-center">
                    <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-gray-400 mb-3"></div>
                    <p className={`text-sm ${isDark ? 'text-gray-400' : 'text-gray-600'}`}>
                      Searching documentation...
                    </p>
                  </div>
                ) : searchResults.length === 0 ? (
                  <div className="px-4 py-8 text-center">
                    <svg className="w-12 h-12 mx-auto mb-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M12 12h.01M12 12h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <p className={`text-sm ${isDark ? 'text-gray-400' : 'text-gray-600'}`}>
                      No results found for &quot;{searchQuery}&quot;
                    </p>
                    <p className={`text-xs mt-2 ${isDark ? 'text-gray-500' : 'text-gray-500'}`}>
                      Try different keywords or check spelling
                    </p>
                  </div>
                ) : (
                  <div className="py-2">
                    {searchResults.map((result, index) => (
                      <a
                        key={index}
                        href={result.url}
                        className={`block px-4 py-3 transition-colors border-l-2 ${
                          index === selectedIndex
                            ? isDark 
                              ? 'bg-neutral-800 border-blue-500' 
                              : 'bg-gray-100 border-blue-600'
                            : isDark
                              ? 'hover:bg-neutral-800 border-transparent'
                              : 'hover:bg-gray-50 border-transparent'
                        }`}
                        onMouseEnter={() => setSelectedIndex(index)}
                      >
                        <div className="flex items-start gap-3">
                          <div className={`mt-0.5 p-1.5 rounded flex-shrink-0 ${
                            isDark ? 'bg-neutral-700' : 'bg-gray-200'
                          }`}>
                            <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                              <div className={`font-medium text-sm ${
                                isDark ? 'text-gray-100' : 'text-gray-900'
                              }`}>
                                {result.title}
                              </div>
                              <span className={`text-xs px-1.5 py-0.5 rounded flex-shrink-0 ${
                                isDark ? 'bg-neutral-700 text-gray-400' : 'bg-gray-200 text-gray-600'
                              }`}>
                                {result.category}
                              </span>
                            </div>
                            <div 
                              className={`text-xs leading-relaxed ${
                                isDark ? 'text-gray-400' : 'text-gray-600'
                              }`}
                              dangerouslySetInnerHTML={{ 
                                __html: result.highlightedSnippet.replace(
                                  /<mark>/g, 
                                  `<mark style="background-color: ${isDark ? '#fbbf24' : '#fef08a'}; color: ${isDark ? '#000' : '#000'}; padding: 2px 4px; border-radius: 2px; font-weight: 500;">`
                                )
                              }}
                            />
                          </div>
                        </div>
                      </a>
                    ))}
                  </div>
                )}
              </div>

              {/* Footer */}
              {searchResults.length > 0 && (
                <div className={`flex items-center justify-between px-4 py-2 text-xs border-t ${
                  isDark 
                    ? 'border-neutral-700 text-gray-500' 
                    : 'border-gray-200 text-gray-600'
                }`}>
                  <div className="flex items-center gap-4">
                    <span className="flex items-center gap-1">
                      <kbd className={`px-1.5 py-0.5 rounded ${isDark ? 'bg-neutral-800' : 'bg-gray-100'}`}>↑</kbd>
                      <kbd className={`px-1.5 py-0.5 rounded ${isDark ? 'bg-neutral-800' : 'bg-gray-100'}`}>↓</kbd>
                      to navigate
                    </span>
                    <span className="flex items-center gap-1">
                      <kbd className={`px-1.5 py-0.5 rounded ${isDark ? 'bg-neutral-800' : 'bg-gray-100'}`}>↵</kbd>
                      to select
                    </span>
                  </div>
                  <span>{searchResults.length} result{searchResults.length !== 1 ? 's' : ''}</span>
                </div>
              )}
            </div>
          </div>
        </>
      )}

      {isMenuOpen && (
        <div className={`md:hidden border-t ${
          isDark ? 'border-neutral-800 bg-[#111]' : 'border-gray-200 bg-white'
        }`}>
          <div className="px-4 py-3 space-y-1 max-h-[calc(100vh-4rem)] overflow-y-auto">
            <div className={`text-xs font-semibold uppercase px-2 py-1.5 ${
              isDark ? 'text-gray-400' : 'text-gray-500'
            }`}>Contribute</div>
            <Link href="/#join-in" className={`block px-3 py-2 text-sm rounded-md transition-colors ${
              isDark
                ? 'text-gray-300 hover:bg-neutral-800'
                : 'text-gray-700 hover:bg-gray-100'
            }`}>
              Join In
            </Link>
            <Link href="/contribute-handbook" className={`block px-3 py-2 text-sm rounded-md transition-colors ${
              isDark
                ? 'text-gray-300 hover:bg-neutral-800'
                : 'text-gray-700 hover:bg-gray-100'
            }`}>
              Contribute Handbook
            </Link>
            <Link href="/#security" className={`block px-3 py-2 text-sm rounded-md transition-colors ${
              isDark
                ? 'text-gray-300 hover:bg-neutral-800'
                : 'text-gray-700 hover:bg-gray-100'
            }`}>
              Security
            </Link>
            
            <div className={`text-xs font-semibold uppercase px-2 py-1.5 mt-3 ${
              isDark ? 'text-gray-400' : 'text-gray-500'
            }`}>Community</div>
            <Link href="/#get-involved" className={`block px-3 py-2 text-sm rounded-md transition-colors ${
              isDark
                ? 'text-gray-300 hover:bg-neutral-800'
                : 'text-gray-700 hover:bg-gray-100'
            }`}>
              Get Involved
            </Link>
            <Link href="/programs" className={`block px-3 py-2 text-sm rounded-md transition-colors ${
              isDark
                ? 'text-gray-300 hover:bg-neutral-800'
                : 'text-gray-700 hover:bg-gray-100'
            }`}>
              Programs
            </Link>
            <Link href="/#ladder" className={`block px-3 py-2 text-sm rounded-md transition-colors ${
              isDark
                ? 'text-gray-300 hover:bg-neutral-800'
                : 'text-gray-700 hover:bg-gray-100'
            }`}>
              Ladder
            </Link>
            <Link href="/#contact-us" className={`block px-3 py-2 text-sm rounded-md transition-colors ${
              isDark
                ? 'text-gray-300 hover:bg-neutral-800'
                : 'text-gray-700 hover:bg-gray-100'
            }`}>
              Contact Us
            </Link>
            <Link href="/#partners" className={`block px-3 py-2 text-sm rounded-md transition-colors ${
              isDark
                ? 'text-gray-300 hover:bg-neutral-800'
                : 'text-gray-700 hover:bg-gray-100'
            }`}>
              Partners
            </Link>

            <div className={`pt-3 border-t mt-3 ${
              isDark ? 'border-neutral-800' : 'border-gray-200'
            }`}>
              <a
                href="https://github.com/kubestellar/kubestellar"
                target="_blank"
                rel="noopener noreferrer"
                className={`flex items-center px-3 py-2 text-sm rounded-md transition-colors ${
                  isDark
                    ? 'text-gray-300 hover:bg-neutral-800'
                    : 'text-gray-700 hover:bg-gray-100'
                }`}
              >
                <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 0C5.374 0 0 5.373 0 12 0 17.302 3.438 21.8 8.207 23.387c.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.300 24 12c0-6.627-5.373-12-12-12z" />
                </svg>
                View on GitHub
                <span className={`ml-auto text-xs px-2 py-0.5 rounded ${
                  isDark ? 'bg-neutral-800' : 'bg-gray-200'
                }`}>
                  {githubStats.stars} ★
                </span>
              </a>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
