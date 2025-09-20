"use client";

import { useEffect } from "react";
import StarField from "../StarField";

export default function HeroSection() {
  useEffect(() => {
    // Interactive grid canvas
    const initGridCanvas = () => {
      const canvas = document.getElementById(
        "grid-canvas"
      ) as HTMLCanvasElement;
      if (!canvas) return;

      const ctx = canvas.getContext("2d");
      if (!ctx) return;

      canvas.width = canvas.offsetWidth;
      canvas.height = canvas.offsetHeight;

      const gridSize = 50;

      const drawGrid = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.strokeStyle = "rgba(99, 102, 241, 0.2)";
        ctx.lineWidth = 0.5;

        // Draw vertical lines
        for (let x = 0; x < canvas.width; x += gridSize) {
          ctx.beginPath();
          ctx.moveTo(x, 0);
          ctx.lineTo(x, canvas.height);
          ctx.stroke();
        }

        // Draw horizontal lines
        for (let y = 0; y < canvas.height; y += gridSize) {
          ctx.beginPath();
          ctx.moveTo(0, y);
          ctx.lineTo(canvas.width, y);
          ctx.stroke();
        }
      };

      drawGrid();

      // Redraw on resize
      window.addEventListener("resize", () => {
        canvas.width = canvas.offsetWidth;
        canvas.height = canvas.offsetHeight;
        drawGrid();
      });
    };

    // Enhanced typing animation for terminal
    const initTypingAnimation = () => {
      const typingText = document.querySelector(".typing-text") as HTMLElement;
      const commandResponse = document.querySelector(
        ".command-response"
      ) as HTMLElement;

      if (typingText && commandResponse) {
        const text = typingText.textContent || "";
        typingText.textContent = "";

        let i = 0;
        const typeInterval = setInterval(() => {
          if (i < text.length) {
            typingText.textContent += text.charAt(i);
            i++;
          } else {
            clearInterval(typeInterval);
            setTimeout(() => {
              commandResponse.style.opacity = "1";
            }, 500);
          }
        }, 50);
      }
    };

    // Animated counters
    const animateCounters = () => {
      const counters = document.querySelectorAll(".counter");
      counters.forEach(counter => {
        const target = parseInt(counter.getAttribute("data-target") || "0");
        const duration = 2000;
        const step = target / (duration / 16);
        let current = 0;

        const timer = setInterval(() => {
          current += step;
          if (current >= target) {
            current = target;
            clearInterval(timer);
          }
          counter.textContent = Math.floor(current).toString();
        }, 16);
      });
    };

    // Initialize all components
    initGridCanvas();
    initTypingAnimation();
    animateCounters();

    // Enhanced 3D hover effect for mission control scene
    const sceneContainer = document.querySelector(
      ".scene-container"
    ) as HTMLElement;
    const missionControlScene = document.querySelector(
      ".mission-control-scene"
    ) as HTMLElement;

    if (sceneContainer && missionControlScene) {
      // Initial animation to bring the scene into view
      setTimeout(() => {
        sceneContainer.style.transform =
          "perspective(1000px) rotateY(3deg) rotateX(2deg) translateZ(0)";
        sceneContainer.style.opacity = "1";
      }, 500);

      missionControlScene.addEventListener("mousemove", e => {
        const rect = missionControlScene.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        const centerX = rect.width / 2;
        const centerY = rect.height / 2;

        // Calculate rotation values with easing
        const rotateY = (x - centerX) / 40;
        const rotateX = (centerY - y) / 40;

        // Apply smooth transition
        sceneContainer.style.transition = "transform 0.1s ease-out";
        sceneContainer.style.transform = `perspective(1000px) rotateY(${rotateY}deg) rotateX(${rotateX}deg) translateZ(10px)`;
      });

      // Reset on mouse leave with smooth transition
      missionControlScene.addEventListener("mouseleave", () => {
        sceneContainer.style.transition = "transform 0.5s ease-out";
        sceneContainer.style.transform =
          "perspective(1000px) rotateY(3deg) rotateX(2deg) translateZ(0)";
      });
    }
  }, []);

  return (
    <section className="relative overflow-hidden bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white min-h-[85vh] flex items-center">
      {/* Animated Background Universe */}
      <div className="absolute inset-0 z-0">
        {/* Dynamic Star Field */}
        <div className="absolute inset-0 bg-[#0a0a0a]">
          <StarField density="high" showComets={true} cometCount={8} />
        </div>

        {/* Interactive Grid Network */}
        <div className="absolute inset-0">
          <canvas
            id="grid-canvas"
            className="w-full h-full opacity-20"
          ></canvas>
        </div>

        {/* Floating Data Particles */}
        <div className="absolute inset-0">
          <div
            className="data-particle"
            style={{ "--delay": "0s" } as React.CSSProperties}
          ></div>
          <div
            className="data-particle"
            style={{ "--delay": "1s" } as React.CSSProperties}
          ></div>
          <div
            className="data-particle"
            style={{ "--delay": "2s" } as React.CSSProperties}
          ></div>
          <div
            className="data-particle"
            style={{ "--delay": "3s" } as React.CSSProperties}
          ></div>
          <div
            className="data-particle"
            style={{ "--delay": "4s" } as React.CSSProperties}
          ></div>
          <div
            className="data-particle"
            style={{ "--delay": "5s" } as React.CSSProperties}
          ></div>
        </div>
      </div>

      {/* Main Content Container */}
      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 lg:py-20">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center min-h-[70vh]">
          {/* Left Column: Interactive Content */}
          <div className="hero-content space-y-6">
            {/* Animated Status Badge */}
            <div
              className="status-badge-container animate-fade-in-up"
              style={{ animationDelay: "0.2s" }}
            >
              <div className="inline-flex items-center px-4 py-2 rounded-full bg-gradient-to-r from-green-500/20 to-blue-500/20 border border-green-500/30 animate-status-glow">
                <div className="relative mr-3">
                  <div className="status-dot"></div>
                  <div className="status-ripple"></div>
                </div>
                <span className="text-sm font-medium text-green-400">
                  Platform Status: Operational
                </span>
              </div>
            </div>

            {/* Dynamic Main Heading */}
            <div
              className="heading-container space-y-4 animate-text-reveal"
              style={{ animationDelay: "0.4s" }}
            >
              <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold leading-tight">
                <span className="block text-white">Multi-Cluster</span>
                <span className="block text-gradient-animated">Kubernetes</span>
                <span className="block text-white">Orchestration</span>
              </h1>
              <p className="text-xl md:text-2xl text-gray-300 leading-relaxed max-w-2xl">
                Simplify multi-cluster Kubernetes operations with intelligent
                workload distribution, unified management, and seamless
                orchestration across any infrastructure.
              </p>
            </div>

            {/* Interactive Command Center */}
            <div
              className="command-center-container animate-slide-in-left"
              style={{ animationDelay: "0.6s" }}
            >
              <div className="bg-gray-900/80 backdrop-blur-sm rounded-lg border border-gray-700/50 p-4 animate-command-glow">
                <div className="flex items-center mb-2">
                  <div className="flex space-x-2">
                    <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                    <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
                    <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  </div>
                  <span className="ml-4 text-gray-400 text-sm">
                    KubeStellar Terminal
                  </span>
                </div>
                <div className="font-mono text-sm">
                  <div className="text-green-400">
                    $ kubectl apply -f kubestellar-deployment.yaml
                  </div>
                  <div
                    className="command-response text-blue-400 mt-2"
                    style={{ opacity: 0 }}
                  >
                    ✅ Deployment successfully distributed to 5 clusters
                    <br />
                    📊 Monitoring workload health across regions
                    <br />
                    <span className="text-green-400">
                      🚀 KubeStellar is ready!
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Interactive Action Buttons */}
            <div
              className="action-buttons-container flex flex-col sm:flex-row gap-4 animate-fade-in"
              style={{ animationDelay: "0.8s" }}
            >
              <button className="primary-action-btn btn-primary px-8 py-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 rounded-lg font-semibold text-lg transition-all duration-300 animate-btn-float relative overflow-hidden group">
                <span className="relative z-10">Get Started</span>
                <div className="btn-shine"></div>
                <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
              </button>
              <button className="secondary-action-btn px-8 py-4 border-2 border-gray-600 hover:border-blue-500 rounded-lg font-semibold text-lg transition-all duration-300 hover-lift">
                <span className="flex items-center">
                  <svg
                    className="w-5 h-5 mr-2"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1M9 16h1m4 0h1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  Live Demo
                </span>
              </button>
            </div>

            {/* Interactive Stats Dashboard */}
            <div
              className="stats-dashboard grid grid-cols-1 sm:grid-cols-3 gap-4 pt-6 animate-fade-in"
              style={{ animationDelay: "1s" }}
            >
              <div className="stat-card stat-card-1 relative bg-gray-800/50 backdrop-blur-sm rounded-lg p-4 border border-gray-700/50 hover-lift animate-stat-float">
                <div className="stat-glow"></div>
                <div className="relative z-10">
                  <div
                    className="text-2xl font-bold text-green-400 counter"
                    data-target="50"
                  >
                    0
                  </div>
                  <div className="text-sm text-gray-400">Active Clusters</div>
                </div>
              </div>
              <div className="stat-card stat-card-2 relative bg-gray-800/50 backdrop-blur-sm rounded-lg p-4 border border-gray-700/50 hover-lift animate-stat-float">
                <div className="stat-glow"></div>
                <div className="relative z-10">
                  <div
                    className="text-2xl font-bold text-pink-400 counter"
                    data-target="1000"
                  >
                    0
                  </div>
                  <div className="text-sm text-gray-400">Workloads Managed</div>
                </div>
              </div>
              <div className="stat-card stat-card-3 relative bg-gray-800/50 backdrop-blur-sm rounded-lg p-4 border border-gray-700/50 hover-lift animate-stat-float">
                <div className="stat-glow"></div>
                <div className="relative z-10">
                  <div className="text-2xl font-bold text-yellow-400">
                    99.9%
                  </div>
                  <div className="text-sm text-gray-400">Uptime SLA</div>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column: Astronaut Mission Control Scene */}
          <div className="mission-control-scene relative h-[500px]">
            {/* Main scene container with perspective */}
            <div className="absolute inset-0 scene-container perspective-1000">
              {/* Holographic Control Panels */}
              <div className="absolute top-8 left-8 control-panel bg-blue-500/10 backdrop-blur-sm rounded-lg p-4 border border-blue-500/30">
                <div className="text-xs text-blue-400 mb-2">CLUSTER STATUS</div>
                <div className="space-y-2">
                  <div className="flex justify-between items-center">
                    <span className="text-xs text-gray-300">US-East</span>
                    <div className="w-12 h-1.5 bg-gray-700 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-green-500 rounded-full"
                        style={{ width: "85%" }}
                      ></div>
                    </div>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-xs text-gray-300">EU-West</span>
                    <div className="w-12 h-1.5 bg-gray-700 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-green-500 rounded-full"
                        style={{ width: "92%" }}
                      ></div>
                    </div>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-xs text-gray-300">Asia-Pacific</span>
                    <div className="w-12 h-1.5 bg-gray-700 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-yellow-500 rounded-full"
                        style={{ width: "78%" }}
                      ></div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Metrics Panel */}
              <div className="absolute top-8 right-8 metrics-panel bg-purple-500/10 backdrop-blur-sm rounded-lg p-4 border border-purple-500/30">
                <div className="text-xs text-purple-400 mb-2">PERFORMANCE</div>
                <div className="space-y-1 text-xs">
                  <div className="flex justify-between">
                    <span className="text-gray-300">CPU</span>
                    <span className="text-green-400">67%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-300">Memory</span>
                    <span className="text-blue-400">43%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-300">Network</span>
                    <span className="text-yellow-400">28%</span>
                  </div>
                </div>
              </div>

              {/* Central Astronaut */}
              <div className="absolute inset-0 flex items-center justify-center astronaut-container">
                <div className="relative astronaut">
                  <svg
                    width="120"
                    height="120"
                    viewBox="0 0 120 120"
                    className="drop-shadow-2xl"
                  >
                    {/* Helmet */}
                    <circle
                      cx="60"
                      cy="45"
                      r="25"
                      fill="url(#helmet-gradient)"
                      stroke="#3b82f6"
                      strokeWidth="2"
                    />

                    {/* Helmet Visor */}
                    <path
                      d="M 40 45 Q 60 25 80 45 Q 60 65 40 45"
                      fill="url(#visor-gradient)"
                      opacity="0.8"
                    />

                    {/* Body */}
                    <path
                      d="M 45 70 L 75 70 L 78 95 L 42 95 Z"
                      fill="url(#suit-gradient)"
                      stroke="#6366f1"
                      strokeWidth="1"
                    />

                    {/* Arms */}
                    <path
                      d="M 45 75 L 30 85 L 35 90 L 50 80"
                      fill="url(#suit-gradient)"
                      stroke="#6366f1"
                      strokeWidth="1"
                    />
                    <path
                      d="M 75 75 L 90 85 L 85 90 L 70 80"
                      fill="url(#suit-gradient)"
                      stroke="#6366f1"
                      strokeWidth="1"
                    />

                    {/* Control Device */}
                    <circle
                      cx="85"
                      cy="87"
                      r="4"
                      fill="#10b981"
                      className="pulsing"
                    />

                    {/* Gradients */}
                    <defs>
                      <linearGradient
                        id="helmet-gradient"
                        x1="0%"
                        y1="0%"
                        x2="100%"
                        y2="100%"
                      >
                        <stop
                          offset="0%"
                          style={{ stopColor: "#1e293b", stopOpacity: 0.9 }}
                        />
                        <stop
                          offset="100%"
                          style={{ stopColor: "#334155", stopOpacity: 0.9 }}
                        />
                      </linearGradient>
                      <linearGradient
                        id="visor-gradient"
                        x1="0%"
                        y1="0%"
                        x2="100%"
                        y2="100%"
                      >
                        <stop
                          offset="0%"
                          style={{ stopColor: "#3b82f6", stopOpacity: 0.3 }}
                        />
                        <stop
                          offset="100%"
                          style={{ stopColor: "#1d4ed8", stopOpacity: 0.6 }}
                        />
                      </linearGradient>
                      <linearGradient
                        id="suit-gradient"
                        x1="0%"
                        y1="0%"
                        x2="100%"
                        y2="100%"
                      >
                        <stop offset="0%" style={{ stopColor: "#374151" }} />
                        <stop offset="100%" style={{ stopColor: "#4b5563" }} />
                      </linearGradient>
                    </defs>
                  </svg>
                </div>
              </div>

              {/* Floating Connection Lines */}
              <svg className="absolute inset-0 w-full h-full pointer-events-none">
                <path
                  className="connection-path"
                  d="M 60 60 Q 100 80 140 100"
                  stroke="url(#connection-gradient)"
                  strokeWidth="1"
                  fill="none"
                  strokeDasharray="5,5"
                />
                <path
                  className="connection-path"
                  d="M 60 60 Q 20 40 -20 20"
                  stroke="url(#connection-gradient)"
                  strokeWidth="1"
                  fill="none"
                  strokeDasharray="5,5"
                />
                <defs>
                  <linearGradient id="connection-gradient">
                    <stop
                      offset="0%"
                      style={{ stopColor: "#3b82f6", stopOpacity: 0.6 }}
                    />
                    <stop
                      offset="100%"
                      style={{ stopColor: "#8b5cf6", stopOpacity: 0.2 }}
                    />
                  </linearGradient>
                </defs>
              </svg>

              {/* Orbital Rings */}
              <div className="orbit-container animate-spin-slow">
                <div className="absolute top-1/2 left-1/2 w-4 h-4 bg-blue-500 rounded-full transform -translate-x-1/2 -translate-y-1/2 animate-pulse"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
