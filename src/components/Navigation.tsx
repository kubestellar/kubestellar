"use client";

import { useState, useEffect } from "react";

export default function Navigation() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  useEffect(() => {
    // Initialize dropdowns functionality
    const initDropdowns = () => {
      const dropdownContainers = document.querySelectorAll("[data-dropdown]");

      dropdownContainers.forEach(container => {
        const button = container.querySelector("[data-dropdown-button]");
        const menu = container.querySelector(
          "[data-dropdown-menu]"
        ) as HTMLElement;

        if (button && menu) {
          button.addEventListener("click", e => {
            e.preventDefault();
            e.stopPropagation();

            // Close other dropdowns
            dropdownContainers.forEach(otherContainer => {
              if (otherContainer !== container) {
                const otherMenu = otherContainer.querySelector(
                  "[data-dropdown-menu]"
                ) as HTMLElement;
                if (otherMenu) {
                  otherMenu.style.display = "none";
                }
              }
            });

            // Toggle current dropdown
            if (menu.style.display === "block") {
              menu.style.display = "none";
            } else {
              menu.style.display = "block";
            }
          });
        }
      });

      // Close dropdowns when clicking outside
      document.addEventListener("click", e => {
        dropdownContainers.forEach(container => {
          if (!container.contains(e.target as Node)) {
            const menu = container.querySelector(
              "[data-dropdown-menu]"
            ) as HTMLElement;
            if (menu) {
              menu.style.display = "none";
            }
          }
        });
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

    // Create starfield and grid for navigation
    const createStarfield = (container: HTMLElement) => {
      if (!container) return;
      container.innerHTML = "";

      for (let layer = 1; layer <= 3; layer++) {
        const layerDiv = document.createElement("div");
        layerDiv.className = `star-layer layer-${layer}`;
        layerDiv.style.position = "absolute";
        layerDiv.style.inset = "0";
        layerDiv.style.zIndex = layer.toString();

        const starCount = layer === 1 ? 50 : layer === 2 ? 30 : 20;

        for (let i = 0; i < starCount; i++) {
          const star = document.createElement("div");
          star.style.position = "absolute";
          star.style.width = `${Math.random() * 2 + 1}px`;
          star.style.height = star.style.width;
          star.style.backgroundColor = "white";
          star.style.borderRadius = "50%";
          star.style.top = `${Math.random() * 100}%`;
          star.style.left = `${Math.random() * 100}%`;
          star.style.opacity = Math.random().toString();
          star.style.animation = `twinkle ${Math.random() * 3 + 2}s infinite alternate`;
          star.style.animationDelay = `${Math.random() * 2}s`;
          layerDiv.appendChild(star);
        }

        container.appendChild(layerDiv);
      }
    };

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

    const starsContainer = document.getElementById("stars-container-nav");
    const gridContainer = document.getElementById("grid-lines-nav");

    if (starsContainer) createStarfield(starsContainer);
    if (gridContainer) createGrid(gridContainer);

    initDropdowns();
  }, []);

  return (
    <nav className="fixed w-full z-50 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900/90 backdrop-blur-md border-b border-gray-700/50 transition-all duration-300">
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]/90"></div>

      {/* Starfield background */}
      <div
        id="stars-container-nav"
        className="absolute inset-0 overflow-hidden"
      ></div>

      {/* Grid lines background */}
      <div id="grid-lines-nav" className="absolute inset-0 opacity-10"></div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="flex justify-between h-16 items-center">
          {/* Left side: Logo */}
          <div className="flex-shrink-0">
            <a href="#" className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">K</span>
              </div>
              <span className="text-xl font-bold text-white">KubeStellar</span>
            </a>
          </div>

          {/* Center: Nav Links */}
          <div className="hidden md:flex flex-1 justify-center">
            <div className="flex items-center space-x-8">
              <a
                href="#about"
                className="nav-link-hover px-3 py-2 rounded-md text-sm font-medium text-gray-300 hover:text-white transition-colors"
              >
                About
              </a>
              <a
                href="#how-it-works"
                className="nav-link-hover px-3 py-2 rounded-md text-sm font-medium text-gray-300 hover:text-white transition-colors"
              >
                How It Works
              </a>
              <a
                href="#use-cases"
                className="nav-link-hover px-3 py-2 rounded-md text-sm font-medium text-gray-300 hover:text-white transition-colors"
              >
                Use Cases
              </a>
              <a
                href="#get-started"
                className="nav-link-hover px-3 py-2 rounded-md text-sm font-medium text-gray-300 hover:text-white transition-colors"
              >
                Get Started
              </a>
              <a
                href="#contact"
                className="nav-link-hover px-3 py-2 rounded-md text-sm font-medium text-gray-300 hover:text-white transition-colors"
              >
                Contact
              </a>
            </div>
          </div>

          {/* Right side: Controls */}
          <div className="flex items-center space-x-4">
            {/* Version Dropdown */}
            <div className="relative group" data-dropdown>
              <button
                data-dropdown-button
                className="nav-link-hover flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-300 hover:text-white transition-colors"
              >
                <svg
                  className="w-4 h-4 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
                  />
                </svg>
                v0.1.0
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
                className="absolute right-0 mt-2 w-48 bg-gray-800/95 backdrop-blur-sm rounded-md shadow-lg border border-gray-700"
              >
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 first:rounded-t-md"
                >
                  v0.1.0 (current)
                </a>
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700"
                >
                  v0.0.9
                </a>
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 last:rounded-b-md"
                >
                  v0.0.8
                </a>
              </div>
            </div>

            {/* Language Dropdown */}
            <div className="relative group" data-dropdown>
              <button
                data-dropdown-button
                className="nav-link-hover flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-300 hover:text-white transition-colors"
              >
                <svg
                  className="w-4 h-4 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129"
                  />
                </svg>
                EN
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
                className="absolute right-0 mt-2 w-32 bg-gray-800/95 backdrop-blur-sm rounded-md shadow-lg border border-gray-700"
              >
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 first:rounded-t-md"
                >
                  English
                </a>
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 last:rounded-b-md"
                >
                  中文
                </a>
              </div>
            </div>

            {/* GitHub Dropdown */}
            <div className="relative group" data-dropdown>
              <button
                data-dropdown-button
                className="nav-link-hover flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-300 hover:text-white transition-colors"
              >
                <svg
                  className="w-4 h-4 mr-2"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path d="M12 0C5.374 0 0 5.373 0 12 0 17.302 3.438 21.8 8.207 23.387c.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.300 24 12c0-6.627-5.373-12-12-12z" />
                </svg>
                GitHub
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
                className="absolute right-0 mt-2 w-48 bg-gray-800/95 backdrop-blur-sm rounded-md shadow-lg border border-gray-700"
              >
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 first:rounded-t-md"
                >
                  Repository
                </a>
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700"
                >
                  Issues
                </a>
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700"
                >
                  Releases
                </a>
                <a
                  href="#"
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 last:rounded-b-md"
                >
                  Contributing
                </a>
              </div>
            </div>

            {/* Mobile menu button */}
            <button
              className="md:hidden p-2 rounded focus:outline-none hover:bg-gray-100 dark:hover:bg-gray-700"
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
          <div className="md:hidden">
            <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3">
              <a
                href="#about"
                className="block px-3 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
              >
                About
              </a>
              <a
                href="#how-it-works"
                className="block px-3 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
              >
                How It Works
              </a>
              <a
                href="#use-cases"
                className="block px-3 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
              >
                Use Cases
              </a>
              <a
                href="#get-started"
                className="block px-3 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
              >
                Get Started
              </a>
              <a
                href="#contact"
                className="block px-3 py-2 rounded-md text-base font-medium text-gray-300 hover:text-white hover:bg-gray-700"
              >
                Contact
              </a>
            </div>
          </div>
        )}
      </div>
    </nav>
  );
}
