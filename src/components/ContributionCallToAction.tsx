"use client";

import { Link as IntlLink } from "@/i18n/navigation";
import { GridLines, StarField } from "./index";
import { useTranslations } from "next-intl";
import { getLocalizedUrl } from "@/lib/url";

export default function ContributionCallToAction() {
  const t = useTranslations("ladderPage.callToAction");

  return (
    <section className="relative py-16 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden">
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
          <h2 className="text-3xl font-extrabold sm:text-[2.4rem] mb-4">
            {t("title")}
          </h2>
          <p className="mt-3 max-w-2xl mx-auto text-lg sm:text-xl text-blue-100 mb-12">
            {t("subtitle")}
          </p>
        </div>

        <div className="mt-8 flex flex-col sm:flex-row gap-6 justify-center items-center">
          {/* Community Meetings Button */}
          <a
            href="https://github.com/kubestellar/kubestellar/wiki"
            target="_blank"
            rel="noopener noreferrer"
            className="group relative overflow-hidden inline-flex items-center justify-center px-8 py-4 text-base sm:text-lg font-bold rounded-xl text-white 
                      bg-gradient-to-r from-blue-600 via-purple-600 to-indigo-600 
                      hover:from-blue-700 hover:via-purple-700 hover:to-indigo-700 
                      transition-all duration-500 transform hover:-translate-y-1 
                      hover:shadow-xl hover:shadow-blue-500/40 
                      animate-btn-float min-w-[200px]"
          >
            <svg
              className="mr-3 h-5 w-5 transition-all duration-300 group-hover:rotate-12"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
            <span className="relative z-10">
              {t("communityMeetingsButton")}
            </span>
            <div className="btn-shine"></div>
          </a>

          {/* View Open Issues Button */}
          <a
            href="https://github.com/kubestellar/kubestellar/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center px-8 py-4 text-base sm:text-lg font-bold rounded-xl text-gray-200 
                      bg-gray-800/40 hover:bg-gray-800/60 
                      backdrop-blur-md border border-gray-700/50 hover:border-gray-600/50 
                      transition-all duration-500 transform hover:-translate-y-1 hover:scale-105 
                      animate-btn-float min-w-[200px]"
            style={{ animationDelay: "0.1s" }}
          >
            <svg
              className="mr-3 h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            {t("viewIssuesButton")}
          </a>
        </div>

        {/* Additional Quick Links */}
        <div className="mt-12 grid grid-cols-1 sm:grid-cols-3 gap-6 max-w-4xl mx-auto">
          {/* GitHub Repository Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-lg p-6 border border-white/20 hover:bg-white/20 transition-colors duration-300">
            <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4 mx-auto">
              <svg
                className="h-6 w-6 text-white"
                fill="currentColor"
                viewBox="0 0 24 24"
              >
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"></path>
              </svg>
            </div>
            <h3 className="text-lg font-bold mb-2 text-center">
              {t("exploreCodeTitle")}
            </h3>
            <p className="text-sm text-blue-100 text-center mb-4">
              {t("exploreCodeDescription")}
            </p>
            <a
              href="https://github.com/kubestellar/kubestellar"
              target="_blank"
              rel="noopener noreferrer"
              className="block text-center text-blue-300 hover:text-blue-200 font-medium"
            >
              {t("viewRepositoryLink")}
            </a>
          </div>

          {/* Community Slack Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-lg p-6 border border-white/20 hover:bg-white/20 transition-colors duration-300">
            <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4 mx-auto">
              <svg
                className="h-6 w-6"
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
            </div>
            <h3 className="text-lg font-bold mb-2 text-center">
              {t("joinSlackTitle")}
            </h3>
            <p className="text-sm text-blue-100 text-center mb-4">
              {t("joinSlackDescription")}
            </p>
            <a
              href={getLocalizedUrl("https://kubestellar.io/slack")}
              target="_blank"
              rel="noopener noreferrer"
              className="block text-center text-blue-300 hover:text-blue-200 font-medium"
            >
              {t("joinCommunityLink")}
            </a>
          </div>

          {/* Contribution Guide Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-lg p-6 border border-white/20 hover:bg-white/20 transition-colors duration-300">
            <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4 mx-auto">
              <svg
                className="h-6 w-6 text-white"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                />
              </svg>
            </div>
            <h3 className="text-lg font-bold mb-2 text-center">
              {t("learnGuideTitle")}
            </h3>
            <p className="text-sm text-blue-100 text-center mb-4">
              {t("learnGuideDescription")}
            </p>
            <IntlLink
              href="/contribute-handbook"
              className="block text-center text-blue-300 hover:text-blue-200 font-medium"
            >
              {t("viewHandbookLink")}
            </IntlLink>
          </div>
        </div>
      </div>
    </section>
  );
}
