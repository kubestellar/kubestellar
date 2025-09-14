"use client";

import { useEffect } from "react";

export default function HowItWorksSection() {
  useEffect(() => {
    // Create starfield and grid for How It Works section
    const createStarfield = (container: HTMLElement) => {
      if (!container) return;
      container.innerHTML = "";

      for (let layer = 1; layer <= 3; layer++) {
        const layerDiv = document.createElement("div");
        layerDiv.className = `star-layer layer-${layer}`;
        layerDiv.style.position = "absolute";
        layerDiv.style.inset = "0";
        layerDiv.style.zIndex = layer.toString();

        const starCount = layer === 1 ? 60 : layer === 2 ? 40 : 25;

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

      for (let i = 0; i < 12; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", "0");
        line.setAttribute("y1", `${i * 8}%`);
        line.setAttribute("x2", "100%");
        line.setAttribute("y2", `${i * 8}%`);
        line.setAttribute("stroke", "#6366F1");
        line.setAttribute("stroke-width", "0.5");
        line.setAttribute("stroke-opacity", "0.3");
        gridSvg.appendChild(line);
      }

      for (let i = 0; i < 12; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", `${i * 8}%`);
        line.setAttribute("y1", "0");
        line.setAttribute("x2", `${i * 8}%`);
        line.setAttribute("y2", "100%");
        line.setAttribute("stroke", "#6366F1");
        line.setAttribute("stroke-width", "0.5");
        line.setAttribute("stroke-opacity", "0.3");
        gridSvg.appendChild(line);
      }

      container.appendChild(gridSvg);
    };

    const starsContainer = document.getElementById("stars-container-how");
    const gridContainer = document.getElementById("grid-lines-how");

    if (starsContainer) createStarfield(starsContainer);
    if (gridContainer) createGrid(gridContainer);
  }, []);

  return (
    <section
      id="how-it-works"
      className="relative py-16 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden"
    >
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]"></div>

      {/* Starfield background */}
      <div
        id="stars-container-how"
        className="absolute inset-0 overflow-hidden"
      ></div>

      {/* Grid lines background */}
      <div id="grid-lines-how" className="absolute inset-0 opacity-20"></div>

      <div className="absolute right-0 top-0 h-full w-1/2 bg-gradient-to-l from-blue-500/10 to-transparent"></div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-extrabold text-white sm:text-4xl">
            How <span className="text-gradient">KubeStellar</span> Works
          </h2>
          <p className="mt-4 max-w-2xl mx-auto text-xl text-gray-300">
            Three simple steps to transform your multi-cluster Kubernetes
            operations
          </p>
        </div>

        {/* Interactive Workflow Steps */}
        <div className="relative">
          {/* Connection Line */}
          <div className="absolute left-1/2 top-0 bottom-0 w-0.5 bg-gradient-to-b from-blue-500 to-purple-600 hidden md:block"></div>

          {/* Step 1 */}
          <div className="relative mb-16">
            <div className="flex flex-col md:flex-row items-center">
              <div className="md:w-1/2 md:pr-12">
                <div className="relative">
                  <div className="absolute -left-2 md:-left-6 top-6 w-4 h-4 bg-blue-500 rounded-full border-4 border-gray-900"></div>
                  <h3 className="text-2xl font-bold text-white mb-4">
                    1. Define Your Workloads
                  </h3>
                  <p className="text-gray-300 mb-6 leading-relaxed">
                    Use standard Kubernetes manifests with KubeStellar
                    annotations to specify placement policies, resource
                    requirements, and distribution preferences.
                  </p>
                  <div className="bg-gray-800/50 backdrop-blur-sm rounded-lg p-4 border border-gray-700/50">
                    <div className="text-xs text-gray-400 mb-2">
                      workload.yaml
                    </div>
                    <pre className="text-sm text-green-400 font-mono">
                      {`apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-app
  annotations:
    kubestellar.io/placement: "region=us-east,tier=prod"`}
                    </pre>
                  </div>
                </div>
              </div>
              <div className="flex justify-center md:w-1/2 md:pl-12 mt-8 md:mt-0">
                <div className="relative w-24 h-24 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg">
                  <svg
                    className="w-12 h-12 text-white"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          {/* Step 2 */}
          <div className="relative mb-16">
            <div className="flex flex-col md:flex-row-reverse items-center">
              <div className="md:w-1/2 md:pl-12">
                <div className="relative">
                  <div className="absolute -right-2 md:-right-6 top-6 w-4 h-4 bg-purple-500 rounded-full border-4 border-gray-900"></div>
                  <h3 className="text-2xl font-bold text-white mb-4">
                    2. Intelligent Placement
                  </h3>
                  <p className="text-gray-300 mb-6 leading-relaxed">
                    KubeStellar&apos;s intelligent engine analyzes cluster
                    capacity, network topology, and policy constraints to
                    determine optimal workload placement across your
                    infrastructure.
                  </p>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="bg-green-500/10 backdrop-blur-sm rounded-lg p-3 border border-green-500/30">
                      <div className="text-green-400 text-sm font-medium">
                        US-East Cluster
                      </div>
                      <div className="text-xs text-gray-400">
                        ✓ Policy Match
                      </div>
                    </div>
                    <div className="bg-blue-500/10 backdrop-blur-sm rounded-lg p-3 border border-blue-500/30">
                      <div className="text-blue-400 text-sm font-medium">
                        EU-West Cluster
                      </div>
                      <div className="text-xs text-gray-400">○ Evaluating</div>
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex justify-center md:w-1/2 md:pr-12 mt-8 md:mt-0">
                <div className="relative w-24 h-24 bg-gradient-to-br from-purple-500 to-pink-600 rounded-full flex items-center justify-center shadow-lg">
                  <svg
                    className="w-12 h-12 text-white"
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
              </div>
            </div>
          </div>

          {/* Step 3 */}
          <div className="relative">
            <div className="flex flex-col md:flex-row items-center">
              <div className="md:w-1/2 md:pr-12">
                <div className="relative">
                  <div className="absolute -left-2 md:-left-6 top-6 w-4 h-4 bg-green-500 rounded-full border-4 border-gray-900"></div>
                  <h3 className="text-2xl font-bold text-white mb-4">
                    3. Deploy & Monitor
                  </h3>
                  <p className="text-gray-300 mb-6 leading-relaxed">
                    Workloads are automatically deployed to selected clusters
                    with continuous monitoring, health checks, and automatic
                    failover capabilities for maximum reliability.
                  </p>
                  <div className="flex items-center space-x-4">
                    <div className="flex items-center space-x-2">
                      <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
                      <span className="text-sm text-green-400">
                        5 Clusters Active
                      </span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <div className="w-3 h-3 bg-blue-500 rounded-full animate-pulse"></div>
                      <span className="text-sm text-blue-400">Monitoring</span>
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex justify-center md:w-1/2 md:pl-12 mt-8 md:mt-0">
                <div className="relative w-24 h-24 bg-gradient-to-br from-green-500 to-teal-600 rounded-full flex items-center justify-center shadow-lg">
                  <svg
                    className="w-12 h-12 text-white"
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
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
