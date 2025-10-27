"use client";

import React, { useState, useEffect, useRef } from "react";
import Link from "next/link";
import Image from "next/image";
import { GridLines } from "./index";

const useMediaQuery = (query: string) => {
  const [matches, setMatches] = useState(false);

  useEffect(() => {
    const media = window.matchMedia(query);
    const updateMatches = () => {
      if (media.matches !== matches) {
        setMatches(media.matches);
      }
    };
    updateMatches();
    media.addEventListener("change", updateMatches);
    return () => media.removeEventListener("change", updateMatches);
  }, [matches, query]);

  return matches;
};


export default function Navigation() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [openMobileDropdown, setOpenMobileDropdown] = useState<string | null>(null);
  const isDesktop = useMediaQuery("(min-width: 1024px)");
  const [githubStats, setGithubStats] = useState({
    stars: "0",
    forks: "0",
    watchers: "0",
  });

  useEffect(() => {
    // Initialize dropdowns functionality
    const initDropdowns = () => {
      const dropdownContainers =
        document.querySelectorAll<HTMLElement>("[data-dropdown]");

      dropdownContainers.forEach(container => {
        const menu = container.querySelector<HTMLElement>(
          "[data-dropdown-menu]"
        );

        if (menu) {
          container.addEventListener("mouseenter", () => {
            if (timeoutRef.current) {
              clearTimeout(timeoutRef.current);
              timeoutRef.current = null;
            }

            dropdownContainers.forEach(otherContainer => {
              if (otherContainer !== container) {
                const otherMenu = otherContainer.querySelector<HTMLElement>(
                  "[data-dropdown-menu]"
                );

                if (otherMenu) {
                  otherMenu.style.display = "none";
                }
              }
            });

            menu.style.display = "block";
          });

          container.addEventListener("mouseleave", () => {
            timeoutRef.current = setTimeout(() => {
              menu.style.display = "none";
            }, 100);
          });

          menu.addEventListener("mouseenter", () => {
            if (timeoutRef.current) {
              clearTimeout(timeoutRef.current);
              timeoutRef.current = null;
            }
          });

          menu.addEventListener("mouseleave", () => {
            menu.style.display = "none";
          });
        }
      });

      // Close on Escape key
      document.addEventListener("keydown", e => {
        if (e.key === "Escape") {
          dropdownContainers.forEach(container => {
            const menu = container.querySelector(
              "[data-dropdown-menu]"
            ) as HTMLElement;
            if (menu) {
              menu.style.display = "none";
            }
          });
        }
      });
    };

    const fetchGithubStats = async () => {
      try {
        const response = await fetch("https://api.github.com/repos/kubestellar/kubestellar");
        if (!response.ok) {
          throw new Error("Network reposone was not okay");
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

    const createGrid = (container: HTMLElement) => {
      if (!container) return;
      container.innerHTML = "";

      const gridSvg = document.createElementNS(
        "http://www.w3.org/2000/svg",
        "svg"
      );
      gridSvg.setAttribute("width", "100%");
      gridSvg.setAttribute("height", "100%");
      gridSvg.style.position = "absolute";
      gridSvg.style.top = "0";
      gridSvg.style.left = "0";

      for (let i = 0; i < 10; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", "0");
        line.setAttribute("y1", `${i * 10}%`);
        line.setAttribute("x2", "100%");
        line.setAttribute("y2", `${i * 10}%`);
        line.setAttribute("stroke", "#6366F1");
        line.setAttribute("stroke-width", "0.5");
        line.setAttribute("stroke-opacity", "0.3");
        gridSvg.appendChild(line);
      }

      for (let i = 0; i < 10; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", `${i * 10}%`);
        line.setAttribute("y1", "0");
        line.setAttribute("x2", `${i * 10}%`);
        line.setAttribute("y2", "100%");
        line.setAttribute("stroke", "#6366F1");
        line.setAttribute("stroke-width", "0.5");
        line.setAttribute("stroke-opacity", "0.3");
        gridSvg.appendChild(line);
      }

      container.appendChild(gridSvg);
    };

    const gridContainer = document.getElementById("grid-lines-nav");

    if (gridContainer) createGrid(gridContainer);

    initDropdowns();
  }, []);

  return (
    <nav className="relative fixed w-full z-50 bg-gradient-to-br from-green-900 via-purple-900 to-green-900/90 backdrop-blur-lg border-b border-gray-700/50 transition-all duration-300">
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]/90 z-[-3]"></div>

      {/* Grid lines background */}
      <GridLines />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="flex justify-between h-16 items-center">
          {/* Left side: Logo */}
          <Link href="/" className="cursor-pointer flex-shrink-0">
            <div className="flex-shrink-0 cursor-pointer relative z-10">
              <Image
                src="/KubeStellar-with-Logo-transparent.png"
                alt="Kubestellar logo"
                width={240}
                height={60}
                className="h-12 w-auto xl:h-10"
              />
            </div>
          </Link>

          {/* Center: Nav Links */}
          <div className="main-nav-links absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
            <div className="flex items-center space-x-4 xl:space-x-8">
              {/* Docs Link */}
              <div className="relative group bg-white/1 backdrop-blur-xl rounded-lg border border-white/4 shadow-inner shadow-black/25">
                <Link
                  href="/docs"
                  className="text-sm font-medium text-gray-300 hover:text-blue-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-blue-500/10 hover:shadow-lg hover:shadow-blue-500/20 nav-link-hover"
                >
                  <div className="relative">
                    <svg
                      className="w-5 h-5 transition-all duration-300 group-hover:scale-102"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      ></path>
                    </svg>
                  </div>
                  <span>Docs</span>
                </Link>
              </div>

              {/* Blog Link */}
              <div className="relative group bg-white/1 backdrop-blur-xl rounded-lg border border-white/4 shadow-inner shadow-black/25">
                <Link
                  target="_blank"
                  href="https://kubestellar.medium.com/list/predefined:e785a0675051:READING_LIST"
                  className="text-sm font-medium text-gray-300 hover:text-purple-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-purple-500/10 hover:shadow-lg hover:shadow-purple-500/20 transform nav-link-hover"
                >
                  <div className="relative">
                    <svg
                      className="w-5 h-5 transition-all duration-300"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                      ></path>
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M16 8a2 2 0 012 2v6a2 2 0 01-2 2H8a2 2 0 01-2-2v-6a2 2 0 012-2h8z"
                        className="opacity-0 group-hover:opacity-100 transition-opacity duration-300"
                      ></path>
                    </svg>
                  </div>
                  <span>Blog</span>
                </Link>
              </div>

              {/* Contribute Dropdown */}
              <div className="relative group bg-white/1 backdrop-blur-xl rounded-lg border border-white/4 shadow-inner shadow-black/25" data-dropdown>
                <button
                  type="button"
                  className="text-sm font-medium text-gray-300 hover:text-emerald-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-emerald-500/10 hover:shadow-lg hover:shadow-emerald-500/20 hover:scale-100 transform nav-link-hover"
                  data-dropdown-button
                  aria-haspopup="true"
                  aria-expanded="false"
                >
                  <div className="relative">
                    <svg
                      className="w-5 h-5 transition-all duration-300 group-hover:scale-102"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                      ></path>
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                        className="opacity-0 group-hover:opacity-100 transition-opacity duration-300"
                      ></path>
                    </svg>
                  </div>
                  <span>Contribute</span>
                  <svg
                    className="ml-1 h-4 w-4 transition-transform duration-300"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                <div
                  className="absolute left-0 mt-2 w-56 bg-gray-800/90 backdrop-blur-lg rounded-xl shadow-2xl py-2 ring-1 ring-gray-700/50 transition-all duration-200 z-50 overflow-hidden"
                  data-dropdown-menu
                  style={{ display: "none" }}
                >
                  <a
                    href="#join-in"
                    className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"
                      ></path>
                    </svg>
                    Join In
                  </a>
                  <a
                    href="/community-handbook"
                    className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                      ></path>
                    </svg>
                    Contribute Handbook
                  </a>
                  <a
                    href="#security"
                    className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                      ></path>
                    </svg>
                    Security
                  </a>
                </div>
              </div>
              {/* Community Dropdown */}
              <div className="relative group bg-white/1 backdrop-blur-xl rounded-lg border border-white/4 shadow-inner shadow-black/25" data-dropdown>
                <button
                  type="button"
                  className="text-sm font-medium text-gray-300 hover:text-cyan-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-cyan-500/10 hover:shadow-lg hover:shadow-cyan-500/20 hover:scale-100 transform nav-link-hover"
                  data-dropdown-button
                  aria-haspopup="true"
                  aria-expanded="false"
                >
                  <div className="relative">
                    <svg
                      className="w-5 h-5 transition-all duration-300 group-hover:scale-102"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                      ></path>
                    </svg>
                  </div>
                  <span>Community</span>
                  <svg
                    className="ml-1 h-4 w-4 transition-transform duration-300"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                <div
                  className="absolute left-0 mt-2 w-56 bg-gray-800/90 backdrop-blur-lg rounded-xl shadow-2xl py-2 ring-1 ring-gray-700/50 transition-all duration-200 z-50 overflow-hidden"
                  data-dropdown-menu
                  style={{ display: "none" }}
                >
                  <a
                    href="#get-involved"
                    className="flex items-center px-10 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"
                      ></path>
                    </svg>
                    Get Involved
                  </a>
                  <Link
                    href="/programs"
                    className="flex items-center px-10 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2H5a2 2 0 00-2 2v2M7 7h10"
                      ></path>
                    </svg>
                    Programs
                  </Link>
                  <a
                    href="#ladder"
                    className="flex items-center px-10 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                      ></path>
                    </svg>
                    Ladder
                  </a>
                  <a
                    href="#contact-us"
                    className="flex items-center px-10 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M3 8l7.89 4.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                      ></path>
                    </svg>
                    Contact Us
                  </a>
                  <a
                    href="#partners"
                    className="flex items-center px-10 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                      ></path>
                    </svg>
                    Partners
                  </a>
                </div>
              </div>
            </div>
          </div>

          {/* Right side: Controls */}
          <div className="flex items-center">
            <div className="secondary-nav-controls flex items-center space-x-1 xl:space-x-2">
              {/* Version Dropdown */}
              <div
                className="relative"
                onMouseEnter={() => isDesktop && setOpenMobileDropdown("version")}
                onMouseLeave={() => isDesktop && setOpenMobileDropdown(null)}
              >
                <button
                  onClick={() => {
                    if (isDesktop) return;
                    setOpenMobileDropdown(
                      openMobileDropdown === "version" ? null : "version"
                    );
                  }}
                  className="w-full flex justify-between items-center px-4 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                >
                  {isDesktop ? (
                    <span>3.8.1</span>
                  ) : (
                    <span>Version</span>
                  )}
                  <svg
                    className={`w-5 h-5 transition-transform ${openMobileDropdown === "version" ? "rotate-180" : ""
                      }`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                {openMobileDropdown === "version" && (
                  <div className="pl-4 mt-1 space-y-1 lg:absolute lg:left-0 lg:top-full lg:mt-2 lg:w-44 lg:rounded-md lg:shadow-xl lg:py-1 lg:bg-gray-800/90 lg:ring lg:ring-gray-700/50">
                    <a href="#" className="block px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      3.8.1 (Current)
                    </a>
                    <a href="#" className="block px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      3.8.0
                    </a>
                    <a href="#" className="block px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      All versions
                    </a>
                  </div>
                )}
              </div>

              {/* Language Dropdown */}
              <div
                className="relative"
                onMouseEnter={() => isDesktop && setOpenMobileDropdown("language")}
                onMouseLeave={() => isDesktop && setOpenMobileDropdown(null)}
              >
                <button
                  onClick={() => {
                    if (isDesktop) return;
                    setOpenMobileDropdown(
                      openMobileDropdown === "language" ? null : "language"
                    );
                  }}
                  className="w-full flex justify-between items-center px-4 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                >
                  {isDesktop ? (
                    <svg
                      className="w-4 h-4 xl:mr-2"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129"
                      />
                    </svg>
                  ) : (
                    <span>Language</span>
                  )}
                  <svg
                    className={`w-5 h-5 transition-transform ${openMobileDropdown === "language" ? "rotate-180" : ""
                      }`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                {openMobileDropdown === "language" && (
                  <div className="pl-4 mt-1 space-y-1 lg:absolute lg:left-0 lg:top-full lg:mt-2 lg:w-44 lg:rounded-md lg:shadow-xl lg:py-1 lg:bg-gray-800/90 lg:ring lg:ring-gray-700/50">
                    <a href="#" className="block px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      English
                    </a>
                    <a href="#" className="block px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      日本語
                    </a>
                    <a href="#" className="block px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      简体中文
                    </a>
                  </div>
                )}
              </div>

              {/* GitHub Dropdown */}
              <div
                className="relative"
                onMouseEnter={() => isDesktop && setOpenMobileDropdown("github")}
                onMouseLeave={() => isDesktop && setOpenMobileDropdown(null)}
              >
                <button
                  onClick={() => {
                    if (isDesktop) return;
                    setOpenMobileDropdown(
                      openMobileDropdown === "github" ? null : "github"
                    );
                  }}
                  className="w-full flex justify-between items-center px-4 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                >
                  {isDesktop ? (
                    <svg
                      className="w-4 h-4 mr-2"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path d="M12 0C5.374 0 0 5.373 0 12 0 17.302 3.438 21.8 8.207 23.387c.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.300 24 12c0-6.627-5.373-12-12-12z" />
                    </svg>
                  ) : (
                    <span>Github</span>
                  )}
                  <svg
                    className={`w-5 h-5 transition-transform ${openMobileDropdown === "github" ? "rotate-180" : ""
                      }`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                {openMobileDropdown === "github" && (
                  <div className="pl-4 mt-1 space-y-1 lg:absolute lg:left-0 lg:top-full lg:mt-2 lg:w-44 lg:rounded-md lg:shadow-xl lg:py-1 lg:bg-gray-800/90 lg:ring lg:ring-gray-700/50">
                    <a href="https://github.com/kubestellar/kubestellar" className="flex justify-between items-center px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      Star
                      <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                        {githubStats.stars}
                      </span>
                    </a>
                    <a href="https://github.com/kubestellar/kubestellar/fork" className="flex justify-between items-center px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      Fork
                      <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                        {githubStats.forks}
                      </span>
                    </a>
                    <a href="https://github.com/kubestellar/kubestellar/watchers" className="flex justify-between items-center px-3 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                      Watch
                      <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                        {githubStats.watchers}
                      </span>
                    </a>
                  </div>
                )}
              </div>
            </div>
            {/* Mobile menu button */}
            <button
              className="mobile-menu-button p-2 rounded focus:outline-none hover:bg-gray-100 dark:hover:bg-gray-700"
              aria-label="Open menu"
              onClick={() => setIsMenuOpen(!isMenuOpen)}
            >
              <svg
                className="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            </button>
          </div>
        </div>

        {/* Mobile menu */}
        {isMenuOpen && (
          <div className="mobile-menu-panel">
            <div className="px-5 mt-4 pb-3 sm:px-3">
              {/* Primary Links for Mobile */}
              <div className="mobile-primary-links">
                <a
                  href="/docs"
                  className="block px-10 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                >
                  Docs
                </a>
                <a
                  href="https://kubestellar.medium.com/list/predefined:e785a0675051:READING_LIST"
                  target="_blank"
                  className="block px-10 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                >
                  Blog
                </a>

                {/* Contribute Dropdown for Mobile */}
                <div>
                  <button
                    onClick={() =>
                      setOpenMobileDropdown(
                        openMobileDropdown === "contribute" ? null : "contribute"
                      )
                    }
                    className="w-full flex justify-between items-center px-10 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                  >
                    <span>Contribute</span>
                    <svg
                      className={`w-5 h-5 transition-transform ${openMobileDropdown === "contribute" ? "rotate-180" : ""
                        }`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                  </button>
                  {openMobileDropdown === "contribute" && (
                    <div className="pl-4 mt-1 space-y-1">
                      <a
                        href="#join-in"
                        className="block px-10 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700"
                      >
                        Join In
                      </a>
                      <a
                        href="/community-handbook"
                        className="block px-10 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700"
                      >
                        Contribute Handbook
                      </a>
                      <a
                        href="#security"
                        className="block px-10 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700"
                      >
                        Security
                      </a>
                    </div>
                  )}
                </div>

                {/* Community Dropdown for Mobile */}
                <div>
                  <button
                    onClick={() =>
                      setOpenMobileDropdown(
                        openMobileDropdown === "community" ? null : "community"
                      )
                    }
                    className="w-full flex justify-between items-center px-10 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                  >
                    <span>Community</span>
                    <svg
                      className={`w-5 h-5 transition-transform ${openMobileDropdown === "community" ? "rotate-180" : ""
                        }`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                  </button>
                  {openMobileDropdown === "community" && (
                    <div className="pl-4 mt-1 space-y-1">
                      <a
                        href="#get-involved"
                        className="block px-10 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700"
                      >
                        Get Involved
                      </a>
                      <Link
                        href="/programs"
                        className="block px-10 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700"
                      >
                        Programs
                      </Link>
                      <a
                        href="#ladder"
                        className="block px-10 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700"
                      >
                        Ladder
                      </a>
                      <a
                        href="#contact-us"
                        className="block px-10 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700"
                      >
                        Contact Us
                      </a>
                      <a
                        href="#partners"
                        className="block px-10 py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700"
                      >
                        Partners
                      </a>
                    </div>
                  )}
                </div>
              </div>

              {/* Divider and Secondary Controls for mobile menu */}
              <div className="mobile-secondary-controls">
                <div className="flex flex-col">

                  {/* Version Dropdown */}
                  <div className="mobile-ipad">
                    <button
                      onClick={() =>
                        setOpenMobileDropdown(
                          openMobileDropdown === "version" ? null : "version"
                        )
                      }
                      className="w-full flex justify-between items-center py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                    >
                      <span>Version</span>
                      <svg
                        className={`w-5 h-5 transition-transform ${openMobileDropdown === "version" ? "rotate-180" : ""
                          }`}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M19 9l-7 7-7-7"
                        />
                      </svg>
                    </button>
                    {openMobileDropdown === "version" && (
                      <div className="mobile-wid pl-5 mt-1 space-y-1">
                        <a href="#" className="block py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          3.8.1 (Current)
                        </a>
                        <a href="#" className="block py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          3.8.0
                        </a>
                        <a href="#" className="block py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          All versions
                        </a>
                      </div>
                    )}
                  </div>

                  {/* Language Dropdown */}
                  <div className="mobile-ipad">
                    <button
                      onClick={() =>
                        setOpenMobileDropdown(
                          openMobileDropdown === "language" ? null : "language"
                        )
                      }
                      className="w-full flex justify-between items-center py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                    >
                      <span>Language</span>
                      <svg
                        className={`w-5 h-5 transition-transform ${openMobileDropdown === "language" ? "rotate-180" : ""
                          }`}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M19 9l-7 7-7-7"
                        />
                      </svg>
                    </button>
                    {openMobileDropdown === "language" && (
                      <div className="mobile-wid pl-5 mt-1 space-y-1">
                        <a href="#" className="block py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          English
                        </a>
                        <a href="#" className="block py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          日本語
                        </a>
                        <a href="#" className="block py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          简体中文
                        </a>
                      </div>
                    )}
                  </div>

                  {/* GitHub Dropdown */}
                  <div className="mobile-ipad">
                    <button
                      onClick={() =>
                        setOpenMobileDropdown(
                          openMobileDropdown === "github" ? null : "github"
                        )
                      }
                      className="w-full flex justify-between items-center py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
                    >
                      <span>GitHub</span>
                      <svg
                        className={`w-5 h-5 transition-transform ${openMobileDropdown === "github" ? "rotate-180" : ""
                          }`}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M19 9l-7 7-7-7"
                        />
                      </svg>
                    </button>
                    {openMobileDropdown === "github" && (
                      <div className="mobile-wid pl-5 mt-1 space-y-1">
                        <a href="https://github.com/kubestellar/kubestellar" className="flex justify-between items-center py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          Star
                          <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                            {githubStats.stars}
                          </span>
                        </a>
                        <a href="https://github.com/kubestellar/kubestellar/fork" className="flex justify-between items-center py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          Fork
                          <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                            {githubStats.forks}
                          </span>
                        </a>
                        <a href="https://github.com/kubestellar/kubestellar/watchers" className="flex justify-between items-center py-2 rounded-md text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-700">
                          Watch
                          <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                            {githubStats.watchers}
                          </span>
                        </a>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </nav>
  );
}