"use client";

import { GridLines, StarField } from "../index";
import { useTranslations } from "next-intl";

export default function UseCasesSection() {
  const t = useTranslations("useCasesSection");
  const getIcon = (iconType: string) => {
    switch (iconType) {
      case "globe":
        return (
          <svg
            width="24"
            height="25"
            viewBox="0 0 24 25"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M3.055 11.5H5C5.53043 11.5 6.03914 11.7107 6.41421 12.0858C6.78929 12.4609 7 12.9696 7 13.5V14.5C7 15.0304 7.21071 15.5391 7.58579 15.9142C7.96086 16.2893 8.46957 16.5 9 16.5C9.53043 16.5 10.0391 16.7107 10.4142 17.0858C10.7893 17.4609 11 17.9696 11 18.5V21.445M8 4.435V6C8 6.66304 8.26339 7.29893 8.73223 7.76777C9.20107 8.23661 9.83696 8.5 10.5 8.5H11C11.5304 8.5 12.0391 8.71071 12.4142 9.08579C12.7893 9.46086 13 9.96957 13 10.5C13 11.0304 13.2107 11.5391 13.5858 11.9142C13.9609 12.2893 14.4696 12.5 15 12.5C15.5304 12.5 16.0391 12.2893 16.4142 11.9142C16.7893 11.5391 17 11.0304 17 10.5C17 9.96957 17.2107 9.46086 17.5858 9.08579C17.9609 8.71071 18.4696 8.5 19 8.5H20.064M15 20.988V18.5C15 17.9696 15.2107 17.4609 15.5858 17.0858C15.9609 16.7107 16.4696 16.5 17 16.5H20.064M21 12.5C21 13.6819 20.7672 14.8522 20.3149 15.9442C19.8626 17.0361 19.1997 18.0282 18.364 18.864C17.5282 19.6997 16.5361 20.3626 15.4442 20.8149C14.3522 21.2672 13.1819 21.5 12 21.5C10.8181 21.5 9.64778 21.2672 8.55585 20.8149C7.46392 20.3626 6.47177 19.6997 5.63604 18.864C4.80031 18.0282 4.13738 17.0361 3.68508 15.9442C3.23279 14.8522 3 13.6819 3 12.5C3 10.1131 3.94821 7.82387 5.63604 6.13604C7.32387 4.44821 9.61305 3.5 12 3.5C14.3869 3.5 16.6761 4.44821 18.364 6.13604C20.0518 7.82387 21 10.1131 21 12.5Z"
              stroke="#60A5FA"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        );
      case "security":
        return (
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M9 12.0003L11 14.0003L15 10.0003M20.618 5.98434C17.4561 6.15225 14.3567 5.05895 12 2.94434C9.64327 5.05895 6.5439 6.15225 3.382 5.98434C3.12754 6.96945 2.99918 7.98289 3 9.00034C3 14.5913 6.824 19.2903 12 20.6223C17.176 19.2903 21 14.5923 21 9.00034C21 7.95834 20.867 6.94834 20.618 5.98434Z"
              stroke="#C084FC"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        );
      case "cloud":
        return (
          <svg
            width="24"
            height="25"
            viewBox="0 0 24 25"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M3 15.75C3 16.8108 3.42143 17.8283 4.17157 18.5784C4.92172 19.3285 5.93913 19.75 7 19.75H16C16.6608 19.75 17.3151 19.619 17.925 19.3646C18.5348 19.1102 19.0882 18.7374 19.5532 18.2678C20.0181 17.7982 20.3853 17.2411 20.6336 16.6287C20.8819 16.0163 21.0064 15.3608 20.9997 14.7C20.9931 14.0392 20.8556 13.3863 20.5951 12.779C20.3346 12.1717 19.9563 11.622 19.4821 11.1618C19.0079 10.7016 18.4471 10.34 17.8323 10.0978C17.2174 9.85564 16.5607 9.73776 15.9 9.75098C15.7734 9.09802 15.5178 8.47687 15.1483 7.92388C14.7787 7.37088 14.3026 6.89714 13.7477 6.53038C13.1928 6.16362 12.5704 5.9112 11.9168 5.78789C11.2632 5.66458 10.5916 5.67285 9.94127 5.81223C9.29092 5.95161 8.6749 6.2193 8.12924 6.59962C7.58359 6.97995 7.11927 7.46527 6.76345 8.02721C6.40763 8.58914 6.16745 9.2164 6.05696 9.87228C5.94648 10.5282 5.96791 11.1995 6.12 11.847C5.23422 12.0469 4.44281 12.5422 3.87581 13.2515C3.30881 13.9608 2.99995 14.8419 3 15.75Z"
              stroke="#FACC15"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        );
      case "clock":
        return (
          <svg
            width="24"
            height="25"
            viewBox="0 0 24 25"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M12 8.5V12.5L15 15.5M21 12.5C21 13.6819 20.7672 14.8522 20.3149 15.9442C19.8626 17.0361 19.1997 18.0282 18.364 18.864C17.5282 19.6997 16.5361 20.3626 15.4442 20.8149C14.3522 21.2672 13.1819 21.5 12 21.5C10.8181 21.5 9.64778 21.2672 8.55585 20.8149C7.46392 20.3626 6.47177 19.6997 5.63604 18.864C4.80031 18.0282 4.13738 17.0361 3.68508 15.9442C3.23279 14.8522 3 13.6819 3 12.5C3 10.1131 3.94821 7.82387 5.63604 6.13604C7.32387 4.44821 9.61305 3.5 12 3.5C14.3869 3.5 16.6761 4.44821 18.364 6.13604C20.0518 7.82387 21 10.1131 21 12.5Z"
              stroke="#F87171"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        );
      case "network":
        return (
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M9 19V13C9 12.4696 8.78929 11.9609 8.41421 11.5858C8.03914 11.2107 7.53043 11 7 11H5C4.46957 11 3.96086 11.2107 3.58579 11.5858C3.21071 11.9609 3 12.4696 3 13V19C3 19.5304 3.21071 20.0391 3.58579 20.4142C3.96086 20.7893 4.46957 21 5 21H7C7.53043 21 8.03914 20.7893 8.41421 20.4142C8.78929 20.0391 9 19.5304 9 19ZM9 19V9C9 8.46957 9.21071 7.96086 9.58579 7.58579C9.96086 7.21071 10.4696 7 11 7H13C13.5304 7 14.0391 7.21071 14.4142 7.58579C14.7893 7.96086 15 8.46957 15 9V19M9 19C9 19.5304 9.21071 20.0391 9.58579 20.4142C9.96086 20.7893 10.4696 21 11 21H13C13.5304 21 14.0391 20.7893 14.4142 20.4142C14.7893 20.0391 15 19.5304 15 19M15 19V5C15 4.46957 15.2107 3.96086 15.5858 3.58579C15.9609 3.21071 16.4696 3 17 3H19C19.5304 3 20.0391 3.21071 20.4142 3.58579C20.7893 3.96086 21 4.46957 21 5V19C21 19.5304 20.7893 20.0391 20.4142 20.4142C20.0391 20.7893 19.5304 21 19 21H17C16.4696 21 15.9609 20.7893 15.5858 20.4142C15.2107 20.0391 15 19.5304 15 19Z"
              stroke="#818CF8"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        );
      case "power":
        return (
          <svg
            width="24"
            height="25"
            viewBox="0 0 24 25"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M13 10.25V3.25L4 14.25H11V21.25L20 10.25H13Z"
              stroke="#4ADE80"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        );
      default:
        return null;
    }
  };

  const useCases = [
    {
      icon: "globe",
      title: t("cases.edge.title"),
      description: t("cases.edge.description"),
      borderColor: "bg-blue-500",
      iconBgColor: "bg-blue-400/20",
      hoverBorderColor: "hover:border-[#60A5FA]",
      backContent: {
        title: t("cases.edge.backContent.title"),
        description: t("cases.edge.backContent.description"),
        features: [
          t("cases.edge.backContent.features.0"),
          t("cases.edge.backContent.features.1"),
          t("cases.edge.backContent.features.2"),
        ],
      },
    },
    {
      icon: "security",
      title: t("cases.compliance.title"),
      description: t("cases.compliance.description"),
      borderColor: "bg-purple-500",
      iconBgColor: "bg-purple-400/20",
      hoverBorderColor: "hover:border-[#C084FC]",
      backContent: {
        title: t("cases.compliance.backContent.title"),
        description: t("cases.compliance.backContent.description"),
        features: [
          t("cases.compliance.backContent.features.0"),
          t("cases.compliance.backContent.features.1"),
          t("cases.compliance.backContent.features.2"),
        ],
      },
    },
    {
      icon: "power",
      title: t("cases.hybrid.title"),
      description: t("cases.hybrid.description"),
      borderColor: "bg-green-500",
      iconBgColor: "bg-green-400/20",
      hoverBorderColor: "hover:border-[#4ADE80]",
      backContent: {
        title: t("cases.hybrid.backContent.title"),
        description: t("cases.hybrid.backContent.description"),
        features: [
          t("cases.hybrid.backContent.features.0"),
          t("cases.hybrid.backContent.features.1"),
          t("cases.hybrid.backContent.features.2"),
        ],
      },
    },
    {
      icon: "clock",
      title: t("cases.dr.title"),
      description: t("cases.dr.description"),
      borderColor: "bg-red-500",
      iconBgColor: "bg-red-400/20",
      hoverBorderColor: "hover:border-[#F87171]",
      backContent: {
        title: t("cases.dr.backContent.title"),
        description: t("cases.dr.backContent.description"),
        features: [
          t("cases.dr.backContent.features.0"),
          t("cases.dr.backContent.features.1"),
          t("cases.dr.backContent.features.2"),
        ],
      },
    },
    {
      icon: "cloud",
      title: t("cases.multitenant.title"),
      description: t("cases.multitenant.description"),
      borderColor: "bg-yellow-500",
      iconBgColor: "bg-yellow-400/20",
      hoverBorderColor: "hover:border-[#FACC15]",
      backContent: {
        title: t("cases.multitenant.backContent.title"),
        description: t("cases.multitenant.backContent.description"),
        features: [
          t("cases.multitenant.backContent.features.0"),
          t("cases.multitenant.backContent.features.1"),
          t("cases.multitenant.backContent.features.2"),
        ],
      },
    },
    {
      icon: "network",
      title: t("cases.performance.title"),
      description: t("cases.performance.description"),
      borderColor: "bg-indigo-500",
      iconBgColor: "bg-indigo-400/20",
      hoverBorderColor: "hover:border-[#818CF8]",
      backContent: {
        title: t("cases.performance.backContent.title"),
        description: t("cases.performance.backContent.description"),
        features: [
          t("cases.performance.backContent.features.0"),
          t("cases.performance.backContent.features.1"),
          t("cases.performance.backContent.features.2"),
        ],
      },
    },
  ];

  return (
    <section
      id="use-cases"
      className="relative py-16 text-white overflow-hidden"
    >
      {/* Dark base background matching the image */}
      <div className="absolute inset-0 bg-[#0a0a0a]"></div>

      {/* Starfield background */}
      <StarField density="low" showComets={true} cometCount={4} />

      {/* Grid lines background */}
      <GridLines horizontalLines={18} verticalLines={15} />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-extrabold text-white sm:text-[2.4rem]">
            {t("title")}{" "}
            <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">
              {t("titleSpan")}
            </span>
          </h2>
          <p className="max-w-2xl mt-3 mx-auto text-lg sm:text-xl text-[#D1D5DB] font-normal px-4">
            {t("subtitle")}
          </p>
        </div>

        <div className="grid gap-6 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 justify-items-center">
          {useCases.map((useCase, index) => (
            <div
              key={index}
              className="group relative w-full max-w-sm h-[320px]"
              style={{ perspective: "1000px" }}
            >
              <div
                className="relative w-full h-full transition-transform duration-700 cursor-pointer"
                style={{
                  transformStyle: "preserve-3d",
                  transform: "rotateY(0deg)",
                }}
                onMouseEnter={e => {
                  e.currentTarget.style.transform = "rotateY(180deg)";
                }}
                onMouseLeave={e => {
                  e.currentTarget.style.transform = "rotateY(0deg)";
                }}
              >
                {/* Front Face */}
                <div
                  className="absolute inset-0 w-full h-full bg-slate-800/50 backdrop-blur-md border border-slate-700 rounded-2xl overflow-hidden transition-all duration-300 hover:shadow-2xl hover:shadow-purple-500/30 hover:border-purple-500/50"
                  style={{ backfaceVisibility: "hidden" }}
                >
                  <div className="p-6 h-full flex flex-col">
                    {/* Logo container */}
                    <div
                      className={`${useCase.iconBgColor} rounded-lg flex items-center justify-center mb-6 w-14 h-14 backdrop-blur-sm border border-white/10`}
                    >
                      <div className="scale-110">{getIcon(useCase.icon)}</div>
                    </div>

                    {/* Main heading */}
                    <h3 className="font-bold text-white mb-5 transition-colors duration-300 group-hover:text-blue-300 text-xl leading-7">
                      {useCase.title}
                    </h3>

                    {/* Description text */}
                    <p className="text-gray-300 font-normal transition-colors duration-300 group-hover:text-gray-200 text-base leading-6 line-clamp-6 flex-grow">
                      {useCase.description}
                    </p>

                    {/* Hover indicator */}
                    <div className="mt-6 opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                      <span className="text-sm text-purple-400 font-medium">
                        Hover to learn more →
                      </span>
                    </div>
                  </div>
                </div>

                {/* Back Face */}
                <div
                  className="absolute inset-0 w-full h-full bg-gradient-to-br from-purple-900/60 via-slate-800/60 to-blue-900/60 backdrop-blur-md border border-purple-500/30 rounded-2xl overflow-hidden"
                  style={{
                    backfaceVisibility: "hidden",
                    transform: "rotateY(180deg)",
                  }}
                >
                  <div className="p-6 h-full flex flex-col justify-center">
                    {/* Enhanced title */}
                    <h3 className="font-bold text-white mb-4 text-xl text-center">
                      {useCase.backContent.title}
                    </h3>

                    {/* Enhanced description */}
                    <p className="text-gray-200 text-sm leading-relaxed mb-6 text-center">
                      {useCase.backContent.description}
                    </p>

                    {/* Features list */}
                    <div className="space-y-2 mb-6">
                      {useCase.backContent.features.map(
                        (feature, featureIndex) => (
                          <div
                            key={featureIndex}
                            className="flex items-center text-xs"
                          >
                            <div className="w-1.5 h-1.5 bg-gradient-to-r from-purple-500 to-blue-500 rounded-full mr-2 flex-shrink-0"></div>
                            <span className="text-gray-300">{feature}</span>
                          </div>
                        )
                      )}
                    </div>

                    {/* Decorative element */}
                    <div className="flex justify-center">
                      <div className="w-12 h-1 bg-gradient-to-r from-purple-500 to-blue-500 rounded-full"></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
