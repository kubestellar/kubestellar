"use client";

import { useEffect } from "react";

export default function UseCasesSection() {
  useEffect(() => {
    // Create starfield and grid for Use Cases section
    const createStarfield = (container: HTMLElement) => {
      if (!container) return;
      container.innerHTML = "";

      for (let layer = 1; layer <= 3; layer++) {
        const layerDiv = document.createElement("div");
        layerDiv.className = `star-layer layer-${layer}`;
        layerDiv.style.position = "absolute";
        layerDiv.style.inset = "0";
        layerDiv.style.zIndex = layer.toString();

        const starCount = layer === 1 ? 70 : layer === 2 ? 45 : 25;

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

    const starsContainer = document.getElementById("stars-container-use");
    const gridContainer = document.getElementById("grid-lines-use");

    if (starsContainer) createStarfield(starsContainer);
    if (gridContainer) createGrid(gridContainer);
  }, []);

  const useCases = [
    {
      icon: (
        <svg
          className="w-8 h-8 text-white"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      ),
      title: "Multi-Cloud Deployment",
      description:
        "Deploy applications across AWS, Azure, GCP, and on-premises clusters with unified management and consistent policies.",
      color: "from-blue-500 to-cyan-500",
    },
    {
      icon: (
        <svg
          className="w-8 h-8 text-white"
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
      ),
      title: "Edge Computing",
      description:
        "Distribute workloads to edge locations for reduced latency and improved user experience while maintaining centralized control.",
      color: "from-purple-500 to-pink-500",
    },
    {
      icon: (
        <svg
          className="w-8 h-8 text-white"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
          />
        </svg>
      ),
      title: "Disaster Recovery",
      description:
        "Implement robust disaster recovery strategies with automatic failover and data replication across geographically distributed clusters.",
      color: "from-green-500 to-teal-500",
    },
    {
      icon: (
        <svg
          className="w-8 h-8 text-white"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"
          />
        </svg>
      ),
      title: "DevOps Acceleration",
      description:
        "Streamline CI/CD pipelines with automated testing across multiple environments and seamless production deployments.",
      color: "from-orange-500 to-red-500",
    },
    {
      icon: (
        <svg
          className="w-8 h-8 text-white"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
          />
        </svg>
      ),
      title: "Enterprise Governance",
      description:
        "Enforce compliance policies, security standards, and resource quotas consistently across all clusters and environments.",
      color: "from-indigo-500 to-purple-500",
    },
    {
      icon: (
        <svg
          className="w-8 h-8 text-white"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
          />
        </svg>
      ),
      title: "Cost Optimization",
      description:
        "Optimize infrastructure costs by intelligently placing workloads based on resource pricing and availability across regions.",
      color: "from-yellow-500 to-orange-500",
    },
  ];

  return (
    <section
      id="use-cases"
      className="relative py-16 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden"
    >
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]"></div>

      {/* Starfield background */}
      <div
        id="stars-container-use"
        className="absolute inset-0 overflow-hidden"
      ></div>

      {/* Grid lines background */}
      <div id="grid-lines-use" className="absolute inset-0 opacity-20"></div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-extrabold text-white sm:text-4xl">
            Use <span className="text-gradient">Cases</span>
          </h2>
          <p className="mt-4 max-w-2xl mx-auto text-xl text-gray-300">
            Discover how KubeStellar solves real-world multi-cluster challenges
            across industries
          </p>
        </div>

        <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
          {useCases.map((useCase, index) => (
            <div
              key={index}
              className="group relative bg-gradient-to-br from-gray-800/50 to-gray-900/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50 transition-all duration-300 hover:scale-105 hover:shadow-2xl"
            >
              <div
                className={`absolute inset-0 bg-gradient-to-br ${useCase.color} opacity-0 group-hover:opacity-10 rounded-xl transition-opacity duration-300`}
              ></div>
              <div className="relative z-10">
                <div
                  className={`w-16 h-16 bg-gradient-to-br ${useCase.color} rounded-lg flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-300`}
                >
                  {useCase.icon}
                </div>
                <h3 className="text-xl font-semibold text-white mb-4 group-hover:text-white transition-colors">
                  {useCase.title}
                </h3>
                <p className="text-gray-300 leading-relaxed group-hover:text-gray-200 transition-colors">
                  {useCase.description}
                </p>
                <div className="mt-6 flex items-center text-sm font-medium opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                  <span
                    className={`bg-gradient-to-r ${useCase.color} bg-clip-text text-transparent`}
                  >
                    Learn more
                  </span>
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
          ))}
        </div>
      </div>
    </section>
  );
}
