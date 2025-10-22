"use client";

import { GridLines, StarField } from "../index";
import { useTranslations } from "next-intl";

export default function HowItWorksSection() {
  const t = useTranslations("howItWorksSection");
  return (
    <section
      id="how-it-works"
      className="relative py-8 sm:py-12 md:py-16 lg:py-20 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden will-change-transform"
    >
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]"></div>

      {/* Starfield background */}
      <StarField density="medium" showComets={true} cometCount={3} />

      {/* Grid lines background */}
      <GridLines horizontalLines={21} verticalLines={18} />

      <div className="absolute right-0 top-0 h-full w-1/2 bg-gradient-to-l from-blue-500/10 to-transparent"></div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="text-center mb-8 sm:mb-12 md:mb-16">
          <h2 className="text-3xl font-extrabold text-white sm:text-[2.4rem]">
            {t("title")}{" "}
            <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">
              {t("titleSpan")}
            </span>
          </h2>
          <p className="mt-3 sm:mt-4 max-w-2xl mx-auto text-base sm:text-lg md:text-xl text-gray-300 px-4">
            {t("subtitle")}
          </p>
        </div>

        {/* Mobile Steps Layout */}
        <div className="lg:hidden relative z-10">
          {/* Mobile Step 1 */}
          <div className="mb-8">
            <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-4 border border-white/10 relative">
              {/* Step Number at Top */}
              <div className="absolute -top-6 left-1/2 transform -translate-x-1/2">
                <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center shadow-lg">
                  <span className="text-white font-bold text-lg">1</span>
                </div>
              </div>

              <div className="pt-4">
                <h3 className="text-lg font-bold text-white mb-2 text-center">
                  {t("step1Title")}
                </h3>
                <p className="text-gray-300 text-sm leading-relaxed mb-3 text-center">
                  {t("step1Description")}
                </p>
                <div className="bg-slate-900/90 rounded-lg p-3 overflow-x-auto">
                  <pre className="text-xs font-mono text-white">
                    <code>
                      <span className="text-yellow-300">apiVersion</span>:{" "}
                      <span className="text-white">apps/v1</span>
                      {"\n"}
                      <span className="text-yellow-300">kind</span>:{" "}
                      <span className="text-white">Deployment</span>
                      {"\n"}
                      <span className="text-yellow-300">metadata</span>:{"\n"}{" "}
                      <span className="text-yellow-300">name</span>:{" "}
                      <span className="text-white">example-app</span>
                      {"\n"}{" "}
                      <span className="text-yellow-300">annotations</span>:
                      {"\n"}{" "}
                      <span className="text-yellow-300">
                        kubestellar.io/placement
                      </span>
                      :{"\n"}{" "}
                      <span className="text-emerald-400">
                        &quot;region=us-east,tier=prod&quot;
                      </span>
                    </code>
                  </pre>
                </div>
              </div>
            </div>
            {/* Mobile Connector */}
            <div className="flex justify-center mt-4">
              <div className="w-0.5 h-6 bg-gradient-to-b from-blue-500 to-purple-500"></div>
            </div>
          </div>

          {/* Mobile Step 2 */}
          <div className="mb-8">
            <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-4 border border-white/10 relative">
              {/* Step Number at Top */}
              <div className="absolute -top-6 left-1/2 transform -translate-x-1/2">
                <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg">
                  <span className="text-white font-bold text-lg">2</span>
                </div>
              </div>

              <div className="pt-4">
                <h3 className="text-lg font-bold text-white mb-2 text-center">
                  {t("step2Title")}
                </h3>
                <p className="text-gray-300 text-sm leading-relaxed mb-3 text-center">
                  {t("step2Description")}
                </p>
                <div className="flex flex-wrap gap-2 justify-center">
                  <div className="bg-blue-900/80 backdrop-blur-lg rounded-full px-3 py-1 text-white text-xs">
                    {t("step2Tag1")}
                  </div>
                  <div className="bg-purple-900/80 backdrop-blur-lg rounded-full px-3 py-1 text-white text-xs">
                    {t("step2Tag2")}
                  </div>
                  <div className="bg-green-900/80 backdrop-blur-lg rounded-full px-3 py-1 text-white text-xs">
                    {t("step2Tag3")}
                  </div>
                </div>
              </div>
            </div>
            {/* Mobile Connector */}
            <div className="flex justify-center mt-4">
              <div className="w-0.5 h-6 bg-gradient-to-b from-purple-500 to-green-500"></div>
            </div>
          </div>

          {/* Mobile Step 3 */}
          <div className="mb-8">
            <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-4 border border-white/10 relative">
              {/* Step Number at Top */}
              <div className="absolute -top-6 left-1/2 transform -translate-x-1/2">
                <div className="w-12 h-12 bg-gradient-to-br from-green-500 to-green-600 rounded-full flex items-center justify-center shadow-lg">
                  <span className="text-white font-bold text-lg">3</span>
                </div>
              </div>

              <div className="pt-4">
                <h3 className="text-lg font-bold text-white mb-2 text-center">
                  {t("step3Title")}
                </h3>
                <p className="text-gray-300 text-sm leading-relaxed mb-3 text-center">
                  {t("step3Description")}
                </p>
                <div className="flex justify-center">
                  <div className="flex items-center space-x-4">
                    <div className="flex flex-col items-center space-y-2">
                      <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
                      <div className="w-3 h-3 bg-purple-500 rounded-full"></div>
                      <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                    </div>
                    <div className="flex flex-col space-y-2">
                      <span className="text-white text-xs">
                        {t("step3Cluster1")}
                      </span>
                      <span className="text-white text-xs">
                        {t("step3Cluster2")}
                      </span>
                      <span className="text-white text-xs">
                        {t("step3Cluster3")}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Desktop Layout */}
        <div className="hidden lg:block relative z-10">
          {/* Connection Line */}
          <div className="absolute left-1/2 top-0 bottom-0 w-0.5 bg-gradient-to-b from-blue-500 to-purple-600 z-5 transform -translate-x-1/2 will-change-transform"></div>

          {/* Desktop Step 1 */}
          <div className="relative mb-16 lg:mb-20 z-20">
            <div className="flex flex-row items-center">
              <div className="w-1/2 pr-12">
                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 z-30 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20">
                  <h3 className="text-2xl font-bold text-white mb-4 flex items-center">
                    <span className="flex items-center justify-center w-10 h-10 rounded-full bg-blue-600 mr-3 text-white font-bold text-base">
                      1
                    </span>
                    {t("step1Title")}
                  </h3>
                  <p className="text-gray-300 mb-6 leading-relaxed">
                    {t("step1DescriptionDesktop")}
                  </p>
                  <div className="bg-slate-900/90 rounded-lg overflow-hidden shadow-lg w-full overflow-x-auto scrollbar-thin scrollbar-thumb-gray-600 scrollbar-track-gray-800">
                    <pre className="text-sm font-mono text-white p-4 leading-6 whitespace-pre-wrap">
                      <code>
                        <span className="text-yellow-300">apiVersion</span>:{" "}
                        <span className="text-white">apps/v1</span>
                        {"\n"}
                        <span className="text-yellow-300">kind</span>:{" "}
                        <span className="text-white">Deployment</span>
                        {"\n"}
                        <span className="text-yellow-300">metadata</span>:{"\n"}{" "}
                        <span className="text-yellow-300">name</span>:{" "}
                        <span className="text-white">example-app</span>
                        {"\n"}{" "}
                        <span className="text-yellow-300">annotations</span>:
                        {"\n"}{" "}
                        <span className="text-yellow-300">
                          kubestellar.io/placement
                        </span>
                        :{" "}
                        <span className="text-emerald-400">
                          &quot;region=us-east,tier=prod&quot;
                        </span>
                      </code>
                    </pre>
                  </div>
                </div>
              </div>

              <div className="flex justify-center w-1/2 pl-12">
                <div className="relative w-24 h-24 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg z-30 transition-all duration-300 hover:scale-105 hover:shadow-xl">
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

          {/* Desktop Step 2 */}
          <div className="relative mb-16 lg:mb-20 z-20">
            <div className="flex flex-row-reverse items-center">
              <div className="w-1/2 pl-12">
                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 z-30 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20">
                  <h3 className="text-2xl font-bold text-white mb-4 flex items-center">
                    <span className="flex items-center justify-center w-10 h-10 rounded-full bg-blue-600 mr-3 text-white font-bold text-base">
                      2
                    </span>
                    {t("step2Title")}
                  </h3>

                  <p className="text-gray-300 mb-6 leading-relaxed">
                    {t("step2DescriptionDesktop")}
                  </p>
                  <div className="grid grid-cols-3 gap-4">
                    <div className="bg-blue-900/80 backdrop-blur-lg rounded-full px-3 py-2 text-white text-sm flex items-center justify-center transition-all duration-300 hover:bg-blue-900/90 hover:scale-105">
                      <div className="text-sm opacity-70 font-semibold text-center">
                        {t("step2Tag1")}
                      </div>
                    </div>
                    <div className="bg-purple-900/80 backdrop-blur-lg rounded-full px-3 py-2 text-white text-sm flex items-center justify-center transition-all duration-300 hover:bg-purple-900/90 hover:scale-105">
                      <div className="text-sm opacity-70 font-semibold text-center">
                        {t("step2Tag2")}
                      </div>
                    </div>
                    <div className="bg-green-900/80 backdrop-blur-lg rounded-full px-3 py-2 text-white text-sm flex items-center justify-center transition-all duration-300 hover:bg-green-900/90 hover:scale-105">
                      <div className="text-sm opacity-70 font-semibold text-center">
                        {t("step2Tag3")}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex justify-center w-1/2 pr-12">
                <div className="relative w-24 h-24 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg z-30 transition-all duration-300 hover:scale-105 hover:shadow-xl">
                  <svg
                    className="w-12 h-12 text-white"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1}
                      d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                    />
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1}
                      d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          {/* Desktop Step 3 */}
          <div className="relative z-20">
            <div className="flex flex-row items-center">
              <div className="w-1/2 pr-12">
                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 z-30 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20">
                  <h3 className="text-2xl font-bold text-white mb-4 flex items-center">
                    <span className="flex items-center justify-center w-10 h-10 rounded-full bg-blue-600 mr-3 text-white font-bold text-base">
                      3
                    </span>
                    {t("step3Title")}
                  </h3>

                  <p className="text-gray-300 mb-6 leading-relaxed">
                    {t("step3DescriptionDesktop")}
                  </p>
                  <div className="flex items-center justify-center space-x-4">
                    <div className="bg-blue-900/40 backdrop-blur-lg px-3 py-2 text-white text-sm flex flex-col items-center justify-center w-40 rounded-lg transition-all duration-300 hover:bg-blue-900/50 hover:scale-105">
                      <span className="text-sm opacity-50 text-center">
                        {t("step3Cluster1")}
                      </span>
                      <div className="w-full h-1 bg-blue-500 mt-2 rounded"></div>
                    </div>

                    <div className="bg-purple-900/40 backdrop-blur-lg px-3 py-2 text-white text-sm flex flex-col items-center justify-center w-40 rounded-lg transition-all duration-300 hover:bg-purple-900/50 hover:scale-105">
                      <span className="text-sm opacity-50 text-center">
                        {t("step3Cluster2")}
                      </span>
                      <div className="w-full h-1 bg-purple-500 mt-2 rounded"></div>
                    </div>

                    <div className="bg-green-900/40 backdrop-blur-lg px-3 py-2 text-white text-sm flex flex-col items-center justify-center w-40 rounded-lg transition-all duration-300 hover:bg-green-900/50 hover:scale-105">
                      <span className="text-sm opacity-50 text-center">
                        {t("step3Cluster3")}
                      </span>
                      <div className="w-full h-1 bg-green-500 mt-2 rounded"></div>
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex justify-center w-1/2 pl-12">
                <div className="relative w-24 h-24 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg z-30 transition-all duration-300 hover:scale-105 hover:shadow-xl">
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
                      d="M7 16a4 4 0 01-.88-7.906A6 6 0 1118 14h-1m-5 6V10m0 10l-3-3m3 3l3-3"
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
