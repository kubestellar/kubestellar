"use client";

import Link from "next/link";
import { GridLines, StarField } from "../index";
import { useTranslations } from "next-intl";
import { getLocalizedUrl } from "@/lib/url";

const Icon = ({
  path,
  className = "h-6 w-6 text-white",
  strokeColor = "currentColor",
  strokeWidth = "2",
}: {
  path: string;
  className?: string;
  strokeColor?: string;
  strokeWidth?: string;
}) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke={strokeColor}
    strokeWidth={strokeWidth}
  >
    <path strokeLinecap="round" strokeLinejoin="round" d={path} />
  </svg>
);

export default function GetStartedSection() {
  const t = useTranslations("getStartedSection");

  return (
    <section
      id="get-started"
      className="relative py-16 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden"
    >
      <div className="absolute inset-0 bg-[#0a0a0a] pointer-events-none"></div>

      {/* Starfield background */}
      <StarField density="low" showComets={true} cometCount={3} />

      {/* Gridlines background */}
      <GridLines verticalLines={20} horizontalLines={30} />

      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <div className="absolute top-2/5 left-2/11 w-[6rem] h-[6rem] bg-purple-500/10 rounded-full blur-[120px]"></div>

        <div className="absolute top-4/5 left-1/2 w-[10rem] h-[10rem] bg-purple-500/5 rounded-full blur-[180px]"></div>
        <div className="absolute -top-40 left-0 z-3 h-[16rem] w-[16rem] rounded-full bg-gradient-to-br from-blue-500/10 to-purple-500/10 blur-[60px] flex-none order-3 grow-0"></div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="text-center">
          <h2 className="text-3xl font-extrabold text-white sm:text-[2.4rem]">
            Ready to{" "}
            <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">
              Get Started?
            </span>
          </h2>
          <p className="mt-3 max-w-2xl mx-auto text-lg sm:text-xl text-blue-100">
            {t("subtitle")}
          </p>
        </div>

        {/* Installation Path Section - Two Cards Side by Side */}
        <div className="mt-12 mb-8 text-center">
          <h3 className="text-2xl sm:text-3xl font-bold text-white mb-3">
            Select Your{" "}
            <span className="text-gradient animated-gradient bg-gradient-to-r from-blue-400 via-purple-400 to-blue-400">
              Deployment Path
            </span>
          </h3>
          <p className="text-blue-100/90 text-base sm:text-lg max-w-2xl mx-auto">
            Start with local testing or deploy to production infrastructure
          </p>
        </div>

        <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-6 lg:gap-8 mb-12">
          {/* Local Development Installation Card */}
          <Link
            href="/quick-installation"
            className="group bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden transition-all duration-300 hover:border-blue-500/60 hover:shadow-xl hover:shadow-blue-500/10 hover:-translate-y-1"
          >
            <div className="p-8 flex flex-col h-full">
              {/* Header */}
              <div className="mb-6">
                <div className="inline-flex items-center justify-center w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500/20 to-blue-600/20 mb-4 group-hover:scale-110 transition-transform duration-300">
                  <svg
                    className="w-7 h-7 text-blue-400"
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
                <h3 className="text-xl font-bold text-white mb-1.5">
                  Local Development
                </h3>
                <p className="text-sm text-blue-300/80 font-medium">
                  Perfect for testing and learning
                </p>
              </div>

              {/* Description */}
              <p className="text-sm text-gray-300/90 mb-6 leading-relaxed">
                Get started quickly with a local Kubernetes environment using
                Docker and Kind. Ideal for development, testing, and exploring
                KubeStellar features.
              </p>

              {/* Features Grid */}
              <div className="grid grid-cols-2 gap-3 mb-6">
                <div className="flex items-center text-sm text-gray-300">
                  <svg
                    className="w-4 h-4 text-emerald-400 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>No cloud costs</span>
                </div>
                <div className="flex items-center text-sm text-gray-300">
                  <svg
                    className="w-4 h-4 text-emerald-400 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>15 min setup</span>
                </div>
                <div className="flex items-center text-sm text-gray-300">
                  <svg
                    className="w-4 h-4 text-emerald-400 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>Docker + Kind</span>
                </div>
                <div className="flex items-center text-sm text-gray-300">
                  <svg
                    className="w-4 h-4 text-emerald-400 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>K8s 1.34+</span>
                </div>
              </div>

              {/* CTA Button */}
              <div className="mt-auto pt-4 border-t border-slate-700/50">
                <div className="flex items-center justify-between text-blue-400 group-hover:text-blue-300 transition-colors">
                  <span className="text-sm font-semibold">
                    Start Local Installation
                  </span>
                  <svg
                    className="w-5 h-5 transition-transform duration-300 group-hover:translate-x-1"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 7l5 5m0 0l-5 5m5-5H6"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </Link>

          {/* AWS EKS Cloud Installation Card */}
          <Link
            href="/docs/getting-started/aws-eks"
            className="group bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden transition-all duration-300 hover:border-purple-500/60 hover:shadow-xl hover:shadow-purple-500/10 hover:-translate-y-1"
          >
            <div className="p-8 flex flex-col h-full">
              {/* Header */}
              <div className="mb-6">
                <div className="inline-flex items-center justify-center w-14 h-14 rounded-xl bg-gradient-to-br from-purple-500/20 to-purple-600/20 mb-4 group-hover:scale-110 transition-transform duration-300">
                  <svg
                    className="w-7 h-7 text-purple-400"
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
                <h3 className="text-xl font-bold text-white mb-1.5">
                  AWS EKS Production
                </h3>
                <p className="text-sm text-purple-300/80 font-medium">
                  Enterprise-ready deployment
                </p>
              </div>

              {/* Description */}
              <p className="text-sm text-gray-300/90 mb-6 leading-relaxed">
                Deploy KubeStellar on AWS EKS for production workloads with
                enterprise-grade scalability and reliability. Full cloud
                infrastructure automation included.
              </p>

              {/* Features Grid */}
              <div className="grid grid-cols-2 gap-3 mb-6">
                <div className="flex items-center text-sm text-gray-300">
                  <svg
                    className="w-4 h-4 text-emerald-400 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>EKS 1.34</span>
                </div>
                <div className="flex items-center text-sm text-gray-300">
                  <svg
                    className="w-4 h-4 text-emerald-400 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>30 min setup</span>
                </div>
                <div className="flex items-center text-sm text-gray-300">
                  <svg
                    className="w-4 h-4 text-emerald-400 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>Auto-scaling</span>
                </div>
                <div className="flex items-center text-sm text-gray-300">
                  <svg
                    className="w-4 h-4 text-emerald-400 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span>AWS Account</span>
                </div>
              </div>

              {/* CTA Button */}
              <div className="mt-auto pt-4 border-t border-slate-700/50">
                <div className="flex items-center justify-between text-purple-400 group-hover:text-purple-300 transition-colors">
                  <span className="text-sm font-semibold">
                    Start AWS Installation
                  </span>
                  <svg
                    className="w-5 h-5 transition-transform duration-300 group-hover:translate-x-1"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 7l5 5m0 0l-5 5m5-5H6"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </Link>
        </div>

        <div className="mt-12 grid grid-cols-1 lgcustom:grid-cols-2 gap-6 lg:gap-8">
          {/* Use Cases & Resources Card */}
          <div className="bg-slate-800/50 border border-slate-700 rounded-xl overflow-hidden transition-all duration-300 cursor-pointer hover:shadow-2xl hover:shadow-purple-500/30 hover:border-purple-500/50">
            <div className="p-6">
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
                <svg
                  className="h-6 w-6 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2 M13 7A4 4 0 1 1 5 7A4 4 0 0 1 13 7 M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"
                  />
                </svg>
              </div>
              <h3 className="text-base sm:text-lg font-bold mb-2 text-white">
                {t("card2Title")}
              </h3>
              <p className="text-sm sm:text-base text-gray-200 mb-6">
                {t("card2Description")}
              </p>
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <a
                    href={getLocalizedUrl("https://kubestellar.io/slack")}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center justify-center px-3 py-2.5 rounded-lg bg-gradient-to-r from-purple-600/80 to-purple-700/80 hover:from-purple-600 hover:to-purple-700 text-white text-sm font-medium transition-all duration-200 border border-purple-500/30"
                  >
                    <svg
                      className="h-4 w-4 mr-2 flex-shrink-0"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 60 60"
                      preserveAspectRatio="xMidYMid meet"
                    >
                      <path
                        d="M22,12 a6,6 0 1 1 6,-6 v6z M22,16 a6,6 0 0 1 0,12 h-16 a6,6 0 1 1 0,-12"
                        fill="#36C5F0"
                      ></path>
                      <path
                        d="M48,22 a6,6 0 1 1 6,6 h-6z M32,6 a6,6 0 1 1 12,0v16a6,6 0 0 1 -12,0z"
                        fill="#2EB67D"
                      ></path>
                      <path
                        d="M38,48 a6,6 0 1 1 -6,6 v-6z M54,32 a6,6 0 0 1 0,12 h-16 a6,6 0 1 1 0,-12"
                        fill="#ECB22E"
                      ></path>
                      <path
                        d="M12,38 a6,6 0 1 1 -6,-6 h6z M16,38 a6,6 0 1 1 12,0v16a6,6 0 0 1 -12,0z"
                        fill="#E01E5A"
                      ></path>
                    </svg>
                    {t("card2Button1")}
                  </a>
                  <a
                    href="https://github.com/kubestellar/kubestellar"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center justify-center px-3 py-2.5 rounded-lg bg-gradient-to-r from-slate-700/80 to-slate-800/80 hover:from-slate-700 hover:to-slate-800 text-white text-sm font-medium transition-all duration-200 border border-slate-600/30"
                  >
                    <svg
                      className="h-4 w-4 mr-2"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 24 24"
                      fill="currentColor"
                    >
                      <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"></path>
                    </svg>
                    {t("card2Button2")}
                  </a>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <Link
                    href="/products"
                    className="flex items-center justify-center px-3 py-2.5 rounded-lg bg-gradient-to-r from-orange-600/80 to-orange-700/80 hover:from-orange-600 hover:to-orange-700 text-white text-sm font-medium transition-all duration-200 border border-orange-500/30"
                  >
                    <svg
                      className="h-4 w-4 mr-2"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2H5a2 2 0 00-2 2v2M7 7h10"
                      ></path>
                    </svg>
                    {t("card2Button3")}
                  </Link>
                  <Link
                    href="/contribute-handbook"
                    className="flex items-center justify-center px-3 py-2.5 rounded-lg bg-gradient-to-r from-emerald-600/80 to-emerald-700/80 hover:from-emerald-600 hover:to-emerald-700 text-white text-sm font-medium transition-all duration-200 border border-emerald-500/30"
                  >
                    <svg
                      className="h-4 w-4 mr-2"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                      ></path>
                    </svg>
                    {t("card2Button4")}
                  </Link>
                </div>
              </div>
            </div>
          </div>

          {/* Documentation Card */}
          <div className="bg-slate-800/50 border border-slate-700 rounded-xl overflow-hidden transition-all duration-300 cursor-pointer hover:shadow-2xl hover:shadow-purple-500/30 hover:border-purple-500/50">
            <div className="p-6">
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
                <Icon path="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </div>
              <h3 className="text-base sm:text-lg font-bold mb-2">
                {t("card3Title")}
              </h3>
              <p className="text-sm sm:text-base text-blue-100 mb-4">
                {t("card3Description")}
              </p>
              <div className="mt-4 grid grid-cols-1 gap-2">
                <Link
                  href="/docs"
                  className="block p-2 rounded bg-white/20 hover:bg-white/30 text-white text-sm pl-5"
                >
                  {t("card3Link1")}
                </Link>
                <Link
                  href="/docs/what-is-kubestellar/architecture"
                  className="block p-2 rounded bg-white/20 hover:bg-white/30 text-white text-sm pl-5"
                >
                  {t("card3Link2")}
                </Link>
                <Link
                  href="docs/use-integrate/kubestellar-api/control"
                  className="block p-2 rounded bg-white/20 hover:bg-white/30 text-white text-sm pl-5"
                >
                  {t("card3Link3")}
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
