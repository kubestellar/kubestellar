"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Link as IntlLink } from "@/i18n/navigation";
import { GridLines, StarField, GlobeAnimation } from "../index";
import { useTranslations } from "next-intl";

export default function HeroSection() {
  const t = useTranslations("heroSection");
  const [copied, setCopied] = useState(false);

  const installScript = `bash <(curl -s \\
  https://raw.githubusercontent.com/kubestellar/kubestellar/refs/tags/v0.27.2/scripts/create-kubestellar-demo-env.sh) \\
  --platform kind`;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(installScript);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  };
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

    // Initialize components
    initTypingAnimation();
    animateCounters();
  }, []);

  return (
    <section className="relative overflow-hidden bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white min-h-[100vh] flex items-center">
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
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center min-h-screen lg:min-h-[70vh]">
          {/* Left Column: Interactive Content */}
          <div className="hero-content space-y-6 sm:space-y-8 lg:space-y-12 min-h-[calc(100vh-4rem)] lg:min-h-0 flex flex-col justify-center lg:block pt-8 md:pt-0">
            {/* Dynamic Main Heading */}
            <div className="heading-container">
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
              <p className="sm:text-xl text-gray-300 max-w-2xl leading-snug animate-fade-in-up opacity-0 [animation-delay:0.6s] [animation-fill-mode:forwards] mt-4">
                {t("subtitle")}
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
                    {t("terminalTitle")}
                  </span>

                  <div className="flex-1"></div>

                  {/* Connection Status */}
                  <div className="connection-status flex items-center space-x-2">
                    <div className="w-2 h-2 bg-green-400 rounded-full animate-ping"></div>
                    <span className="text-green-400 text-xs">{t("terminalStatus")}</span>
                  </div>

                  {/* Copy Button */}
                  <button
                    onClick={handleCopy}
                    className={`copy-button ml-3 rounded-md p-2 
                              transition-all duration-200 group relative ${copied ? "copy-success" : ""}`}
                  >
                    {copied ? (
                      <svg
                        className="w-4 h-4 text-green-400"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    ) : (
                      <svg
                        className="w-4 h-4 text-sky-400 hover:text-sky-300 transition-colors"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                        />
                      </svg>
                    )}
                  </button>
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
                            https://raw.githubusercontent.com/kubestellar/kubestellar/refs/tags/v0.27.2/scripts/create-kubestellar-demo-env.sh) \
                          </div>
                          <div className="ml-4">
                            --platform kind
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Command Output */}
                  <div className="command-output space-y-2 ml-6 animate-fade-in [animation-delay:0.8s] [animation-fill-mode:forwards]">
                    {/* Header */}
                    <div className="output-line flex animate-slide-in-left [animation-delay:1s]">
                      <span className="text-cyan-400 font-bold w-22 inline-block">
                        {t("terminalOutputInfo")}
                      </span>
                      <span className="text-gray-300 flex-1">
                        {t("terminalOutputInfoText")}
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
                    </div>

                    {/* Creating clusters */}
                    <div className="output-line flex animate-slide-in-left [animation-delay:1.2s]">
                      <span className="text-blue-400 font-bold w-22 inline-block">
                        {t("terminalOutputSetup")}
                      </span>
                      <span className="text-gray-300 flex-1">
                        {t("terminalOutputSetupText")}
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
                    </div>

                    {/* Installing KubeFlex */}
                    <div className="output-line flex animate-slide-in-left [animation-delay:1.4s]">
                      <span className="text-purple-400 font-bold w-22 inline-block">
                        {t("terminalOutputInstall")}
                      </span>
                      <span className="text-gray-300 flex-1">
                        {t("terminalOutputInstallText")}
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
                    </div>

                    {/* Configuring OCM */}
                    <div className="output-line flex animate-slide-in-left [animation-delay:1.6s]">
                      <span className="text-yellow-400 font-bold w-22 inline-block">
                        {t("terminalOutputConfig")}
                      </span>
                      <span className="text-gray-300 flex-1">
                        {t("terminalOutputConfigText")}
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
                    </div>

                    {/* Final Success */}
                    <div className="output-line flex animate-slide-in-left [animation-delay:1.8s]">
                      <span className="text-emerald-400 font-bold w-22 inline-block">
                        {t("terminalOutputSuccess")}
                      </span>
                      <span className="text-gray-300 flex-1">
                        {t("terminalOutputSuccessText")}
                      </span>
                      <span className="text-emerald-400 ml-2 text-xs">✓</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Interactive Action Buttons */}
            <div
              className="action-buttons-container space-y-5 animate-btn-float"
              style={{ animationDelay: "0.8s" }}
            >
              {/* Installation Path Heading */}
              <div className="text-center sm:text-left">
                <h3 className="text-xl font-bold text-white mb-2">
                  {t("getStartedWith")}{" "}
                  <span className="text-gradient animated-gradient bg-gradient-to-r from-blue-400 via-purple-400 to-blue-400">
                    KubeStellar
                  </span>
                </h3>
                <p className="text-sm text-blue-100/80">
                  {t("chooseEnvironment")}
                </p>
              </div>

              {/* Installation Buttons Row */}
              <div className="flex flex-col sm:flex-row gap-4">
                {/* Local Development Button */}
                <IntlLink
                  href="/quick-installation"
                  className="group relative overflow-hidden flex items-center justify-between px-5 py-4 rounded-lg text-white 
                            bg-slate-800/60 backdrop-blur-sm
                            hover:bg-slate-800/80
                            transition-all duration-300 transform hover:scale-[1.02]
                            border border-blue-500/30 hover:border-blue-400/50
                            hover:shadow-lg hover:shadow-blue-500/20"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center flex-shrink-0">
                      <svg
                        className="w-5 h-5 text-blue-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                        />
                      </svg>
                    </div>
                    <div className="text-left">
                      <div className="text-sm font-semibold text-white">
                        {t("localDevelopment")}
                      </div>
                      <div className="text-xs text-blue-200/70">
                        {t("localDevelopmentTime")}
                      </div>
                    </div>
                  </div>
                  <svg
                    className="w-5 h-5 text-blue-400 transition-transform duration-300 group-hover:translate-x-1 flex-shrink-0"
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
                </IntlLink>

                {/* AWS EKS Cloud Button */}
                <Link
                  href="/docs/getting-started/aws-eks"
                  className="group relative overflow-hidden flex items-center justify-between px-5 py-4 rounded-lg text-white 
                            bg-slate-800/60 backdrop-blur-sm
                            hover:bg-slate-800/80
                            transition-all duration-300 transform hover:scale-[1.02]
                            border border-purple-500/30 hover:border-purple-400/50
                            hover:shadow-lg hover:shadow-purple-500/20"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-purple-500/20 flex items-center justify-center flex-shrink-0">
                      <svg
                        className="w-5 h-5 text-purple-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z"
                        />
                      </svg>
                    </div>
                    <div className="text-left">
                      <div className="text-sm font-semibold text-white">
                        {t("awsEksProduction")}
                      </div>
                      <div className="text-xs text-purple-200/70">
                        {t("awsEksProductionTime")}
                      </div>
                    </div>
                  </div>
                  <svg
                    className="w-5 h-5 text-purple-400 transition-transform duration-300 group-hover:translate-x-1 flex-shrink-0"
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
                </Link>
              </div>

              {/* Explore Docs Button */}
              <div className="flex justify-center sm:justify-start">
                <Link
                  href="/docs"
                  className="secondary-action-btn inline-flex items-center justify-center px-6 py-3 text-sm font-semibold rounded-lg text-gray-200 
                            bg-gray-800/40 hover:bg-gray-800/60 
                            backdrop-blur-md border border-gray-700/50 hover:border-gray-600/50 
                            transition-all duration-500 transform hover:-translate-y-1 
                            animate-btn-float"
                  style={{ animationDelay: "0.2s" }}
                >
                  <svg
                    className="mr-2 h-4 w-4"
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
                  {t("buttonDocs")}
                </Link>
              </div>
            </div>
          </div>

          {/* Right Column: Globe Animation */}
          <div className="globe-animation-container relative h-[500px] flex items-center justify-center">
            <GlobeAnimation
              width="100%"
              height="600px"
              className="rounded-xl overflow-hidden"
              showLoader={true}
              enableControls={true}
              enablePan={false}
              autoRotate={true}
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
