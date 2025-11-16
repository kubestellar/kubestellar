"use client";

import {
  GridLines,
  StarField,
  ContributionCallToAction,
  Navbar,
  Footer,
} from "../../../components/index";
import { useTranslations } from "next-intl";

export default function MaintainerLadderPage() {
  const t = useTranslations("ladderPage");

  const levels = [
    {
      id: 1,
      title: t("levels.contributor.title"),
      nextLevel: t("levels.contributor.nextLevel"),
      description: t("levels.contributor.description"),
      requirements: [
        t("levels.contributor.requirements.0"),
        t("levels.contributor.requirements.1"),
        t("levels.contributor.requirements.2"),
        t("levels.contributor.requirements.3"),
      ],
      icon: (
        <svg
          className="w-8 h-8"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
          />
        </svg>
      ),
      gradient: "from-blue-500 to-blue-600",
    },
    {
      id: 2,
      title: t("levels.unpaidIntern.title"),
      nextLevel: t("levels.unpaidIntern.nextLevel"),
      description: t("levels.unpaidIntern.description"),
      requirements: [
        t("levels.unpaidIntern.requirements.0"),
        t("levels.unpaidIntern.requirements.1"),
        t("levels.unpaidIntern.requirements.2"),
        t("levels.unpaidIntern.requirements.3"),
        t("levels.unpaidIntern.requirements.4"),
      ],
      timeframe: t("levels.unpaidIntern.timeframe"),
      icon: (
        <svg
          className="w-8 h-8"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
          />
        </svg>
      ),
      gradient: "from-purple-500 to-purple-600",
    },
    {
      id: 3,
      title: t("levels.paidIntern.title"),
      nextLevel: t("levels.paidIntern.nextLevel"),
      description: t("levels.paidIntern.description"),
      requirements: [
        t("levels.paidIntern.requirements.0"),
        t("levels.paidIntern.requirements.1"),
        t("levels.paidIntern.requirements.2"),
        t("levels.paidIntern.requirements.3"),
        t("levels.paidIntern.requirements.4"),
      ],
      icon: (
        <svg
          className="w-8 h-8"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"
          />
        </svg>
      ),
      gradient: "from-green-500 to-green-600",
    },
    {
      id: 4,
      title: t("levels.mentor.title"),
      nextLevel: t("levels.mentor.nextLevel"),
      description: t("levels.mentor.description"),
      requirements: [
        t("levels.mentor.requirements.0"),
        t("levels.mentor.requirements.1"),
        t("levels.mentor.requirements.2"),
        t("levels.mentor.requirements.3"),
      ],
      icon: (
        <svg
          className="w-8 h-8"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
          />
        </svg>
      ),
      gradient: "from-orange-500 to-orange-600",
    },
    {
      id: 5,
      title: t("levels.maintainer.title"),
      nextLevel: t("levels.maintainer.nextLevel"),
      description: t("levels.maintainer.description"),
      requirements: [
        t("levels.maintainer.requirements.0"),
        t("levels.maintainer.requirements.1"),
        t("levels.maintainer.requirements.2"),
        t("levels.maintainer.requirements.3"),
      ],
      icon: (
        <svg
          className="w-8 h-8"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"
          />
        </svg>
      ),
      gradient: "from-red-500 to-red-600",
    },
  ];

  return (
    <div className="bg-[#0a0a0a] text-white overflow-x-hidden min-h-screen">
      <Navbar />

      {/* Full page background with starfield */}
      <div className="fixed inset-0 z-0">
        {/* Dark base background */}
        <div className="absolute inset-0 bg-[#0a0a0a]"></div>

        {/* Starfield background */}
        <StarField density="medium" showComets={true} cometCount={3} />

        {/* Grid lines background */}
        <GridLines horizontalLines={21} verticalLines={18} />
      </div>

      <div className="relative z-10 pt-7">
        {" "}
        {/* Add padding-top to account for fixed navbar */}
        {/* Header Section */}
        <section className="pt-12 pb-12 sm:pt-28 sm:pb-16 lg:pt-24 lg:pb-6">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center">
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-3">
                {t("title")}{" "}
                <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">
                  {t("titleSpan")}
                </span>
              </h1>
              <p className="text-xl md:text-2xl text-gray-300 max-w-4xl mx-auto leading-relaxed">
                {t("subtitle")}
              </p>

              {/* Tracking Sheet CTA */}
              <div className="mt-8 flex justify-center">
                <div className="relative group">
                  <div className="absolute -inset-0.5 bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 rounded-lg blur opacity-60 group-hover:opacity-100 transition duration-300 animate-pulse"></div>
                  <a
                    href="http://kubestellar.io/ladder_stats"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="relative flex items-center gap-3 px-8 py-4 bg-gray-900 rounded-lg border border-blue-500/30 hover:border-blue-400/50 transition-all duration-300 group-hover:scale-105"
                  >
                    <div className="text-left">
                      <div className="text-sm text-gray-400 group-hover:text-gray-300 transition-colors">
                        How do we audit our contributor ladder?
                      </div>
                      <div className="text-lg font-semibold text-white flex items-center gap-2">
                        View Real-Time Statistics
                        <span className="text-blue-400 group-hover:translate-x-1 transition-transform inline-block">
                          →
                        </span>
                      </div>
                    </div>
                  </a>
                </div>
              </div>
            </div>
          </div>
        </section>
        {/* Ladder Section */}
        <section className="py-8 sm:py-12 md:py-16 lg:py-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
            {/* Mobile Layout */}
            <div className="lg:hidden">
              {levels.map((level, index) => (
                <div key={level.id} className="mb-4">
                  {/* Level Card */}
                  <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-6 border border-white/10 relative">
                    {/* Level Number */}
                    <div className="absolute -top-6 left-1/2 transform -translate-x-1/2">
                      <div
                        className={`w-12 h-12 bg-gradient-to-br ${level.gradient} rounded-full flex items-center justify-center shadow-lg`}
                      >
                        <span className="text-white font-bold text-lg">
                          {level.id}
                        </span>
                      </div>
                    </div>

                    <div className="pt-4">
                      <div className="text-center mb-4">
                        <div
                          className={`inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br ${level.gradient} rounded-full mb-3`}
                        >
                          <div className="text-white">{level.icon}</div>
                        </div>
                        <h3 className="text-xl font-bold text-white mb-2">
                          {level.title}
                        </h3>
                        {level.timeframe && (
                          <div className="inline-block bg-blue-900/50 rounded-full px-3 py-1 text-xs text-blue-200 mb-2">
                            {level.timeframe}
                          </div>
                        )}
                        <p className="text-gray-300 text-sm mb-4">
                          {level.description}
                        </p>
                      </div>

                      <div className="space-y-2">
                        <h4 className="text-sm font-semibold text-white mb-2">
                          {t("requirementsLabel")}
                        </h4>
                        <ul className="space-y-1">
                          {level.requirements.map((req, reqIndex) => (
                            <li
                              key={reqIndex}
                              className="text-xs text-gray-300 flex items-start"
                            >
                              <span className="text-green-400 mr-2 mt-1">
                                •
                              </span>
                              <span>{req}</span>
                            </li>
                          ))}
                        </ul>
                      </div>

                      {index < levels.length - 1 && (
                        <div className="text-center mt-4">
                          <div className="text-xs text-gray-400">
                            {t("nextLevelLabel")}
                          </div>
                          <div className="text-sm font-semibold text-blue-400">
                            {level.nextLevel}
                          </div>
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Connector */}
                  {index < levels.length - 1 && (
                    <div className="flex justify-center mt-2">
                      <div className="w-0.5 h-6 bg-gradient-to-b from-blue-500 to-purple-500"></div>
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* Desktop Layout */}
            <div className="hidden lg:block">
              {/* Central Ladder Line */}
              <div className="absolute left-1/2 top-0 bottom-0 w-1 bg-gradient-to-b from-blue-500 via-purple-500 to-red-500 transform -translate-x-1/2 z-5"></div>

              {levels.map((level, index) => (
                <div key={level.id} className="relative z-20">
                  <div
                    className={`flex items-center ${index % 2 === 0 ? "flex-row" : "flex-row-reverse"}`}
                  >
                    {/* Content Side */}
                    <div
                      className={`w-5/12 ${index % 2 === 0 ? "pr-12" : "pl-12"}`}
                    >
                      <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-8 border border-white/10 transition-all duration-300 hover:bg-gray-800/50 hover:border-white/20 hover:scale-105">
                        <div className="flex items-center mb-4">
                          <div
                            className={`w-12 h-12 bg-gradient-to-br ${level.gradient} rounded-full flex items-center justify-center mr-4`}
                          >
                            <div className="text-white">{level.icon}</div>
                          </div>
                          <div>
                            <h3 className="text-2xl font-bold text-white">
                              {level.title}
                            </h3>
                            {level.timeframe && (
                              <div className="inline-block bg-blue-900/50 rounded-full px-3 py-1 text-xs text-blue-200 mt-1">
                                {level.timeframe}
                              </div>
                            )}
                          </div>
                        </div>

                        <p className="text-gray-300 mb-6 leading-relaxed">
                          {level.description}
                        </p>

                        <div className="space-y-3">
                          <h4 className="text-lg font-semibold text-white">
                            {t("requirementsLabel")}
                          </h4>
                          <ul className="space-y-2">
                            {level.requirements.map((req, reqIndex) => (
                              <li
                                key={reqIndex}
                                className="text-gray-300 flex items-start"
                              >
                                <span className="text-green-400 mr-3 mt-1 text-lg">
                                  •
                                </span>
                                <span className="text-sm leading-relaxed">
                                  {req}
                                </span>
                              </li>
                            ))}
                          </ul>
                        </div>

                        {index < levels.length - 1 && (
                          <div className="mt-6 pt-4 border-t border-gray-700/50">
                            <div className="text-sm text-gray-400">
                              {t("nextLevelLabel")}
                            </div>
                            <div className="text-lg font-semibold text-blue-400">
                              {level.nextLevel}
                            </div>
                          </div>
                        )}
                      </div>
                    </div>

                    {/* Center Circle */}
                    <div className="w-2/12 flex justify-center">
                      <div
                        className={`w-20 h-20 bg-gradient-to-br ${level.gradient} rounded-full flex items-center justify-center shadow-2xl z-30 transition-all duration-300 hover:scale-110`}
                      >
                        <span className="text-white font-bold text-2xl">
                          {level.id}
                        </span>
                      </div>
                    </div>

                    {/* Empty Side */}
                    <div className="w-5/12"></div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>
        {/* Ready To Contribute Section */}
        <ContributionCallToAction />
      </div>
      <Footer />
    </div>
  );
}
