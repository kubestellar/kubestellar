"use client";

import { useEffect } from "react";
import { GridLines, StarField } from "../index";
import StatCard from "../StatsCard";

interface StatData {
  id: number;
  icon: React.ReactNode;
  value: number;
  suffix: string;
  title: string;
  color: "blue" | "purple" | "emerald";
  animationDelay: string;
}

const statsData: StatData[] = [
  {
    id: 1,
    icon: (
      <svg
        className="w-8 h-8 text-blue-400"
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
    ),
    value: 40,
    suffix: "x",
    title: "Performance Boost",
    color: "blue" as const,
    animationDelay: "0s",
  },
  {
    id: 2,
    icon: (
      <svg
        className="w-8 h-8 text-purple-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z"
        ></path>
      </svg>
    ),
    value: 99,
    suffix: "%",
    title: "Uptime Guarantee",
    color: "purple" as const,
    animationDelay: "0.2s",
  },
  {
    id: 3,
    icon: (
      <svg
        className="w-8 h-8 text-emerald-400"
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
    ),
    value: 50,
    suffix: "k+",
    title: "Active Users",
    color: "emerald" as const,
    animationDelay: "0.4s",
  },
];

export default function HeroSection() {
  useEffect(() => {
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
        {/*!-- Floating Nebula Clouds */}
        {/* IMPLEMENTATION REMAINING WILL DO*/}
        {/* Dynamic Star Field */}
        <div className="absolute inset-0 bg-[#0a0a0a]">
          <StarField density="high" showComets={true} cometCount={8} />
        </div>

        {/* Interactive Grid Network */}
        <div className="absolute inset-0">
          <GridLines verticalLines={15} horizontalLines={18} />
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
            {/* Dynamic Main Heading */}
            <div className="heading-container space-y-4">
              <h1 className="text-4xl sm:text-5xl lg:text-6xl font-black tracking-tight leading-none">
                {/* First Line */}
                <span className="block text-white mb-3 animate-text-reveal pt-5">
                  <span className="text-gradient">Multi-Cluster</span>
                </span>

                {/* Second Line with delay */}
                <span className="block animate-text-reveal">
                  <span className="text-gradient-animated">Kubernetes</span>
                </span>

                {/* Third Line with longer delay */}
                <span className="block animate-text-reveal [animation-delay:0.4s]">
                  <span className="text-gradient-animated">Orchestration</span>
                </span>
              </h1>

              {/* Paragraph with fade-in-up effect and delay */}
              <p className="text-lg sm:text-xl text-gray-300 max-w-2xl leading-snug animate-fade-in-up opacity-0 [animation-delay:0.6s] [animation-fill-mode:forwards]">
                Experience the future of cloud-native orchestration. KubeStellar
                revolutionizes multi-cluster management with AI-powered
                automation and real-time intelligence.
              </p>
            </div>

            {/* Interactive Command Center */}
            <div className="command-center-container">
              <div className="bg-black/40 backdrop-blur-xl border border-gray-700/50 rounded-2xl p-6 shadow-2xl animate-command-glow">
                {/* Terminal Header */}
                <div className="terminal-header flex items-center space-x-3 mb-4">
                  {/* Terminal Control Buttons */}
                  <div className="terminal-controls flex space-x-2">
                    <div className="w-4 h-4 rounded-full bg-red-500 animate-pulse"></div>
                    <div className="w-4 h-4 rounded-full bg-yellow-500 animate-pulse [animation-delay:0.2s]"></div>
                    <div className="w-4 h-4 rounded-full bg-green-500 animate-pulse [animation-delay:0.4s]"></div>
                  </div>

                  {/* Terminal Title */}
                  <span className="text-gray-400 text-sm font-mono">
                    kubestellar-control-center
                  </span>

                  <div className="flex-1"></div>

                  {/* Connection Status */}
                  <div className="connection-status flex items-center space-x-2">
                    <div className="w-2 h-2 bg-green-400 rounded-full animate-ping"></div>
                    <span className="text-green-400 text-xs">CONNECTED</span>
                  </div>
                </div>

                {/* Terminal Content */}
                <div className="terminal-content space-y-3 font-mono text-sm">
                  {/* Command Line */}
                  <div className="command-line animate-command-typing">
                    <span className="text-green-400 mr-5">$</span>
                    <span className="typing-text text-blue-300">
                      kubectl kubestellar deploy --multi-cluster --ai-optimized
                    </span>
                    &nbsp;
                    <span className="typing-cursor bg-blue-300 w-0.5 h-6 animate-blink"></span>
                  </div>

                  {/* Command Output */}
                  <div className="command-output space-y-2 ml-6 animate-fade-in [animation-delay:0.8s] [animation-fill-mode:forwards]">
                    {/* Line 1 */}
                    <div className="output-line animate-slide-in-left [animation-delay:1s]">
                      <span className="text-cyan-400 font-bold">AI</span>
                      <span className="text-gray-300 ml-4">
                        Analyzing cluster topology and workload patterns...
                      </span>
                      <div className="loading-dots ml-2" aria-label="Loading">
                        <span aria-hidden="true"></span>
                        <span aria-hidden="true"></span>
                        <span aria-hidden="true"></span>
                      </div>
                    </div>

                    {/* Line 2 */}
                    <div className="output-line animate-slide-in-left [animation-delay:1.4s]">
                      <span className="text-blue-400 font-bold">INFO</span>
                      <span className="text-gray-300 ml-4">
                        Discovered 8 clusters across 3 regions
                      </span>
                      <svg
                        className="w-4 h-4 text-green-400 ml-2 animate-bounce"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M5 13l4 4L19 7"
                        ></path>
                      </svg>
                    </div>

                    {/* Line 3 */}
                    <div className="output-line animate-slide-in-left [animation-delay:1.8s]">
                      <span className="text-purple-400 font-bold">
                        OPTIMIZE
                      </span>
                      <span className="text-gray-300 ml-4">
                        AI optimizing resource allocation...
                      </span>
                      <div className="optimization-bar ml-2">
                        <div className="optimization-progress"></div>
                      </div>
                    </div>

                    {/* Line 4 */}
                    <div className="output-line animate-slide-in-left [animation-delay:2.2s]">
                      <span className="text-emerald-400 font-bold">
                        SUCCESS
                      </span>
                      <span className="text-gray-300 ml-4">
                        Deployment completed with 40% efficiency gain
                      </span>
                      <div className="success-indicator ml-2">
                        <svg
                          className="w-5 h-5 text-emerald-400 animate-ping"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth="2"
                            d="M5 13l4 4L19 7"
                          ></path>
                        </svg>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Interactive Action Buttons */}
            <div
              className="action-buttons-container flex flex-col sm:flex-row gap-4 animate-btn-float"
              style={{ animationDelay: "0.8s" }}
            >
              <a
                href="/quick-installation"
                className="primary-action-btn group relative overflow-hidden inline-flex items-center justify-center px-8 py-4 text-lg font-bold rounded-xl text-white 
                          bg-gradient-to-r from-blue-600 via-purple-600 to-indigo-600 
                          hover:from-blue-700 hover:via-purple-700 hover:to-indigo-700 
                          transition-all duration-500 transform hover:-translate-y-1 
                          hover:shadow-xl hover:shadow-blue-500/40 
                          animate-btn-float"
              >
                <span className="relative z-10">Install KubeStellar</span>
                <svg
                  className="relative z-10 ml-2 h-5 w-5 transition-all duration-300 group-hover:translate-x-1 group-hover:rotate-12"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M10.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L12.586 11H5a1 1 0 110-2h7.586l-2.293-2.293a1 1 0 010-1.414z"
                    clipRule="evenodd"
                  ></path>
                </svg>
                <div className="btn-shine"></div>
              </a>

              <a
                href="#documentation"
                className="secondary-action-btn inline-flex items-center justify-center px-8 py-4 text-lg font-bold rounded-xl text-gray-200 
                          bg-gray-800/40 hover:bg-gray-800/60 
                          backdrop-blur-md border border-gray-700/50 hover:border-gray-600/50 
                          transition-all duration-500 transform hover:-translate-y-1 hover:scale-105 
                          animate-btn-float"
                style={{ animationDelay: "0.1s" }}
              >
                <svg
                  className="mr-2 h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                  ></path>
                </svg>
                Explore Docs
              </a>
            </div>

            {/* STATS DASHBOARD */}
            <div className="stats-dashboard grid grid-cols-1 sm:grid-cols-3 gap-4 pt-6">
              {statsData.map(stat => (
                <StatCard
                  key={stat.id}
                  icon={stat.icon}
                  value={stat.value}
                  suffix={stat.suffix}
                  title={stat.title}
                  color={stat.color}
                  animationDelay={stat.animationDelay}
                />
              ))}
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
