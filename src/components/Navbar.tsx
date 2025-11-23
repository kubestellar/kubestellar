"use client";

import React, { useState, useEffect, useRef } from "react";
import Link from "next/link";
import Image from "next/image";
import { GridLines, StarField, LanguageSwitcher } from "./index";
import { useTranslations } from "next-intl";

export default function Navbar() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [githubStats, setGithubStats] = useState({
    stars: "0",
    forks: "0",
    watchers: "0",
  });

  const t = useTranslations("navigation");
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
          const showMenu = () => {
            if (timeoutRef.current) {
              clearTimeout(timeoutRef.current);
              timeoutRef.current = null;
            }

            // Close all other dropdowns including language switcher
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

            // Close language switcher when hovering other dropdowns
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            if (typeof (window as any).closeLangSwitcher === "function") {
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              (window as any).closeLangSwitcher();
            }

            // Ensure menu is visible
            menu.style.display = "block";
            menu.style.opacity = "1";
            menu.style.visibility = "visible";
            setIsDropdownOpen(true);
          };

          const hideMenu = () => {
            timeoutRef.current = setTimeout(() => {
              menu.style.display = "none";
              menu.style.opacity = "0";
              menu.style.visibility = "hidden";
              setIsDropdownOpen(false);
            }, 300);
          };

          container.addEventListener("mouseenter", showMenu);
          container.addEventListener("mouseleave", hideMenu);

          menu.addEventListener("mouseenter", () => {
            if (timeoutRef.current) {
              clearTimeout(timeoutRef.current);
              timeoutRef.current = null;
            }
          });

          menu.addEventListener("mouseleave", hideMenu);
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
              menu.style.opacity = "0";
              menu.style.visibility = "hidden";
            }
          });

          // Also close language switcher
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          if (typeof (window as any).closeLangSwitcher === "function") {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (window as any).closeLangSwitcher();
          }

          setIsDropdownOpen(false);
        }
      });
    };

    const fetchGithubStats = async () => {
      try {
        const response = await fetch(
          "https://api.github.com/repos/kubestellar/kubestellar"
        );
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

    // Initialize LanguageSwitcher hover functionality
    const initLanguageSwitcher = () => {
      const langSwitcher = document.querySelector<HTMLElement>(
        ".language-switcher-container"
      );

      if (langSwitcher) {
        let isLangHovered = false;
        let langDropdownElement: HTMLElement | null = null;

        const handleMouseEnter = () => {
          if (timeoutRef.current) {
            clearTimeout(timeoutRef.current);
            timeoutRef.current = null;
          }

          // Close other dropdowns
          const dropdownContainers =
            document.querySelectorAll<HTMLElement>("[data-dropdown]");
          dropdownContainers.forEach(container => {
            const menu = container.querySelector<HTMLElement>(
              "[data-dropdown-menu]"
            );
            if (menu) {
              menu.style.display = "none";
            }
          });

          // Trigger LanguageSwitcher open
          const langButton = langSwitcher.querySelector("button");
          if (langButton) {
            // Check if dropdown is currently open
            const dropdown = document.querySelector('[role="listbox"]');
            const isDropdownVisible =
              dropdown && window.getComputedStyle(dropdown).display !== "none";

            if (!isDropdownVisible) {
              isLangHovered = true;
              langButton.click();
              setIsDropdownOpen(true);
            } else {
              isLangHovered = true;
            }
          }
        };

        const handleMouseLeave = () => {
          timeoutRef.current = setTimeout(() => {
            const langButton = langSwitcher.querySelector("button");
            const dropdown = document.querySelector('[role="listbox"]');
            const isDropdownVisible =
              dropdown && window.getComputedStyle(dropdown).display !== "none";

            if (langButton && isLangHovered && isDropdownVisible) {
              isLangHovered = false;
              langButton.click();
              setIsDropdownOpen(false);
            } else if (!isDropdownVisible) {
              isLangHovered = false;
              setIsDropdownOpen(false);
            }
          }, 300);
        };

        const closeLangSwitcher = () => {
          const langButton = langSwitcher.querySelector("button");
          const dropdown = document.querySelector('[role="listbox"]');
          const isDropdownVisible =
            dropdown && window.getComputedStyle(dropdown).display !== "none";

          if (langButton && isDropdownVisible) {
            isLangHovered = false;
            langButton.click();
            setIsDropdownOpen(false);
          } else if (isLangHovered) {
            isLangHovered = false;
            setIsDropdownOpen(false);
          }
        };

        // Add method to global scope for other dropdowns to call
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (window as any).closeLangSwitcher = closeLangSwitcher;

        langSwitcher.addEventListener("mouseenter", handleMouseEnter);
        langSwitcher.addEventListener("mouseleave", handleMouseLeave);

        // Handle dropdown menu hover with improved detection
        const observer = new MutationObserver(mutations => {
          mutations.forEach(mutation => {
            mutation.addedNodes.forEach(node => {
              if (node.nodeType === Node.ELEMENT_NODE) {
                const element = node as HTMLElement;
                const dropdown =
                  (element.querySelector?.(
                    '[role="listbox"]'
                  ) as HTMLElement) ||
                  (element.getAttribute?.("role") === "listbox"
                    ? (element as HTMLElement)
                    : null);

                if (dropdown) {
                  langDropdownElement = dropdown;

                  dropdown.addEventListener("mouseenter", () => {
                    if (timeoutRef.current) {
                      clearTimeout(timeoutRef.current);
                      timeoutRef.current = null;
                    }
                  });

                  dropdown.addEventListener("mouseleave", () => {
                    handleMouseLeave();
                  });
                }
              }
            });

            // Handle removed nodes (when dropdown closes)
            mutation.removedNodes.forEach(node => {
              if (node.nodeType === Node.ELEMENT_NODE) {
                const element = node as HTMLElement;
                if (
                  element === langDropdownElement ||
                  element.contains?.(langDropdownElement)
                ) {
                  langDropdownElement = null;
                  if (isLangHovered) {
                    isLangHovered = false;
                    setIsDropdownOpen(false);
                  }
                }
              }
            });
          });
        });

        observer.observe(document.body, { childList: true, subtree: true });
      }
    };

    // Small delay to ensure LanguageSwitcher is mounted
    setTimeout(initLanguageSwitcher, 100);
  }, []);

  return (
    <>
      {/* Blur overlay when dropdown is open */}
      {isDropdownOpen && (
        <div
          className="fixed inset-0 bg-black/20 backdrop-blur-md z-40 transition-all duration-300"
          style={{ backdropFilter: "blur(8px)" }}
        />
      )}

      <nav className="fixed w-full z-50 bg-gradient-to-br from-green-900 via-purple-900 to-green-900/90 backdrop-blur-md border-b border-gray-700/50 transition-all duration-300">
        {/* Dark base background */}
        <div className="absolute inset-0 bg-[#0a0a0a]/90 z-[-3]"></div>

        {/* Starfield background */}
        <StarField
          density="low"
          showComets={true}
          cometCount={2}
          className="z-[-2]"
        />

        {/* Grid lines background */}
        <GridLines />

        <div className="max-w-7xl mx-auto px-0.5 sm:px-2 lg:px-1 relative">
          <div className="flex justify-between h-16 items-center">
            {/* Left side: Logo */}
            <Link href="/" className="cursor-pointer">
              <div className="flex-shrink-0 cursor-pointer relative z-10">
                <Image
                  src="/KubeStellar-with-Logo-transparent.png"
                  alt="Kubestellar logo"
                  width={160}
                  height={40}
                  className="h-10 w-auto object-contain"
                />
              </div>
            </Link>

            {/* Center: Nav Links */}
            <div className="hidden lg:flex flex-1 justify-center">
              <div className="flex items-center space-x-8">
                {/* Docs Link */}
                <div className="relative group">
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
                    <span>{t("docs")}</span>
                  </Link>
                </div>

                {/* Playground Link */}
                <div className="relative group">
                  <Link
                    href="/playground"
                    className="text-sm font-medium text-gray-300 hover:text-orange-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-orange-500/10 hover:shadow-lg hover:shadow-orange-500/20 transform nav-link-hover"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 transition-all duration-300 group-hover:scale-110"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
                          className="group-hover:stroke-[2.5] transition-all duration-300"
                        ></path>
                      </svg>
                    </div>
                    <span>{t("playground")}</span>
                  </Link>
                </div>

                {/* Marketplace Link */}
                <div className="relative group">
                  <Link
                    href="/marketplace"
                    className="text-sm font-medium text-gray-300 hover:text-pink-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-pink-500/10 hover:shadow-lg hover:shadow-pink-500/20 transform nav-link-hover"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 transition-all duration-300 group-hover:scale-110"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"
                        ></path>
                      </svg>
                    </div>
                    <span>{t("marketplace")}</span>
                  </Link>
                </div>

                {/* Contribute Dropdown */}
                <div className="relative group" data-dropdown>
                  <button
                    type="button"
                    className="text-sm font-medium text-gray-300 hover:text-emerald-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-emerald-500/10 hover:shadow-lg hover:shadow-emerald-500/20 hover:scale-100 transform nav-link-hover cursor-pointer"
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
                    <span>{t("contribute")}</span>
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
                    className="absolute left-0 mt-1 w-56 bg-gray-800/90 backdrop-blur-md rounded-xl shadow-2xl py-2 ring-1 ring-gray-700/50 transition-all duration-200 z-50 before:content-[''] before:absolute before:bottom-full before:left-0 before:right-0 before:h-2 before:bg-transparent"
                    data-dropdown-menu
                    style={{ display: "none" }}
                  >
                    <a
                      href="https://kubestellar.io/joinus"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
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
                      {t("joinIn")}
                    </a>
                    <Link
                      href="/contribute-handbook"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
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
                      {t("contributeHandbook")}
                    </Link>
                    <Link
                      href="/quick-installation"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
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
                          d="M13 10V3L4 14h7v7l9-11h-7z"
                        ></path>
                      </svg>
                      {t("quickInstallation")}
                    </Link>
                    <Link
                      href="/products"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
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
                      {t("products")}
                    </Link>
                    <Link
                      href="/ladder"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
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
                      {t("ladder")}
                    </Link>
                    <Link
                      href="/docs/contribution-guidelines/security/security-inc"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
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
                      {t("security")}
                    </Link>
                  </div>
                </div>
                {/* Community Dropdown */}
                <div className="relative group" data-dropdown>
                  <button
                    type="button"
                    className="text-sm font-medium text-gray-300 hover:text-cyan-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-cyan-500/10 hover:shadow-lg hover:shadow-cyan-500/20 hover:scale-100 transform nav-link-hover cursor-pointer"
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
                    <span>{t("community")}</span>
                    <svg
                      className="ml-1 h-4 w-4 transition-transform duration-300 "
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
                    className="absolute left-0  mt-1 w-56 bg-gray-800/90 backdrop-blur-md rounded-xl shadow-2xl py-2 ring-1 ring-gray-700/50 transition-all duration-200 z-50 before:content-[''] before:absolute before:bottom-full before:left-0 before:right-0 before:h-2 before:bg-transparent"
                    data-dropdown-menu
                    style={{ display: "none" }}
                  >
                    <Link
                      href="/docs"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 rounded transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
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
                      {t("getInvolved")}
                    </Link>
                    <Link
                      href="/programs"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 rounded transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
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
                          d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
                        ></path>
                      </svg>
                      {t("programs")}
                    </Link>
                    <a
                      href="#contact-us"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 rounded transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
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
                      {t("contactUs")}
                    </a>
                    <Link
                      href="/partners"
                      className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-cyan-900/30 rounded transition-all duration-200 hover:text-cyan-300 hover:shadow-md"
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
                      {t("partners")}
                    </Link>
                  </div>
                </div>
              </div>
            </div>

            {/* Right side: Controls */}
            <div className="flex items-center sm:space-x-4">
              {/* Language Switcher */}
              <div className="language-switcher-container">
                <LanguageSwitcher className="relative group" />
              </div>

              {/* GitHub Dropdown */}
              <div className="hidden lg:flex relative group" data-dropdown>
                <button
                  data-dropdown-button
                  className="hidden lg:flex text-sm font-medium text-gray-300 hover:text-green-400 transition-all duration-300 items-center space-x-1 px-3 py-2 rounded-lg hover:bg-green-500/10 hover:shadow-lg hover:shadow-green-500/20 hover:scale-100 transform nav-link-hover"
                >
                  <svg
                    className="w-4 h-4 mr-2"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M12 0C5.374 0 0 5.373 0 12 0 17.302 3.438 21.8 8.207 23.387c.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.300 24 12c0-6.627-5.373-12-12-12z" />
                  </svg>
                  <svg
                    className="w-4 h-4 ml-1"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                <div
                  data-dropdown-menu
                  className="absolute hidden lg:flex right-0 mt-1 w-48 bg-gray-800/95 backdrop-blur-sm rounded-md shadow-lg border border-gray-700 before:content-[''] before:absolute before:bottom-full before:left-0 before:right-0 before:h-2 before:bg-transparent"
                  style={{ display: "none" }}
                >
                  <a
                    href="https://github.com/kubestellar/kubestellar"
                    className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
                  >
                    <svg
                      className="w-5 h-5 mr-2"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"></path>
                    </svg>
                    {t("githubStar")}
                    <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                      {githubStats.stars}
                    </span>
                  </a>
                  <a
                    href="https://github.com/kubestellar/kubestellar/fork"
                    className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-2"
                      fill="currentColor"
                      viewBox="0 0 16 16"
                    >
                      <path d="M5 5.372v.878c0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75v-.878a2.25 2.25 0 1 1 1.5 0v.878a2.25 2.25 0 0 1-2.25 2.25h-1.5v2.128a2.251 2.251 0 1 1-1.5 0V8.5h-1.5A2.25 2.25 0 0 1 3.5 6.25v-.878a2.25 2.25 0 1 1 1.5 0ZM5 3.25a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Zm6.75.75a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm-3 8.75a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Z" />
                    </svg>

                    {t("githubFork")}
                    <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                      {githubStats.forks}
                    </span>
                  </a>
                  <a
                    href="https://github.com/kubestellar/kubestellar/watchers"
                    className="flex items-center px-4 py-2 text-sm text-gray-300 hover:bg-emerald-900/30 rounded transition-all duration-200 hover:text-emerald-300 hover:shadow-md"
                  >
                    <svg
                      className="w-4 h-4 mr-2"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path d="M10 2C5.454 2 1.73 5.11.458 9.09a1.5 1.5 0 000 1.82C1.73 14.89 5.454 18 10 18s8.27-3.11 9.542-7.09a1.5 1.5 0 000-1.82C18.27 5.11 14.546 2 10 2zm0 14c-3.866 0-7.09-2.61-8.13-6C2.91 6.61 6.134 4 10 4s7.09 2.61 8.13 6c-1.04 3.39-4.264 6-8.13 6zm0-8a2 2 0 110 4 2 2 0 010-4z" />
                    </svg>
                    {t("githubWatch")}
                    <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                      {githubStats.watchers}
                    </span>
                  </a>
                </div>
              </div>

              {/* Mobile menu button */}
              <button
                className="lg:hidden p-2 rounded focus:outline-none hover:bg-gray-100 dark:hover:bg-gray-700 group cursor-pointer"
                aria-label={isMenuOpen ? "Close menu" : "Open menu"}
                onClick={() => setIsMenuOpen(!isMenuOpen)}
              >
                {isMenuOpen ? (
                  <svg
                    className="w-6 h-6 stroke-black dark:stroke-white transition-all duration-300 group-hover:scale-110 group-hover:rotate-90"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                ) : (
                  <svg
                    className="w-6 h-6 stroke-black dark:stroke-white transition-all duration-300 group-hover:scale-110"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 6h16M4 12h16M4 18h16"
                    />
                  </svg>
                )}
              </button>
            </div>
          </div>

          {/* Mobile menu */}
          {isMenuOpen && (
            <div className="lg:hidden max-h-[calc(100vh-4rem)] overflow-y-auto">
              <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3">
                {/*DOCS */}
                <div className="relative mb-2">
                  <Link
                    href="/docs"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    <span>{t("docs")}</span>
                  </Link>
                </div>
                {/*BLOG */}
                <div className="relative mb-4">
                  <Link
                    target="_blank"
                    href="https://kubestellar.io/blog"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    <span>{t("blog")}</span>
                  </Link>
                </div>
                <div className="relative mb-4">
                  <Link
                    href="/playground"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
                          className="group-hover:stroke-[2.5] transition-all duration-300"
                        ></path>
                      </svg>
                    </div>
                    <span>{t("playground")}</span>
                  </Link>
                </div>
                {/* MARKETPLACE */}
                <div className="relative mb-4">
                  <Link
                    href="/marketplace"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"
                        ></path>
                      </svg>
                    </div>
                    <span>{t("marketplace")}</span>
                  </Link>
                </div>
                <div className="border-t border-gray-400/50 mb-4"></div>
                <div className="mb-2">
                  <span className="text-sm sm:text-base px-3 font-medium tracking-wider text-gray-400 uppercase">
                    {t("contribute")}
                  </span>
                </div>
                <div className="relative mb-2">
                  <a
                    href="https://kubestellar.io/joinus"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    </div>
                    <span>{t("joinIn")}</span>
                  </a>
                </div>
                <div className="relative mb-2">
                  <Link
                    href="/contribute-handbook"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    </div>
                    <span>{t("contributeHandbook")}</span>
                  </Link>
                </div>
                <div className="relative mb-2">
                  <Link
                    href="/quick-installation"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <svg
                      className="w-5 h-5 mr-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                      ></path>
                    </svg>
                    {t("quickInstallation")}
                  </Link>
                </div>
                <div className="relative mb-2">
                  <Link
                    href="/products"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <svg
                      className="w-5 h-5 mr-3"
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
                    {t("products")}
                  </Link>
                </div>
                <div className="relative mb-2">
                  <Link
                    href="/ladder"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    </div>
                    <span>{t("ladder")}</span>
                  </Link>
                </div>
                <div className="relative mb-4">
                  <Link
                    href="/docs/contribution-guidelines/security/security-inc"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    </div>
                    <span>{t("security")}</span>
                  </Link>
                </div>
                <div className="border-t border-gray-400/50 mb-4"></div>
                <div className="mb-2">
                  <span className="px-3 text-sm sm:text-base font-medium tracking-wider text-gray-400 uppercase">
                    {t("community")}
                  </span>
                </div>
                <div className="relative mb-2">
                  <Link
                    href="/docs"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    </div>
                    <span>{t("getInvolved")}</span>
                  </Link>
                </div>
                <div className="relative mb-2">
                  <Link
                    href="/programs"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
                        ></path>
                      </svg>
                    </div>
                    <span>{t("programs")}</span>
                  </Link>
                </div>
                <div className="relative mb-2">
                  <a
                    href="#contact-us"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    </div>
                    <span>{t("contactUs")}</span>
                  </a>
                </div>
                <div className="relative mb-4">
                  <Link
                    href="/partners"
                    className="text-sm sm:text-base font-medium text-gray-300 flex items-center space-x-1 px-3 py-2 rounded-lg"
                  >
                    <div className="relative">
                      <svg
                        className="w-5 h-5 mr-3"
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
                    <span>{t("partners")}</span>
                  </Link>
                </div>
                <div className="border-t border-gray-400/50 mb-4"></div>
                <div className="mb-4">
                  <span className="text-sm sm:text-base px-3 font-medium tracking-wider text-gray-400 uppercase">
                    {t("github")}
                  </span>
                </div>

                {/* GitHub Stats */}
                <div className="mt-2 mb-4 px-3 space-y-2">
                  {/* Stars */}
                  <a
                    href="https://github.com/kubestellar/kubestellar"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center justify-between px-3 py-2 rounded-lg  max-w-xs"
                  >
                    <div className="flex items-center space-x-3">
                      <svg
                        className="w-5 h-5 text-gray-300"
                        fill="currentColor"
                        viewBox="0 0 20 20"
                      >
                        <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"></path>
                      </svg>
                      <span className="text-sm font-medium text-gray-300">
                        {t("githubStar")}
                      </span>
                    </div>
                    <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                      {githubStats.stars}
                    </span>
                  </a>

                  {/* Forks */}
                  <a
                    href="https://github.com/kubestellar/kubestellar/fork"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center justify-between px-3 py-2 rounded-lg  max-w-xs"
                  >
                    <div className="flex items-center space-x-3">
                      <svg
                        className="w-5 h-5 mr-3 text-gray-300"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"
                        ></path>
                      </svg>
                      <span className="text-sm font-medium text-gray-300">
                        {t("githubFork")}
                      </span>
                    </div>
                    <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                      {githubStats.forks}
                    </span>
                  </a>

                  {/* Watchers */}
                  <a
                    href="https://github.com/kubestellar/kubestellar/watchers"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center justify-between px-3 py-2 rounded-lg  max-w-xs"
                  >
                    <div className="flex items-center space-x-3">
                      <svg
                        className="w-5 h-5 mr-3 text-gray-300"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                        ></path>
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                        ></path>
                      </svg>
                      <span className="text-sm font-medium text-gray-300">
                        {t("githubWatch")}
                      </span>
                    </div>
                    <span className="ml-auto bg-gray-700 text-gray-300 text-xs rounded px-2 py-0.5">
                      {githubStats.watchers}
                    </span>
                  </a>
                </div>
              </div>
            </div>
          )}
        </div>
      </nav>
    </>
  );
}
