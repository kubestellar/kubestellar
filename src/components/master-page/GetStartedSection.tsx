"use client";

import Link from "next/link";
import { GridLines, StarField } from "../index";
import { useTranslations } from "next-intl";

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
          <h2 className="text-3xl font-extrabold sm:text-[2.4rem]">
            {t("title")}
          </h2>
          <p className="mt-3 max-w-2xl mx-auto text-lg sm:text-xl text-blue-100">
            {t("subtitle")}
          </p>
        </div>

        <div className="mt-12 grid grid-cols-1 lgcustom:grid-cols-3 gap-6 lg:gap-8">
          {/* Installation Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-xl overflow-hidden border border-white/20 hover:bg-white/20 transition-colors duration-300 w-11/12 max-w-xl mx-auto lgcustom:w-auto lgcustom:max-w-none">
            <div className="p-6 flex flex-col h-full">
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
                <Icon path="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
              </div>
              <h3 className="text-base sm:text-lg font-bold mb-2">
                {t("card1Title")}
              </h3>
              <p className="text-sm sm:text-base text-blue-100 mb-4">
                {t("card1Description")}
              </p>
              <div>
                <Link
                  href="/quick-installation"
                  className="inline-flex items-center justify-center w-full px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 sm:px-6 sm:py-3 sm:text-base transition-colors duration-20"
                >
                  {t("card1Button")}
                </Link>
              </div>
            </div>
          </div>

          {/* Community Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-xl overflow-hidden border border-white/20 hover:bg-white/20 transition-colors duration-300 w-11/12 max-w-xl mx-auto lgcustom:w-auto lgcustom:max-w-none">
            <div className="p-6">
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
                <Icon path="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2 M13 7A4 4 0 1 1 5 7A4 4 0 0 1 13 7 M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
              </div>
              <h3 className="text-base sm:text-lg font-bold mb-2">
                {t("card2Title")}
              </h3>
              <p className="text-sm sm:text-base text-blue-100 mb-4">
                {t("card2Description")}
              </p>
              <div className="mt-4 grid grid-cols-2 gap-2">
                <a
                  href="https://kubestellar.io/slack"
                  target="_blank"
                  className="flex items-center justify-center p-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-sm"
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
                  className="flex items-center justify-center p-2 rounded-lg bg-gray-700 hover:bg-gray-800 text-white text-sm"
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
            </div>
          </div>

          {/* Documentation Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-xl overflow-hidden border border-white/20 hover:bg-white/20 transition-colors duration-300 w-11/12 max-w-xl mx-auto lgcustom:w-auto lgcustom:max-w-none">
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
                  href="#"
                  className="block p-2 rounded bg-white/20 hover:bg-white/30 text-white text-sm pl-5"
                >
                  {t("card3Link1")}
                </Link>
                <Link
                  href="#"
                  className="block p-2 rounded bg-white/20 hover:bg-white/30 text-white text-sm pl-5"
                >
                  {t("card3Link2")}
                </Link>
                <Link
                  href="#"
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
