"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { GridLines, StarField } from "../index";
import Link from "next/link";
import { useTheme } from "next-themes";

export default function Footer() {
  const [mounted, setMounted] = useState(false);
  const { resolvedTheme } = useTheme();
  const isDark = resolvedTheme === 'dark';

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    // Back to top functionality
    const backToTopButton = document.getElementById("back-to-top");
    if (!backToTopButton) return;

    const toggleButton = () => {
      if (window.pageYOffset > 300) {
        backToTopButton.style.opacity = "1";
        backToTopButton.style.transform = "translateY(-30px)";
      } else {
        backToTopButton.style.opacity = "0";
        backToTopButton.style.transform = "translateY(10px)";
      }
    };

    const handleClick = () => {
      window.scrollTo({
        top: 0,
        behavior: "smooth",
      });
    };

    window.addEventListener("scroll", toggleButton);
    backToTopButton.addEventListener("click", handleClick);

    // Initial check
    toggleButton();

    // Cleanup function to prevent memory leaks
    return () => {
      window.removeEventListener("scroll", toggleButton);
      backToTopButton.removeEventListener("click", handleClick);
    };
  }, []);

  // Prevent hydration mismatch by rendering dark theme until mounted
  if (!mounted) {
    return (
      <footer className="bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white relative overflow-hidden pt-8 sm:pt-12 md:pt-16 pb-6 sm:pb-8">
        <div className="absolute inset-0 bg-[#0a0a0a]"></div>
        <StarField density="low" showComets={true} cometCount={2} />
        <GridLines horizontalLines={21} verticalLines={15} />
        <div className="absolute inset-0 z-0">
          <div className="absolute bottom-0 left-0 w-full h-1/2 bg-gradient-to-t from-blue-900/10 to-transparent"></div>
          <div className="absolute top-0 right-0 w-full h-1/2 bg-gradient-to-b from-purple-900/10 to-transparent"></div>
        </div>
        <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-10 gap-6 sm:gap-4 lg:gap-12 mb-8 sm:mb-10 lg:mb-12">
            <div className="col-span-1 sm:col-span-3 lg:col-span-4 mb-4 sm:mb-0">
              <div className="flex items-center space-x-2 mb-3 sm:mb-4">
                <Image
                  src="/KubeStellar-with-Logo-transparent.png"
                  alt="Kubestellar logo"
                  width={160}
                  height={40}
                  className="h-8 sm:h-9 md:h-10 w-auto"
                />
              </div>
              <p className="text-gray-300 mb-4 sm:mb-6 leading-relaxed text-base sm:text-base">
                Multi-Cluster Kubernetes orchestration platform that simplifies
                distributed workload management across diverse infrastructure.
              </p>
            </div>
          </div>
        </div>
      </footer>
    );
  }

  return (
    <footer className={`relative overflow-hidden pt-8 sm:pt-12 md:pt-16 pb-6 sm:pb-8 ${
      isDark
        ? 'bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white'
        : 'bg-gradient-to-br from-blue-50 via-purple-50 to-blue-50 text-gray-900'
    }`}>
      {/* Base background */}
      <div className={`absolute inset-0 ${isDark ? 'bg-[#0a0a0a]' : 'bg-white/80'}`}></div>

      {/* Starfield background */}
      <StarField density="low" showComets={true} cometCount={2} />

      {/* Grid lines background */}
      <GridLines horizontalLines={21} verticalLines={15} />

      {/* Background elements */}
      <div className="absolute inset-0 z-0">
        <div className={`absolute bottom-0 left-0 w-full h-1/2 bg-gradient-to-t ${
          isDark ? 'from-blue-900/10' : 'from-blue-200/30'
        } to-transparent`}></div>
        <div className={`absolute top-0 right-0 w-full h-1/2 bg-gradient-to-b ${
          isDark ? 'from-purple-900/10' : 'from-purple-200/30'
        } to-transparent`}></div>
      </div>

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Main footer content */}
        <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-10 gap-6 sm:gap-4 lg:gap-12 mb-8 sm:mb-10 lg:mb-12">
          {/* Brand Section */}
          <div className="col-span-1 sm:col-span-3 lg:col-span-4 mb-4 sm:mb-0">
            <div className="flex items-center space-x-2 mb-3 sm:mb-4">
              <Image
                src="/KubeStellar-with-Logo-transparent.png"
                alt="Kubestellar logo"
                width={160}
                height={40}
                className="h-8 sm:h-9 md:h-10 w-auto"
              />
            </div>
            <p className={`mb-4 sm:mb-6 leading-relaxed text-base sm:text-base ${
              isDark ? 'text-gray-300' : 'text-gray-700'
            }`}>
              Multi-Cluster Kubernetes orchestration platform that simplifies
              distributed workload management across diverse infrastructure.
            </p>
            <div className="flex space-x-3 sm:space-x-4">
              <a
                href="https://x.com/KubeStellar"
                className="group relative w-12 h-12 rounded-lg flex items-center justify-center transition-all duration-300"
                aria-label="X (Twitter)"
              >
                <svg
                  className={`w-5 h-5 sm:w-6 sm:h-6 md:w-7 md:h-7 transition-all duration-300 ${
                    isDark
                      ? 'group-hover:drop-shadow-[0_0_8px_rgba(255,255,255,0.8)]'
                      : 'group-hover:drop-shadow-[0_0_8px_rgba(0,0,0,0.3)]'
                  }`}
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    className={`transition-colors duration-300 ${
                      isDark
                        ? 'text-gray-400 group-hover:text-white'
                        : 'text-gray-600 group-hover:text-gray-900'
                    }`}
                    d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"
                  />
                </svg>
              </a>
              <a
                href="https://github.com/kubestellar"
                className="group relative w-9 h-9 sm:w-10 sm:h-10 md:w-12 md:h-12 rounded-lg flex items-center justify-center transition-all duration-300"
                aria-label="GitHub"
              >
                <svg
                  className={`w-5 h-5 sm:w-6 sm:h-6 md:w-7 md:h-7 transition-all duration-300 ${
                    isDark
                      ? 'group-hover:drop-shadow-[0_0_8px_rgba(255,255,255,0.6)]'
                      : 'group-hover:drop-shadow-[0_0_8px_rgba(0,0,0,0.3)]'
                  }`}
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    className={`transition-colors duration-300 ${
                      isDark
                        ? 'text-gray-400 group-hover:text-white'
                        : 'text-gray-600 group-hover:text-gray-900'
                    }`}
                    d="M12 0C5.374 0 0 5.373 0 12 0 17.302 3.438 21.8 8.207 23.387c.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.300 24 12c0-6.627-5.373-12-12-12z"
                  />
                </svg>
              </a>
              <a
                href="https://www.linkedin.com/company/kubestellar/"
                className="group relative w-9 h-9 sm:w-10 sm:h-10 md:w-12 md:h-12 rounded-lg flex items-center justify-center transition-all duration-300"
                aria-label="LinkedIn"
              >
                <svg
                  className="w-5 h-5 sm:w-6 sm:h-6 md:w-7 md:h-7 transition-all duration-300 group-hover:drop-shadow-[0_0_8px_rgba(0,119,181,0.8)]"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    className={`transition-colors duration-300 ${
                      isDark ? 'text-gray-400' : 'text-gray-600'
                    } group-hover:text-[#0077B5]`}
                    d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"
                  />
                </svg>
              </a>
              <a
                href="https://www.youtube.com/@kubestellar"
                className="group relative w-9 h-9 sm:w-10 sm:h-10 md:w-12 md:h-12 rounded-lg flex items-center justify-center transition-all duration-300"
                aria-label="YouTube"
              >
                <svg
                  className="w-5 h-5 sm:w-6 sm:h-6 md:w-7 md:h-7 transition-all duration-300 group-hover:drop-shadow-[0_0_8px_rgba(255,0,0,0.8)]"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    className={`transition-colors duration-300 ${
                      isDark ? 'text-gray-400' : 'text-gray-600'
                    } group-hover:text-[#FF0000]`}
                    d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"
                  />
                </svg>
              </a>
            </div>
          </div>


          {/* Navigation Links Container - 3 columns on mobile */}
          <div className="col-span-1 sm:col-span-3 lg:col-span-6 grid grid-cols-3 gap-4 sm:gap-4 lg:gap-8">
            {/* Docs Links */}
            <div>
              <h3 className={`text-sm sm:text-base md:text-lg font-semibold mb-2 sm:mb-4 ${
                isDark ? 'text-white' : 'text-gray-900'
              }`}>
                Docs
              </h3>
              <ul className="space-y-1 sm:space-y-3">
                <li>
                  <Link
                    href="/docs"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Overview
                  </Link>
                </li>
                <li>
                  <Link
                    href="/docs/user-guide/guide-overview"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    User Guide
                  </Link>
                </li>
                <li>
                  <Link
                    href="/docs/contribution-guidelines/onboarding-inc"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Onboarding
                  </Link>
                </li>
                <li>
                  <Link
                    href="/docs/what-is-kubestellar/release-notes"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Release Notes
                  </Link>
                </li>
              </ul>
            </div>

            {/* Getting Started Links */}
            <div>
              <h3 className={`text-sm sm:text-base md:text-lg font-semibold mb-2 sm:mb-4 ${
                isDark ? 'text-white' : 'text-gray-900'
              }`}>
                Getting Started
              </h3>
              <ul className="space-y-1 sm:space-y-3">
                <li>
                  <Link
                    href="/quick-installation"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Installation Page
                  </Link>
                </li>
                <li>
                  <Link
                    href="/ladder"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Ladder
                  </Link>
                </li>
                <li>
                  <Link
                    href="/products"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Products
                  </Link>
                </li>
                <li>
                  <Link
                    href="/contribute-handbook"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Contributor Handbook
                  </Link>
                </li>
              </ul>
            </div>

            {/* Resources Links */}
            <div>
              <h3 className={`text-sm sm:text-base md:text-lg font-semibold mb-2 sm:mb-4 ${
                isDark ? 'text-white' : 'text-gray-900'
              }`}>
                Resources
              </h3>
              <ul className="space-y-1 sm:space-y-3">
                <li>
                  <Link
                    href="/playground"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Playground
                  </Link>
                </li>
                <li>
                  <Link
                    href="/programs"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Programs
                  </Link>
                </li>
                <li>
                  <Link
                    href="/partners"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Partners
                  </Link>
                </li>
                <li>
                  <a
                    href="https://blog.kubestellar.io"
                    target="_blank"
                    rel="noopener noreferrer"
                    className={`text-xs sm:text-sm transition-colors duration-200 inline-block ${
                      isDark
                        ? 'text-gray-400 hover:text-white'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    Blog
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </div>

        {/* Newsletter Section */}
        <div className="flex flex-col items-center md:items-end justify-center w-full mb-6 sm:mb-8">
          <div className="w-full max-w-3xl lg:pr-28">
            <div className="flex flex-col md:flex-row items-center md:items-center justify-between lg:pl-12 gap-4 mb-4">
              {/* Title */}
              <div className="flex items-center justify-center w-full md:w-auto text-center md:text-left">
                <h3 className={`text-sm sm:text-sm md:text-base font-semibold uppercase tracking-wide whitespace-nowrap ${
                  isDark ? 'text-white' : 'text-gray-900'
                }`}>
                  Stay Updated
                </h3>
              </div>

              {/* Form container */}
              <div className="flex-1 w-full md:w-auto">
                <form
                  id="newsletter-form"
                  className="flex flex-col sm:flex-row gap-3 items-center w-full sm:w-auto"
                >
                  <div className="relative flex-1 w-full min-w-[260px] sm:min-w-[280px] md:min-w-[300px]">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-4 w-4 text-gray-400"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                        />
                      </svg>
                    </div>

                    <input
                      id="email-address"
                      type="email"
                      className={`block w-full pl-10 pr-3 py-3 text-sm border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200 ${
                        isDark
                          ? 'text-white placeholder-gray-400 bg-gray-700/50 border-gray-600'
                          : 'text-gray-900 placeholder-gray-500 bg-white border-gray-300'
                      }`}
                      placeholder="Email"
                      required
                    />
                  </div>

                  <button
                    type="submit"
                    className={`w-full sm:w-auto px-6 py-3 text-sm font-medium text-white bg-gradient-to-r from-blue-600 to-purple-600 border border-transparent rounded-lg shadow-sm hover:from-blue-700 hover:to-purple-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-all duration-200 transform hover:-translate-y-0.5 whitespace-nowrap ${
                      isDark ? 'focus:ring-offset-gray-800' : 'focus:ring-offset-white'
                    }`}
                  >
                    <span>Subscribe</span>
                  </button>
                </form>
              </div>
            </div>
          </div>
        </div>

        {/* Divider and bottom section */}
        <div className={`border-t pt-4 sm:pt-6 md:pt-8 ${
          isDark ? 'border-gray-800' : 'border-gray-300'
        }`}>
          <div className="flex flex-col md:flex-row justify-between items-center gap-3 sm:gap-4">
            {/* Left side - copyright */}
            <p className={`text-xs sm:text-sm text-center md:text-left order-2 md:order-1 ${
              isDark ? 'text-gray-400' : 'text-gray-600'
            }`}>
              © 2025 KubeStellar. All rights reserved. Apache 2.0 Licence
            </p>

            {/* Right side - policy links */}
            <div className="flex flex-wrap items-center justify-center gap-3 sm:gap-4 md:gap-6 lg:gap-8 order-1 md:order-2">
              <Link
                href="/docs/contribution-guidelines/license-inc"
                className={`text-xs sm:text-sm transition-colors duration-300 whitespace-nowrap ${
                  isDark
                    ? 'text-gray-400 hover:text-white'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Privacy Policy
              </Link>
              <Link
                href="/docs/contribution-guidelines/security/security-inc"
                className={`text-xs sm:text-sm transition-colors duration-300 whitespace-nowrap ${
                  isDark
                    ? 'text-gray-400 hover:text-white'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Terms of Service
              </Link>
              <Link
                href="/docs/contribution-guidelines/security/security_contacts-inc"
                className={`text-xs sm:text-sm transition-colors duration-300 whitespace-nowrap ${
                  isDark
                    ? 'text-gray-400 hover:text-white'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Cookie Policy
              </Link>
            </div>
          </div>
        </div>
      </div>

      {/* Floating back to top button */}
      <button
        id="back-to-top"
        className="fixed bottom-18 right-8 p-3 rounded-full bg-blue-600 text-white shadow-lg z-40 transition-all duration-300 opacity-0 translate-y-10 hover:bg-blue-700 hover:scale-110"
        aria-label="Back to top"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-5 w-5 sm:h-5 sm:w-5 md:h-6 md:w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M5 10l7-7m0 0l7 7m-7-7v18"
          />
        </svg>
      </button>
    </footer>
  );
}