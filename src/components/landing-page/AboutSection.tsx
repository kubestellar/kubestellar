"use client";

import { useEffect } from "react";

export default function AboutSection() {
  useEffect(() => {
    // Create starfield and grid for About section
    const createStarfield = (container: HTMLElement) => {
      if (!container) return;
      container.innerHTML = "";

      for (let layer = 1; layer <= 3; layer++) {
        const layerDiv = document.createElement("div");
        layerDiv.className = `star-layer layer-${layer}`;
        layerDiv.style.position = "absolute";
        layerDiv.style.inset = "0";
        layerDiv.style.zIndex = layer.toString();

        const starCount = layer === 1 ? 80 : layer === 2 ? 50 : 30;

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

      for (let i = 0; i < 15; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", "0");
        line.setAttribute("y1", `${i * 7}%`);
        line.setAttribute("x2", "100%");
        line.setAttribute("y2", `${i * 7}%`);
        line.setAttribute("stroke", "#6366F1");
        line.setAttribute("stroke-width", "0.5");
        line.setAttribute("stroke-opacity", "0.3");
        gridSvg.appendChild(line);
      }

      for (let i = 0; i < 15; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", `${i * 7}%`);
        line.setAttribute("y1", "0");
        line.setAttribute("x2", `${i * 7}%`);
        line.setAttribute("y2", "100%");
        line.setAttribute("stroke", "#6366F1");
        line.setAttribute("stroke-width", "0.5");
        line.setAttribute("stroke-opacity", "0.3");
        gridSvg.appendChild(line);
      }

      container.appendChild(gridSvg);
    };

    // Feature cards animation
    const initFeatureCards = () => {
      const featureCards = document.querySelectorAll(".feature-card");
      featureCards.forEach(card => {
        const cardElement = card as HTMLElement;

        // Add 3D hover effect
        cardElement.addEventListener("mouseenter", () => {
          cardElement.style.transform =
            "perspective(1000px) rotateY(5deg) rotateX(5deg) translateZ(20px)";
        });

        cardElement.addEventListener("mouseleave", () => {
          cardElement.style.transform =
            "perspective(1000px) rotateY(0deg) rotateX(0deg) translateZ(0px)";
        });

        // Add parallax effect on mouse move
        cardElement.addEventListener("mousemove", e => {
          const rect = cardElement.getBoundingClientRect();
          const x = e.clientX - rect.left;
          const y = e.clientY - rect.top;

          const centerX = rect.width / 2;
          const centerY = rect.height / 2;

          const rotateY = (x - centerX) / 10;
          const rotateX = (centerY - y) / 10;

          cardElement.style.transform = `perspective(1000px) rotateY(${rotateY}deg) rotateX(${rotateX}deg) translateZ(20px)`;
        });
      });
    };

    const starsContainer = document.getElementById("stars-container-about");
    const gridContainer = document.getElementById("grid-lines-about");

    if (starsContainer) createStarfield(starsContainer);
    if (gridContainer) createGrid(gridContainer);

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
      <div
        id="stars-container-about"
        className="absolute inset-0 overflow-hidden"
      ></div>

      {/* Grid lines background */}
      <div id="grid-lines-about" className="absolute inset-0 opacity-20"></div>

      {/* Background decorative elements */}
      <div className="absolute -top-24 -right-24 w-96 h-96 bg-gradient-to-br from-blue-500/10 to-purple-500/10 rounded-full blur-3xl"></div>
      <div className="absolute -bottom-24 -left-24 w-96 h-96 bg-gradient-to-tr from-purple-500/10 to-blue-500/10 rounded-full blur-3xl"></div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="text-center">
          <h2 className="text-3xl font-extrabold text-white sm:text-4xl">
            About <span className="text-gradient">KubeStellar</span>
          </h2>
          <p className="mt-4 max-w-2xl mx-auto text-xl text-gray-300">
            Transform your multi-cluster Kubernetes operations with intelligent
            automation, seamless scalability, and enterprise-grade reliability.
          </p>
        </div>

        <div className="mt-20 grid gap-10 lg:grid-cols-3 feature-cards">
          {/* Feature 1 - Advanced card with 3D hover effect */}
          <div className="feature-card relative group perspective bg-gradient-to-br from-gray-800/50 to-gray-900/50 backdrop-blur-sm rounded-xl p-8 border border-gray-700/50 transition-all duration-300">
            <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-purple-500/10 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div className="relative z-10">
              <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center mb-6">
                <svg
                  className="w-6 h-6 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 11H5m14-7H5m14 14H5"
                  />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-white mb-4">
                Unified Management
              </h3>
              <p className="text-gray-300 leading-relaxed">
                Manage multiple Kubernetes clusters from a single control plane.
                Deploy, monitor, and maintain workloads across hybrid and
                multi-cloud environments with ease.
              </p>
              <div className="mt-6 flex items-center text-blue-400 text-sm font-medium">
                <span>Learn more</span>
                <svg
                  className="w-4 h-4 ml-2 transform group-hover:translate-x-1 transition-transform"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>
          </div>

          {/* Feature 2 - Advanced card with 3D hover effect */}
          <div className="feature-card relative group perspective bg-gradient-to-br from-gray-800/50 to-gray-900/50 backdrop-blur-sm rounded-xl p-8 border border-gray-700/50 transition-all duration-300">
            <div className="absolute inset-0 bg-gradient-to-br from-purple-500/10 to-pink-500/10 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div className="relative z-10">
              <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center mb-6">
                <svg
                  className="w-6 h-6 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-white mb-4">
                Intelligent Distribution
              </h3>
              <p className="text-gray-300 leading-relaxed">
                Automatically distribute workloads based on resource
                availability, geographic constraints, and business policies.
                Optimize performance and reduce costs with smart placement.
              </p>
              <div className="mt-6 flex items-center text-purple-400 text-sm font-medium">
                <span>Learn more</span>
                <svg
                  className="w-4 h-4 ml-2 transform group-hover:translate-x-1 transition-transform"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>
          </div>

          {/* Feature 3 - Advanced card with 3D hover effect */}
          <div className="feature-card relative group perspective bg-gradient-to-br from-gray-800/50 to-gray-900/50 backdrop-blur-sm rounded-xl p-8 border border-gray-700/50 transition-all duration-300">
            <div className="absolute inset-0 bg-gradient-to-br from-green-500/10 to-teal-500/10 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div className="relative z-10">
              <div className="w-12 h-12 bg-gradient-to-br from-green-500 to-green-600 rounded-lg flex items-center justify-center mb-6">
                <svg
                  className="w-6 h-6 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-white mb-4">
                Enterprise Security
              </h3>
              <p className="text-gray-300 leading-relaxed">
                Built-in security controls, policy enforcement, and compliance
                monitoring. Ensure your multi-cluster deployments meet
                enterprise security standards.
              </p>
              <div className="mt-6 flex items-center text-green-400 text-sm font-medium">
                <span>Learn more</span>
                <svg
                  className="w-4 h-4 ml-2 transform group-hover:translate-x-1 transition-transform"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
