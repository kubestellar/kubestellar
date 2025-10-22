"use client";

import { useEffect } from "react";
import Link from "next/link";
import { Link as IntlLink } from "@/i18n/navigation";
import { GridLines, StarField, EarthAnimation } from "../index";
import { useTranslations } from "next-intl";

export default function HeroSection() {
  const t = useTranslations("heroSection");
  useEffect(() => {
    // Enhanced typing animation for terminal
    const initTypingAnimation = () => {
      const typingText = document.querySelector(".typing-text") as HTMLElement;
      const commandResponse = document.querySelector(
        ".command-response"
      ) as HTMLElement;

      if (typingText && commandResponse) {
        const lines = [
          "bash <(curl -s \\",
          "    https://raw.githubusercontent.com/kubestellar/ \\",
          "    kubestellar/refs/tags/v0.27.2/scripts/ \\",
          "    create-kubestellar-demo-env.sh) --platform kind",
        ];

        // Clear initial content
        const divElements = typingText.querySelectorAll("div");
        divElements.forEach(div => (div.textContent = ""));

        let lineIndex = 0;
        let charIndex = 0;

        const typeNextChar = () => {
          if (lineIndex < lines.length) {
            const currentLine = lines[lineIndex];
            const currentDiv = divElements[lineIndex];

            if (charIndex < currentLine.length) {
              if (currentDiv) {
                currentDiv.textContent += currentLine.charAt(charIndex);
              }
              charIndex++;
              setTimeout(typeNextChar, 30);
            } else {
              lineIndex++;
              charIndex = 0;
              setTimeout(typeNextChar, 100); // Pause between lines
            }
          } else {
            setTimeout(() => {
              commandResponse.style.opacity = "1";
            }, 500);
          }
        };

        typeNextChar();
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

    // Initialize components
    initTypingAnimation();
    animateCounters();
  }, []);

  return (
    <section className="relative overflow-hidden bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white min-h-[85vh] flex items-center">
      {/* Animated Background Universe */}
      <div className="absolute inset-0 z-0">
        {/*!-- Floating Nebula Clouds */}
        {/* Dynamic Star Field */}
        <div className="absolute inset-0 bg-[#0a0a0a]">
          <StarField density="medium" showComets={true} cometCount={8} />
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
                  <span className="text-gradient">{t("line1")}</span>
                </span>

                {/* Second Line with delay */}
                <span className="block animate-text-reveal">
                  <span className="text-gradient-animated">{t("line2")}</span>
                </span>

                {/* Third Line with longer delay */}
                <span className="block animate-text-reveal [animation-delay:0.4s]">
                  <span className="text-gradient-animated">{t("line3")}</span>
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
                    <span className="text-green-400 text-xs">READY</span>
                  </div>
                </div>

                {/* Terminal Content */}
                <div className="terminal-content space-y-3 font-mono text-sm">
                  {/* Command Line */}
                  <div className="command-line animate-command-typing">
                    <div className="flex flex-col space-y-1">
                      <div className="flex items-start">
                        <span className="text-green-400 mr-2 flex-shrink-0">
                          $
                        </span>
                        <div className="typing-text text-blue-300 leading-relaxed text-xs sm:text-sm">
                          <div>bash &lt;(curl -s \</div>
                          <div className="ml-4">
                            https://raw.githubusercontent.com/kubestellar/ \
                          </div>
                          <div className="ml-4">
                            kubestellar/refs/tags/v0.27.2/scripts/ \
                          </div>
                          <div className="ml-4">
                            create-kubestellar-demo-env.sh) --platform kind
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Command Output */}
                  <div className="command-output space-y-2 ml-6 animate-fade-in [animation-delay:0.8s] [animation-fill-mode:forwards]">
                    {/* Header */}
                    <div className="output-line animate-slide-in-left [animation-delay:1s]">
                      <span className="text-cyan-400 font-bold">INFO</span>
                      <span className="text-gray-300 ml-4">
                        Installing KubeStellar demo environment...
                      </span>
                    </div>

                    {/* Creating clusters */}
                    <div className="output-line animate-slide-in-left [animation-delay:1.2s]">
                      <span className="text-blue-400 font-bold">SETUP</span>
                      <span className="text-gray-300 ml-4">
                        Creating kind clusters: kubeflex, cluster1, cluster2
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
                    </div>

                    {/* Installing KubeFlex */}
                    <div className="output-line animate-slide-in-left [animation-delay:1.4s]">
                      <span className="text-purple-400 font-bold">INSTALL</span>
                      <span className="text-gray-300 ml-4">
                        Deploying KubeFlex control plane components
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
                    </div>

                    {/* Configuring OCM */}
                    <div className="output-line animate-slide-in-left [animation-delay:1.6s]">
                      <span className="text-yellow-400 font-bold">CONFIG</span>
                      <span className="text-gray-300 ml-4">
                        Configuring Open Cluster Management
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
                    </div>

                    {/* Final Success */}
                    <div className="output-line animate-slide-in-left [animation-delay:1.8s]">
                      <span className="text-emerald-400 font-bold">
                        SUCCESS
                      </span>
                      <span className="text-gray-300 ml-4">
                        KubeStellar demo environment ready! Setup complete
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
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
              <IntlLink
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
              </IntlLink>

              <Link
                href="/docs"
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
              </Link>
            </div>
          </div>

          {/* Right Column: Earth Animation */}
          <div className="earth-animation-container relative h-[500px] flex items-center justify-center">
            <EarthAnimation
              width="100%"
              height="500px"
              scale={3.5}
              autoRotate={true}
              autoRotateSpeed={0.5}
              enableZoom={false}
              fov={50}
              cameraPosition={[-4, 2, 6]}
              className="rounded-xl overflow-hidden"
              style={{
                filter: "drop-shadow(0 25px 50px rgba(59, 130, 246, 0.3))",
              }}
            />
          </div>
        </div>
      </div>
    </section>
  );
}
