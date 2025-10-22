"use client";

import { GridLines, StarField } from "../../../components/index";
import Navigation from "../../../components/Navigation";
import Footer from "../../../components/Footer";

export default function MaintainerLadderPage() {
  const levels = [
    {
      id: 1,
      title: "Contributor",
      nextLevel: "Unpaid Intern",
      description: "Start your journey in the KubeStellar community",
      requirements: [
        "Minimum of 3 contributions (bug reports, documentation, or code PRs)",
        "Display enthusiasm and interest in long-term participation",
        "Be active on GitHub and Slack",
        "Informal application or nomination to join the intern program",
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
      title: "Unpaid Intern",
      nextLevel: "Paid Intern",
      description: "12-week journey to demonstrate commitment and skill",
      requirements: [
        "Open at least 6 'help wanted' issues",
        "Merge at least 20 PRs (8 within first 6 weeks)",
        "Attend weekly team meetings or submit summaries",
        "Work collaboratively with mentors",
        "Receive mentor's recommendation",
      ],
      timeframe: "12 weeks",
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
      title: "Paid Intern",
      nextLevel: "Mentor",
      description:
        "Recognized contributor with compensation and responsibility",
      requirements: [
        "Successfully complete at least one 12-week paid internship cycle",
        "Help onboard and support at least one new intern or contributor",
        "Submit ≥3 PR reviews",
        "Submit ≥5 helpful comments on PRs or issues",
        "Present or co-present at a community call",
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
      title: "Mentor",
      nextLevel: "Maintainer",
      description: "Guide and support the next generation of contributors",
      requirements: [
        "Demonstrate technical leadership in one or more key areas",
        "Maintain consistent contribution activity",
        "Engage with the community in GitHub and Slack",
        "Approved by core maintainers following a public review process",
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
      title: "Maintainer",
      nextLevel: "Core Team",
      description: "Trusted leader with full project responsibilities",
      requirements: [
        "≥2 'Help Wanted' issues every 2 months",
        "≥3 PRs merged every 2 months",
        "≥8 PR reviews or constructive comments every 2 months",
        "≥3 community meeting attendances every 2 months",
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
      <Navigation />

      {/* Full page background with starfield */}
      <div className="fixed inset-0 z-0">
        {/* Dark base background */}
        <div className="absolute inset-0 bg-[#0a0a0a]"></div>

        {/* Starfield background */}
        <StarField density="medium" showComets={true} cometCount={3} />

        {/* Grid lines background */}
        <GridLines horizontalLines={21} verticalLines={18} />
      </div>

      <div className="relative z-10 pt-16">
        {" "}
        {/* Add padding-top to account for fixed navbar */}
        {/* Header Section */}
        <section className="py-16 sm:py-20 lg:py-24">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center">
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-6">
                Contribution{" "}
                <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">
                  Ladder
                </span>
              </h1>
              <p className="text-xl md:text-2xl text-gray-300 max-w-4xl mx-auto leading-relaxed">
                A transparent, merit-based path from first-time contributor to
                trusted maintainer in the KubeStellar community
              </p>
            </div>
          </div>
        </section>
        {/* Ladder Section */}
        <section className="py-8 sm:py-12 md:py-16 lg:py-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
            {/* Mobile Layout */}
            <div className="lg:hidden">
              {levels.map((level, index) => (
                <div key={level.id} className="mb-8">
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
                          Requirements:
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
                            Next Level:
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
                    <div className="flex justify-center mt-4">
                      <div className="w-0.5 h-8 bg-gradient-to-b from-blue-500 to-purple-500"></div>
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
                <div key={level.id} className="relative mb-24 z-20">
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
                            Requirements:
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
                              Next Level:
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
        {/* Maintainer Activity Requirements */}
        <section className="py-16 sm:py-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-8 border border-white/10">
              <h2 className="text-3xl font-bold text-white mb-8 text-center">
                Maintainer Activity Requirements
              </h2>
              <p className="text-gray-300 text-center mb-8 text-lg">
                Maintainers must meet these bi-monthly (every 2 months)
                contribution minimums:
              </p>

              <div className="overflow-x-auto">
                <table className="w-full text-left">
                  <thead>
                    <tr className="border-b border-gray-600">
                      <th className="py-4 px-6 text-white font-semibold">
                        Metric
                      </th>
                      <th className="py-4 px-6 text-white font-semibold">
                        Requirement (Per 2 Months)
                      </th>
                    </tr>
                  </thead>
                  <tbody className="text-gray-300">
                    <tr className="border-b border-gray-700/50">
                      <td className="py-4 px-6">&quot;Help Wanted&quot; Issues</td>
                      <td className="py-4 px-6 text-green-400 font-semibold">
                        ≥ 2
                      </td>
                    </tr>
                    <tr className="border-b border-gray-700/50">
                      <td className="py-4 px-6 font-semibold">PRs Merged</td>
                      <td className="py-4 px-6 text-green-400 font-semibold">
                        ≥ 3
                      </td>
                    </tr>
                    <tr className="border-b border-gray-700/50">
                      <td className="py-4 px-6">
                        PR Reviews or Constructive Comments
                      </td>
                      <td className="py-4 px-6 text-green-400 font-semibold">
                        ≥ 8
                      </td>
                    </tr>
                    <tr>
                      <td className="py-4 px-6">
                        Community Meeting Attendance
                      </td>
                      <td className="py-4 px-6 text-green-400 font-semibold">
                        ≥ 3
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>
        {/* Call to Action */}
        <section className="py-16 sm:py-20">
          <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
            <h2 className="text-3xl font-bold text-white mb-6">
              Ready to Start Your Journey?
            </h2>
            <p className="text-xl text-gray-300 mb-8 leading-relaxed">
              Join the KubeStellar community and begin climbing the maintainer
              ladder today
            </p>
            <div className="space-y-4 sm:space-y-0 sm:space-x-4 sm:flex sm:justify-center">
              <a
                href="https://cloud-native.slack.com/archives/C097094RZ3M"
                className="inline-block bg-gradient-to-r from-blue-600 to-purple-600 text-white font-semibold py-3 px-8 rounded-lg transition-all duration-300 hover:from-blue-700 hover:to-purple-700 hover:scale-105"
              >
                Community Meetings
              </a>
              <a
                href="https://github.com/kubestellar/kubestellar/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22"
                className="inline-block bg-gray-700 text-white font-semibold py-3 px-8 rounded-lg border border-gray-600 transition-all duration-300 hover:bg-gray-600 hover:scale-105"
              >
                View Open Issues
              </a>
            </div>
          </div>
        </section>
      </div>
      <Footer />
    </div>
  );
}
