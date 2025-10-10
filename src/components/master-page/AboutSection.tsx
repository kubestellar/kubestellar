"use client";

import { useEffect } from "react";
import { GridLines, StarField } from "../index";

export default function AboutSection() {
  useEffect(() => {
    // Feature cards animation
    const initFeatureCards = () => {
      const featureCards = document.querySelectorAll(".feature-card");

      // Cards appear on scroll
      const observer = new IntersectionObserver(
        entries => {
          entries.forEach((entry, index) => {
            if (entry.isIntersecting) {
              setTimeout(() => {
                entry.target.classList.add("animate-in");
              }, index * 150);
              observer.unobserve(entry.target);
            }
          });
        },
        {
          threshold: 0.2,
        }
      );

      featureCards.forEach(card => {
        card.classList.add("opacity-0", "translate-y-10");
        observer.observe(card);
      });

      // Add CSS to handle animation
      const style = document.createElement("style");
      style.textContent = `
        .feature-card {
          transition: opacity 0.6s ease-out, transform 0.6s ease-out;
        }
        .feature-card.animate-in {
          opacity: 1 !important;
          transform: translateY(0) !important;
        }
        .perspective {
          perspective: 1000px;
        }
        .transform-style-3d {
          transform-style: preserve-3d;
        }
        .rotate-y-10 {
          transform: rotateY(10deg);
        }
      `;
      document.head.appendChild(style);

      // 3D tilt effect on mouse move
      featureCards.forEach(card => {
        card.addEventListener("mousemove", (e: Event) => {
          const mouseEvent = e as MouseEvent;
          const container = card.querySelector(".card-3d-container");
          const rect = card.getBoundingClientRect();
          const x = mouseEvent.clientX - rect.left;
          const y = mouseEvent.clientY - rect.top;

          const centerX = rect.width / 2;
          const centerY = rect.height / 2;

          // Calculate rotation values (reduced intensity for subtlety)
          const rotateY = (x - centerX) / 15;
          const rotateX = (centerY - y) / 15;

          if (container) {
            (container as HTMLElement).style.transform =
              `rotateY(${rotateY}deg) rotateX(${rotateX}deg)`;
          }
        });

        // Reset on mouse leave
        card.addEventListener("mouseleave", () => {
          const container = card.querySelector(".card-3d-container");
          if (container) {
            (container as HTMLElement).style.transform =
              "rotateY(0deg) rotateX(0deg)";
          }
        });
      });
    };

    initFeatureCards();
  }, []);

  return (
    <section
      id="about"
      className="relative py-24 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden"
    >
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]"></div>

      {/* Starfield background */}
      <StarField density="medium" showComets={true} cometCount={3} />

      {/* Grid lines background */}
      <GridLines horizontalLines={21} verticalLines={15} />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="text-center">
          <h2 className="text-3xl font-extrabold text-white sm:text-4xl">
            What is{" "}
            <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">
              KubeStellar
            </span>
          </h2>
          <p className="mt-4 max-w-2xl mx-auto text-xl text-gray-300">
            A multi-cluster Kubernetes orchestration platform that simplifies
            how you manage distributed workloads.
          </p>
        </div>

        <div className="mt-20 grid gap-10 lg:grid-cols-3 feature-cards">
          {/* Feature 1 - Advanced card with 3D hover effect */}
          <div className="feature-card relative group perspective">
            <div className="card-3d-container relative transition-all duration-500 group-hover:rotate-y-10 w-full h-full transform-style-3d">
              <div className="absolute -inset-0.5 bg-gradient-to-r from-primary-600 to-purple-600 rounded-xl blur opacity-30 group-hover:opacity-90 transition duration-500"></div>
              <div className="relative bg-gray-800/50 backdrop-blur-md rounded-xl shadow-lg p-8 transition-all duration-300 transform group-hover:translate-y-[-8px] group-hover:shadow-xl border border-gray-700/50 h-full">
                {/* Icon with animation */}
                <div className="w-16 h-16 rounded-full bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-300">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-8 w-8 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M5 12h14M12 5l7 7-7 7"
                    />
                  </svg>
                </div>

                <h3 className="text-2xl font-bold text-white mb-3 group-hover:text-blue-400 transition-colors duration-300">
                  Single Control Plane
                </h3>

                <p className="text-gray-300 mb-6">
                  Manage multiple Kubernetes clusters from a unified control
                  plane, eliminating the need to switch contexts and
                  streamlining operations.
                </p>

                {/* Animated arrow on hover */}
                <div className="h-8 overflow-hidden">
                  <div className="transform translate-y-8 group-hover:translate-y-0 transition-transform duration-300 text-primary-600 dark:text-primary-400 flex items-center">
                    <span className="text-sm font-medium">Learn more</span>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="ml-1 h-4 w-4 transform group-hover:translate-x-1 transition-transform duration-300"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M10.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L12.586 11H5a1 1 0 110-2h7.586l-2.293-2.293a1 1 0 010-1.414z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Feature 2 - Advanced card with 3D hover effect */}
          <div className="feature-card relative group perspective">
            <div className="card-3d-container relative transition-all duration-500 group-hover:rotate-y-10 w-full h-full transform-style-3d">
              <div className="absolute -inset-0.5 bg-gradient-to-r from-primary-600 to-purple-600 rounded-xl blur opacity-30 group-hover:opacity-90 transition duration-500"></div>
              <div className="relative bg-gray-800/50 backdrop-blur-md rounded-xl shadow-lg p-8 transition-all duration-300 transform group-hover:translate-y-[-8px] group-hover:shadow-xl border border-gray-700/50">
                {/* Icon with animation */}
                <div className="w-16 h-16 rounded-full bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-300">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-8 w-8 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M13 10V3L4 14h7v7l9-11h-7z"
                    />
                  </svg>
                </div>

                <h3 className="text-2xl font-bold text-white mb-3 group-hover:text-blue-400 transition-colors duration-300">
                  Intelligent Workload Placement
                </h3>

                <p className="text-gray-300 mb-6">
                  Automatically distribute workloads based on cluster
                  capabilities, availability, and custom constraints for optimal
                  performance.
                </p>

                {/* Animated arrow on hover */}
                <div className="h-8 overflow-hidden">
                  <div className="transform translate-y-8 group-hover:translate-y-0 transition-transform duration-300 text-primary-600 dark:text-primary-400 flex items-center">
                    <span className="text-sm font-medium">Learn more</span>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="ml-1 h-4 w-4 transform group-hover:translate-x-1 transition-transform duration-300"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M10.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L12.586 11H5a1 1 0 110-2h7.586l-2.293-2.293a1 1 0 010-1.414z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Feature 3 - Advanced card with 3D hover effect */}
          <div className="feature-card relative group perspective">
            <div className="card-3d-container relative transition-all duration-500 group-hover:rotate-y-10 w-full h-full transform-style-3d">
              <div className="absolute -inset-0.5 bg-gradient-to-r from-primary-600 to-purple-600 rounded-xl blur opacity-30 group-hover:opacity-90 transition duration-500"></div>
              <div className="relative bg-gray-800/50 backdrop-blur-md rounded-xl shadow-lg p-8 transition-all duration-300 transform group-hover:translate-y-[-8px] group-hover:shadow-xl border border-gray-700/50">
                {/* Icon with animation */}
                <div className="w-16 h-16 rounded-full bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-300">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-8 w-8 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                    />
                  </svg>
                </div>

                <h3 className="text-2xl font-bold text-white mb-3 group-hover:text-blue-400 transition-colors duration-300">
                  Policy-Driven Management
                </h3>

                <p className="text-gray-300 mb-6">
                  Define policies for governance, security, and compliance
                  across your entire Kubernetes estate with centralized
                  enforcement.
                </p>

                {/* Animated arrow on hover */}
                <div className="h-8 overflow-hidden">
                  <div className="transform translate-y-8 group-hover:translate-y-0 transition-transform duration-300 text-primary-600 dark:text-primary-400 flex items-center">
                    <span className="text-sm font-medium">Learn more</span>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="ml-1 h-4 w-4 transform group-hover:translate-x-1 transition-transform duration-300"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M10.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L12.586 11H5a1 1 0 110-2h7.586l-2.293-2.293a1 1 0 010-1.414z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
