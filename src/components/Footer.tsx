"use client";

import { useEffect } from "react";
import Image from "next/image";
import { GridLines, StarField } from "./index";
import Link from "next/link";

export default function Footer() {
  useEffect(() => {
    // Back to top functionality
    const initBackToTop = () => {
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

      window.addEventListener("scroll", toggleButton);

      backToTopButton.addEventListener("click", () => {
        window.scrollTo({
          top: 0,
          behavior: "smooth",
        });
      });

      // Initial check
      toggleButton();
    };

    initBackToTop();
  }, []);

  return (
    <footer className="bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white relative overflow-hidden pt-16 pb-8">
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]"></div>

      {/* Starfield background */}
      <StarField density="low" showComets={true} cometCount={2} />

      {/* Grid lines background */}
      <GridLines horizontalLines={21} verticalLines={15} />

      {/* Background elements */}
      <div className="absolute inset-0 z-0">
        <div className="absolute bottom-0 left-0 w-full h-1/2 bg-gradient-to-t from-blue-900/10 to-transparent"></div>
        <div className="absolute top-0 right-0 w-full h-1/2 bg-gradient-to-b from-purple-900/10 to-transparent"></div>
      </div>

      <div className="relative z-10 max-w-7xl mx-auto px-8 sm:px-6 lg:px-8">
        {/* Main footer content */}
        <div className="grid grid-cols-1 min-[450px]:grid-cols-2 md:grid-cols-12 gap-8 mb-12">
          {/* Brand Section */}
          <div className="min-[450px]:col-span-2 md:col-span-4">
            <div className="flex items-center-space-x-2 mb-2  ml-[-7px]">
              <Image
                src="/KubeStellar-with-Logo-transparent.png"
                alt="Kubestellar logo"
                width={160}
                height={40}
                className="h-10 w-auto"
              />
            </div>
            <p className="text-gray-300 mb-6 leading-relaxed">
              Multi-Cluster Kubernetes orchestration platform that simplifies
              distributed workload management across diverse infrastructure.
            </p>
            <div className="flex space-x-4">
                <Link
                target="_blank"
                href="#"
                className="w-12 h-12 rounded-lg flex items-center justify-center transition-colors duration-300"
                >
                <svg
                  opacity={0.5}
                  className="w-6 h-6 hover:opacity-100"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
                </svg>
                </Link>
              <Link
                target="_blank"
                href="https://github.com/kubestellar"
                className="group relative w-12 h-12 rounded-lg flex items-center justify-center transition-all duration-300"
              >
                <svg
                  className="w-6 h-6 hover:opacity-100"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    className="transition-colors duration-300 text-gray-400 group-hover:text-white"
                    d="M12 0C5.374 0 0 5.373 0 12 0 17.302 3.438 21.8 8.207 23.387c.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.300 24 12c0-6.627-5.373-12-12-12z"
                  />
                </svg>
              </Link>
              <Link
                target="_blank"
                href="https://www.linkedin.com/company/kubestellar/"
                className="group relative w-12 h-12 rounded-lg flex items-center justify-center transition-all duration-300"
              >
                <svg
                  className="w-6 h-6 hover:opacity-100"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    className="transition-colors duration-300 text-gray-400 group-hover:text-[#0077B5]"
                    d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"
                  />
                </svg>
              </Link>
              <Link
                target="_blank"
                href="https://www.youtube.com/@kubestellar"
                className="group relative w-12 h-12 rounded-lg flex items-center justify-center transition-all duration-300"
              >
                <svg
                  className="w-6 h-6 hover:opacity-100"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    className="transition-colors duration-300 text-gray-400 group-hover:text-[#FF0000]"
                    d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"
                  />
                </svg>
              </Link>
            </div>
          </div>

          {/* Column 2 on iPad: Docs + Contribute */}
          <div className="space-y-8 md:col-span-4 md:space-y-8 lg:space-y-0 lg:grid lg:grid-cols-2 lg:gap-8 lg:contents">
            {/* Docs Links */}
            <div className="lg:col-span-2">
              <h3 className="text-lg font-semibold text-white mb-4">Docs</h3>
              <ul className="space-y-3">
                <li>
                  <a
                    href="https://docs.kubestellar.io/release-0.28.0/readme/"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Overview
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Install & Configure
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Uses & Integrate
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    User Guide & Support
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    UI Tools
                  </a>
                </li>
              </ul>
            </div>

            {/* Contribute Links */}
            <div className="lg:col-span-2">
              <h3 className="text-lg font-semibold text-white mb-4">
                Contribute
              </h3>
              <ul className="space-y-3">
                <li>
                  <a
                    href="https://docs.kubestellar.io/release-0.28.0/direct/contribute/"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Overview
                  </a>
                </li>
                <li>
                  <a
                    href="https://docs.kubestellar.io/release-0.28.0/contribution-guidelines/coc-inc/"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Code of Conduct
                  </a>
                </li>
                <li>
                  <a
                    href="https://docs.kubestellar.io/release-0.28.0/contribution-guidelines/contributing-inc/"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Guidelines
                  </a>
                </li>
                <li>
                  <a
                    href="https://docs.kubestellar.io/release-0.28.0/contribution-guidelines/license-inc/"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    License
                  </a>
                </li>
                <li>
                  <a
                    href="https://docs.kubestellar.io/release-0.28.0/contribution-guidelines/onboarding-inc/"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Onboarding
                  </a>
                </li>
              </ul>
            </div>
          </div>

          {/* Column 3 on iPad: Community + Stay Updated */}
          <div className="space-y-8 md:col-span-4 md:space-y-8 lg:space-y-0 lg:grid lg:grid-cols-2 lg:gap-8 lg:contents">
            {/* Community Links */}
            <div className="lg:col-span-2">
              <h3 className="text-lg font-semibold text-white mb-4">
                Community
              </h3>
              <ul className="space-y-3">
                <li>
                  <a
                    href="http://docs.kubestellar.io/release-0.28.0/Community/_index/"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Get Involved
                  </a>
                </li>
                <li>
                  <a
                    href=""
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Programs
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Ladder
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-gray-400 hover:text-white transition-colors duration-200 text-sm"
                  >
                    Partners
                  </a>
                </li>
              </ul>
            </div>

            {/* Stay updated */}
            <div className="lg:col-span-2">
              <h3 className="text-sm font-semibold text-white uppercase tracking-wider mb-4">
                Stay Updated
              </h3>
              <div className="bg-gray-800/50 backdrop-blur-md rounded-lg p-4 border border-gray-700/50 transform transition-all duration-300 hover:border-blue-500/30">
                <form id="newsletter-form" className="flex flex-col space-y-3">
                  <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-5 w-5 text-gray-400"
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
                      className="block w-full pl-10 pr-3 py-2 text-sm text-white placeholder-gray-400 bg-gray-700/50 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200"
                      placeholder="Email"
                    />
                  </div>
                  <button
                    type="submit"
                    className="w-full px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-blue-600 to-purple-600 border border-transparent rounded-md shadow-sm hover:from-blue-700 hover:to-purple-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900 transition-all duration-200 transform hover:translate-y-[-1px] flex items-center justify-center"
                  >
                    <span>Subscribe</span>
                  </button>
                </form>

                {/* Success message (hidden by default) */}
                <div
                  id="newsletter-success"
                  className="hidden mt-3 text-sm text-green-400 flex items-center"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-4 w-4 mr-1"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>Subscribed!</span>
                </div>
              </div>
              <p className="mt-3 text-xs text-gray-400">
                We respect your privacy. No spam.
              </p>
            </div>
          </div>
        </div>

        {/* Divider and bottom section */}
        <div className="border-t border-gray-800 pt-8">
          <div className="flex flex-col md:flex-row justify-between items-center">
            <div><p className="text-gray-400">
                © 2025 KubeStellar. All rights reserved.
              </p></div>
            <div className="flex items-center space-x-6 mb-4 md:mb-0">
              <div className="flex items-center space-x-4">
                <a
                  href="#"
                  className="text-gray-400 hover:text-white transition-colors duration-300 text-sm"
                >
                  Privacy Policy
                </a>
                <span className="text-gray-600">•</span>
                <a
                  href="#"
                  className="text-gray-400 hover:text-white transition-colors duration-300 text-sm"
                >
                  Terms of Service
                </a>
                <span className="text-gray-600">•</span>
                <a
                  href="#"
                  className="text-gray-400 hover:text-white transition-colors duration-300 text-sm"
                >
                  Cookie Policy
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Floating back to top button */}
      <button
        id="back-to-top"
        className="fixed bottom-8 right-8 p-3 rounded-full bg-blue-600 text-white shadow-lg z-50 transition-all duration-300 opacity-0 translate-y-10 hover:bg-blue-700 hover:scale-110"
        aria-label="Back to top"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
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