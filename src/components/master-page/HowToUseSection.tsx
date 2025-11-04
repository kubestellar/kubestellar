"use client";

import { GridLines, StarField } from "../index";
import { useTranslations } from "next-intl";

export default function HowToUseSection() {
  const t = useTranslations("howToUseSection");
  return (
    <section
      id="how-to-use"
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
              <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center shadow-lg">
                  <span className="text-white font-bold text-sm">1</span>
                </div>
              </div>

              <div className="pt-2">
                <h3 className="text-lg font-bold text-white mb-2 text-center">
                  {t("step1Title")}
                </h3>
                <p className="text-gray-300 text-sm leading-relaxed mb-3 text-center">
                  {t("step1Description")}
                </p>
                <div className="bg-slate-900/90 rounded-lg p-3 overflow-x-auto">
                  <pre className="text-xs font-mono text-white">
                    <code>
                      <span className="text-gray-400">
                        {t("step1CodeComment")}
                      </span>
                      {"\n"}
                      <span className="text-blue-400">
                        {t("step1Tool1")}
                      </span>,{" "}
                      <span className="text-blue-400">{t("step1Tool2")}</span>,{" "}
                      <span className="text-blue-400">{t("step1Tool3")}</span>,{" "}
                      <span className="text-blue-400">{t("step1Tool4")}</span>
                      {"\n"}
                      <span className="text-blue-400">{t("step1Tool5")}</span>
                      {"\n"}
                      <span className="text-blue-400">{t("step1Tool6")}</span>
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
              <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg">
                  <span className="text-white font-bold text-sm">2</span>
                </div>
              </div>

              <div className="pt-2">
                <h3 className="text-lg font-bold text-white mb-2 text-center">
                  {t("step2Title")}
                </h3>
                <p className="text-gray-300 text-sm leading-relaxed mb-3 text-center">
                  {t("step2Description")}
                </p>
                <div className="bg-slate-900/90 rounded-lg p-3 overflow-x-auto">
                  <pre className="text-xs font-mono text-white">
                    <code>
                      <span className="text-gray-400">
                        {t("step2CodeComment")}
                      </span>
                      {"\n"}
                      <span className="text-blue-400">
                        {t("step2Command")}
                      </span>{" "}
                      <span className="text-white">{t("step2Cluster")}</span> \
                      {"\n"}
                      <span className="text-emerald-400">
                        {t("step2Label1")}
                      </span>{" "}
                      \{"\n"}
                      <span className="text-emerald-400">
                        {t("step2Label2")}
                      </span>
                    </code>
                  </pre>
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
              <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                <div className="w-8 h-8 bg-gradient-to-br from-green-500 to-green-600 rounded-full flex items-center justify-center shadow-lg">
                  <span className="text-white font-bold text-sm">3</span>
                </div>
              </div>

              <div className="pt-2">
                <h3 className="text-lg font-bold text-white mb-2 text-center">
                  {t("step3Title")}
                </h3>
                <p className="text-gray-300 text-sm leading-relaxed mb-3 text-center">
                  {t("step3Description")}
                </p>
                <div className="bg-slate-900/90 rounded-lg p-3 overflow-x-auto">
                  <pre className="text-xs font-mono text-white">
                    <code>
                      <span className="text-yellow-300">apiVersion</span>:{" "}
                      <span className="text-white">{t("step3ApiVersion")}</span>
                      {"\n"}
                      <span className="text-yellow-300">kind</span>:{" "}
                      <span className="text-white">{t("step3Kind")}</span>
                      {"\n"}
                      <span className="text-yellow-300">spec</span>:{"\n"}
                      {"  "}
                      <span className="text-yellow-300">
                        {t("step3SpecClusterSelectors")}
                      </span>
                      :{"\n"}
                      {"  - "}
                      <span className="text-yellow-300">
                        {t("step3MatchLabels")}
                      </span>
                      :{"\n"}
                      {"      "}
                      <span className="text-emerald-400">
                        {t("step3LocationGroup")}
                      </span>
                    </code>
                  </pre>
                </div>
              </div>
            </div>
            {/* Mobile Connector */}
            <div className="flex justify-center mt-4">
              <div className="w-0.5 h-6 bg-gradient-to-b from-green-500 to-orange-500"></div>
            </div>
          </div>

          {/* Mobile Step 4 */}
          <div className="mb-8">
            <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-4 border border-white/10 relative">
              {/* Step Number at Top */}
              <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                <div className="w-8 h-8 bg-gradient-to-br from-orange-500 to-orange-600 rounded-full flex items-center justify-center shadow-lg">
                  <span className="text-white font-bold text-sm">4</span>
                </div>
              </div>

              <div className="pt-2">
                <h3 className="text-lg font-bold text-white mb-2 text-center">
                  {t("step4Title")}
                </h3>
                <p className="text-gray-300 text-sm leading-relaxed mb-3 text-center">
                  {t("step4Description")}
                </p>
                <div className="bg-slate-900/90 rounded-lg p-3 overflow-x-auto">
                  <pre className="text-xs font-mono text-white">
                    <code>
                      <span className="text-yellow-300">apiVersion</span>:{" "}
                      <span className="text-white">{t("step4ApiVersion")}</span>
                      {"\n"}
                      <span className="text-yellow-300">kind</span>:{" "}
                      <span className="text-white">{t("step4Kind")}</span>
                      {"\n"}
                      <span className="text-yellow-300">metadata</span>:{"\n"}
                      {"  "}
                      <span className="text-yellow-300">name</span>:{" "}
                      <span className="text-white">
                        {t("step4MetadataName")}
                      </span>
                      {"\n"}
                      {"  "}
                      <span className="text-yellow-300">
                        {t("step4Labels")}
                      </span>
                      :{"\n"}
                      {"    "}
                      <span className="text-emerald-400">
                        {t("step4AppName")}
                      </span>
                    </code>
                  </pre>
                </div>
              </div>
            </div>
            {/* Mobile Connector */}
            <div className="flex justify-center mt-4">
              <div className="w-0.5 h-6 bg-gradient-to-b from-orange-500 to-purple-500"></div>
            </div>
          </div>

          {/* Mobile Step 5 */}
          <div className="mb-8">
            <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-4 border border-white/10 relative">
              {/* Step Number at Top */}
              <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg">
                  <span className="text-white font-bold text-sm">5</span>
                </div>
              </div>

              <div className="pt-2">
                <h3 className="text-lg font-bold text-white mb-2 text-center">
                  {t("step5Title")}
                </h3>
                <p className="text-gray-300 text-sm leading-relaxed mb-3 text-center">
                  {t("step5Description")}
                </p>
                <div className="bg-slate-900/90 rounded-lg p-3 overflow-x-auto">
                  <pre className="text-xs font-mono text-white">
                    <code>
                      <span className="text-gray-400">
                        {t("step5Command1Comment")}
                      </span>
                      {"\n"}
                      <span className="text-emerald-400">
                        kubectl get pods -A
                      </span>
                      {"\n"}
                      {"\n"}
                      <span className="text-gray-400">
                        {t("step5Command2Comment")}
                      </span>
                      {"\n"}
                      <span className="text-emerald-400">
                        kubectl get deployments -A
                      </span>
                      {"\n"}
                      {"\n"}
                      <span className="text-gray-400">
                        {t("step5Command3Comment")}
                      </span>
                      {"\n"}
                      <span className="text-emerald-400">
                        kubectl describe deployment example-app
                      </span>
                    </code>
                  </pre>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Desktop Layout */}
        <div className="hidden lg:block relative z-10">
          {/* Connection Line */}
          <div className="absolute left-1/2 top-0 bottom-0 w-0.5 bg-gradient-to-b from-blue-500 via-purple-500 via-green-500 via-orange-500 to-purple-500 z-5 transform -translate-x-1/2 will-change-transform"></div>

          {/* Desktop Step 1 */}
          <div className="relative mb-4 lg:mb-6 z-20">
            <div className="flex flex-row items-center">
              <div className="w-1/2 pr-12">
                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 z-30 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20">
                  <h3 className="text-2xl font-bold text-white mb-4 flex items-center">
                    <span className="flex items-center justify-center w-8 h-8 rounded-full bg-blue-600 mr-3 text-white font-bold text-sm">
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
                        <span className="text-gray-400">
                          {t("step1CodeComment")}
                        </span>
                        {"\n"}
                        <span className="text-blue-400">{t("step1Tool1")}</span>
                        ,{" "}
                        <span className="text-blue-400">{t("step1Tool2")}</span>
                        ,{" "}
                        <span className="text-blue-400">{t("step1Tool3")}</span>
                        ,{" "}
                        <span className="text-blue-400">{t("step1Tool4")}</span>
                        {"\n"}
                        <span className="text-blue-400">{t("step1Tool5")}</span>
                        {"\n"}
                        <span className="text-blue-400">{t("step1Tool6")}</span>
                      </code>
                    </pre>
                  </div>
                </div>
              </div>
              <div className="w-1/2 pl-12"></div>
            </div>
          </div>

          {/* Desktop Step 2 */}
          <div className="relative mb-4 lg:mb-6 z-20 -mt-20">
            <div className="flex flex-row-reverse items-center">
              <div className="w-1/2 pl-12">
                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 z-30 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20">
                  <h3 className="text-2xl font-bold text-white mb-4 flex items-center">
                    <span className="flex items-center justify-center w-8 h-8 rounded-full bg-purple-600 mr-3 text-white font-bold text-sm">
                      2
                    </span>
                    {t("step2Title")}
                  </h3>
                  <p className="text-gray-300 mb-6 leading-relaxed">
                    {t("step2DescriptionDesktop")}
                  </p>
                  <div className="bg-slate-900/90 rounded-lg overflow-hidden shadow-lg w-full overflow-x-auto scrollbar-thin scrollbar-thumb-gray-600 scrollbar-track-gray-800">
                    <pre className="text-sm font-mono text-white p-4 leading-6 whitespace-pre-wrap">
                      <code>
                        <span className="text-gray-400">
                          {t("step2CodeComment")}
                        </span>
                        {"\n"}
                        <span className="text-blue-400">
                          {t("step2Command")}
                        </span>{" "}
                        <span className="text-white">{t("step2Cluster")}</span>{" "}
                        \{"\n"}
                        <span className="text-emerald-400">
                          {t("step2Label1")}
                        </span>{" "}
                        \{"\n"}
                        <span className="text-emerald-400">
                          {t("step2Label2")}
                        </span>
                      </code>
                    </pre>
                  </div>
                </div>
              </div>
              <div className="w-1/2 pr-12"></div>
            </div>
          </div>

          {/* Desktop Step 3 */}
          <div className="relative mb-4 lg:mb-6 z-20 -mt-20">
            <div className="flex flex-row items-center">
              <div className="w-1/2 pr-12">
                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 z-30 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20">
                  <h3 className="text-2xl font-bold text-white mb-4 flex items-center">
                    <span className="flex items-center justify-center w-8 h-8 rounded-full bg-green-600 mr-3 text-white font-bold text-sm">
                      3
                    </span>
                    {t("step3Title")}
                  </h3>
                  <p className="text-gray-300 mb-6 leading-relaxed">
                    {t("step3DescriptionDesktop")}
                  </p>
                  <div className="bg-slate-900/90 rounded-lg overflow-hidden shadow-lg w-full overflow-x-auto scrollbar-thin scrollbar-thumb-gray-600 scrollbar-track-gray-800">
                    <pre className="text-sm font-mono text-white p-4 leading-6 whitespace-pre-wrap">
                      <code>
                        <span className="text-yellow-300">apiVersion</span>:{" "}
                        <span className="text-white">
                          {t("step3ApiVersion")}
                        </span>
                        {"\n"}
                        <span className="text-yellow-300">kind</span>:{" "}
                        <span className="text-white">{t("step3Kind")}</span>
                        {"\n"}
                        <span className="text-yellow-300">metadata</span>:{"\n"}
                        {"  "}
                        <span className="text-yellow-300">name</span>:{" "}
                        <span className="text-white">
                          {t("step3MetadataName")}
                        </span>
                        {"\n"}
                        <span className="text-yellow-300">spec</span>:{"\n"}
                        {"  "}
                        <span className="text-yellow-300">
                          {t("step3SpecClusterSelectors")}
                        </span>
                        :{"\n"}
                        {"  - "}
                        <span className="text-yellow-300">
                          {t("step3MatchLabels")}
                        </span>
                        :{"\n"}
                        {"      "}
                        <span className="text-emerald-400">
                          {t("step3LocationGroup")}
                        </span>
                      </code>
                    </pre>
                  </div>
                </div>
              </div>
              <div className="w-1/2 pl-12"></div>
            </div>
          </div>

          {/* Desktop Step 4 */}
          <div className="relative mb-4 lg:mb-6 z-20 -mt-24">
            <div className="flex flex-row-reverse items-center">
              <div className="w-1/2 pl-12">
                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 z-30 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20">
                  <h3 className="text-2xl font-bold text-white mb-4 flex items-center">
                    <span className="flex items-center justify-center w-8 h-8 rounded-full bg-orange-600 mr-3 text-white font-bold text-sm">
                      4
                    </span>
                    {t("step4Title")}
                  </h3>
                  <p className="text-gray-300 mb-6 leading-relaxed">
                    {t("step4DescriptionDesktop")}
                  </p>
                  <div className="bg-slate-900/90 rounded-lg overflow-hidden shadow-lg w-full overflow-x-auto scrollbar-thin scrollbar-thumb-gray-600 scrollbar-track-gray-800">
                    <pre className="text-sm font-mono text-white p-4 leading-6 whitespace-pre-wrap">
                      <code>
                        <span className="text-yellow-300">apiVersion</span>:{" "}
                        <span className="text-white">
                          {t("step4ApiVersion")}
                        </span>
                        {"\n"}
                        <span className="text-yellow-300">kind</span>:{" "}
                        <span className="text-white">{t("step4Kind")}</span>
                        {"\n"}
                        <span className="text-yellow-300">metadata</span>:{"\n"}
                        {"  "}
                        <span className="text-yellow-300">name</span>:{" "}
                        <span className="text-white">
                          {t("step4MetadataName")}
                        </span>
                        {"\n"}
                        {"  "}
                        <span className="text-yellow-300">
                          {t("step4Labels")}
                        </span>
                        :{"\n"}
                        {"    "}
                        <span className="text-emerald-400">
                          {t("step4AppName")}
                        </span>
                        {"\n"}
                        <span className="text-yellow-300">
                          {t("step4Spec")}
                        </span>
                        :{"\n"}
                        {"  "}
                        <span className="text-yellow-300">
                          {t("step4Replicas")}
                        </span>
                        :{" "}
                        <span className="text-white">
                          {t("step4ReplicasValue")}
                        </span>
                      </code>
                    </pre>
                  </div>
                </div>
              </div>
              <div className="w-1/2 pr-12"></div>
            </div>
          </div>

          {/* Desktop Step 5 */}
          <div className="relative z-20 -mt-24">
            <div className="flex flex-row items-center">
              <div className="w-1/2 pr-12">
                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 z-30 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20">
                  <h3 className="text-2xl font-bold text-white mb-4 flex items-center">
                    <span className="flex items-center justify-center w-8 h-8 rounded-full bg-purple-600 mr-3 text-white font-bold text-sm">
                      5
                    </span>
                    {t("step5Title")}
                  </h3>
                  <p className="text-gray-300 mb-6 leading-relaxed">
                    {t("step5DescriptionDesktop")}
                  </p>
                  <div className="bg-slate-900/90 rounded-lg overflow-hidden shadow-lg w-full overflow-x-auto scrollbar-thin scrollbar-thumb-gray-600 scrollbar-track-gray-800">
                    <pre className="text-sm font-mono text-white p-4 leading-6 whitespace-pre-wrap">
                      <code>
                        <span className="text-gray-400">
                          {t("step5Command1Comment")}
                        </span>
                        {"\n"}
                        <span className="text-emerald-400">
                          kubectl get pods -A
                        </span>
                        {"\n"}
                        {"\n"}
                        <span className="text-gray-400">
                          {t("step5Command2Comment")}
                        </span>
                        {"\n"}
                        <span className="text-emerald-400">
                          kubectl get deployments -A
                        </span>
                        {"\n"}
                        {"\n"}
                        <span className="text-gray-400">
                          {t("step5Command3Comment")}
                        </span>
                        {"\n"}
                        <span className="text-emerald-400">
                          kubectl describe deployment example-app
                        </span>
                      </code>
                    </pre>
                  </div>
                </div>
              </div>
              <div className="w-1/2 pl-12"></div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
