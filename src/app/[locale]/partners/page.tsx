"use client";

import {
  Navbar,
  Footer,
  GridLines,
  StarField,
} from "../../../components/index";
import { useTranslations } from "next-intl";
import Link from "next/link";
import Image from "next/image";

export default function PartnersPage() {
  const t = useTranslations("partnersPage");

  const partners = [
    {
      id: 1,
      name: "ArgoCD",
      slug: "argocd",
      description: t("partners.argocd.description"),
      logo: "/partners/argocd.png",
      link: "https://argo-cd.readthedocs.io/",
      bgColor: "bg-orange-500/10",
      iconColor: "text-orange-400",
    },
    {
      id: 2,
      name: "FluxCD",
      slug: "fluxcd",
      description: t("partners.fluxcd.description"),
      logo: "/partners/fluxcd.png",
      link: "https://fluxcd.io/",
      bgColor: "bg-blue-500/10",
      iconColor: "text-blue-400",
    },
    {
      id: 3,
      name: "Kyverno",
      slug: "kyverno",
      description: t("partners.kyverno.description"),
      logo: "/partners/kyverno.png",
      link: "https://kyverno.io/",
      bgColor: "bg-green-500/10",
      iconColor: "text-green-400",
    },
    {
      id: 4,
      name: "Mvi",
      slug: "mvi",
      description: t("partners.mvi.description"),
      logo: "/partners/mvi.png",
      link: "/",
      bgColor: "bg-cyan-500/10",
      iconColor: "text-cyan-400",
    },
    {
      id: 5,
      name: "OpenZiti",
      slug: "openziti",
      description: t("partners.openziti.description"),
      logo: "/partners/openziti.png",
      link: "https://openziti.io/",
      bgColor: "bg-purple-500/10",
      iconColor: "text-purple-400",
    },
    {
      id: 6,
      name: "Turbonomic",
      slug: "turbonomic",
      description: t("partners.turbonomic.description"),
      logo: "/partners/turbonomic.webp",
      link: "https://www.ibm.com/products/turbonomic",
      bgColor: "bg-yellow-500/10",
      iconColor: "text-yellow-400",
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

      <div className="relative z-10 pt-16">
        {" "}
        {/* Add padding-top to account for fixed navbar */}
        {/* Header Section */}
        <section className="py-16 sm:py-20 lg:py-24 pb-8">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center">
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-white">
                {t("title")}{" "}
                <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">
                  {t("titleSpan")}
                </span>
              </h1>
              <p className="text-xl md:text-2xl text-gray-300 max-w-4xl mx-auto leading-relaxed">
                {t("subtitle")}
              </p>
            </div>
          </div>
        </section>
        {/* Partners Carousel Section */}
        <section className="py-4 sm:py-8 md:py-12 lg:py-16 overflow-hidden">
          {/* Desktop Sliding View */}
          <div className="hidden lg:block">
            <div className="relative overflow-hidden">
              <div className="flex gap-6 animate-slide-partners group-hover:pause">
                {/* Triple the partners for seamless infinite loop */}
                {[...partners, ...partners, ...partners].map(
                  (partner, index) => (
                    <div
                      key={index}
                      className="flex-shrink-0 w-[400px] group/card"
                      onMouseEnter={e => {
                        e.currentTarget
                          .closest(".animate-slide-partners")
                          ?.classList.add("pause-animation");
                      }}
                      onMouseLeave={e => {
                        e.currentTarget
                          .closest(".animate-slide-partners")
                          ?.classList.remove("pause-animation");
                      }}
                    >
                      <Link
                        href={partner.link}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="relative block bg-slate-800/50 border border-slate-700 rounded-2xl p-10 h-96 overflow-hidden transition-all duration-300 cursor-pointer hover:shadow-2xl hover:shadow-purple-500/30 hover:border-purple-500/50"
                      >
                        <div className="transition-all duration-300 group-hover/card:-translate-y-2 h-full flex flex-col">
                          <div className="mb-6">
                            <Image
                              src={partner.logo}
                              alt={`${partner.name} logo`}
                              width={partner.slug === "argocd" ? 130 : 96}
                              height={partner.slug === "argocd" ? 130 : 96}
                              className={`${partner.slug === "argocd" ? "w-[130px] h-[130px]" : "w-24 h-24"} object-contain ${partner.slug === "mvi" || partner.slug === "turbonomic" ? "bg-white rounded-lg p-2" : ""}`}
                            />
                          </div>
                          <h3 className="text-3xl font-bold text-white mb-5">
                            {partner.name}
                          </h3>
                          <p className="text-gray-300 leading-relaxed text-base flex-grow">
                            {partner.description}
                          </p>
                        </div>
                        <div className="absolute bottom-6 right-6 opacity-0 group-hover/card:opacity-100 transition-opacity duration-300">
                          <span className="learn-more-enhanced">
                            {t("learnMore")}
                          </span>
                        </div>
                      </Link>
                    </div>
                  )
                )}
              </div>
            </div>
          </div>

          {/* Mobile/Tablet Grid View */}
          <div className="lg:hidden max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {partners.map(partner => (
                <Link
                  key={partner.id}
                  href={partner.link}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="relative group bg-slate-800/50 border border-slate-700 rounded-xl p-10 h-96 overflow-hidden transition-all duration-300 cursor-pointer hover:shadow-2xl hover:shadow-purple-500/30"
                >
                  <div className="transition-all duration-300 group-hover:-translate-y-2 h-full flex flex-col">
                    <div className="mb-6">
                      <Image
                        src={partner.logo}
                        alt={`${partner.name} logo`}
                        width={partner.slug === "argocd" ? 130 : 96}
                        height={partner.slug === "argocd" ? 130 : 96}
                        className={`${partner.slug === "argocd" ? "w-[130px] h-[130px]" : "w-24 h-24"} object-contain ${partner.slug === "mvi" || partner.slug === "turbonomic" ? "bg-white rounded-lg p-2" : ""}`}
                      />
                    </div>
                    <h3 className="text-3xl font-bold text-white mb-5">
                      {partner.name}
                    </h3>
                    <p className="text-gray-300 leading-relaxed text-base flex-grow">
                      {partner.description}
                    </p>
                  </div>
                  <div className="absolute bottom-6 right-6 opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                    <span className="learn-more-enhanced">
                      {t("learnMore")}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        </section>
        {/* Why Partner With Us Section */}
        <section className="py-16 sm:py-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="bg-gray-800/40 backdrop-blur-md rounded-lg p-8 md:p-12 border border-white/10">
              <h2 className="text-3xl md:text-4xl font-bold text-white mb-8 text-center">
                {t("whyPartner.title")}
              </h2>
              <p className="text-gray-300 text-center mb-12 text-lg max-w-3xl mx-auto">
                {t("whyPartner.subtitle")}
              </p>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                <div className="text-center">
                  <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4">
                    <svg
                      className="w-8 h-8 text-white"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                      />
                    </svg>
                  </div>
                  <h3 className="text-xl font-bold text-white mb-3">
                    {t("whyPartner.benefits.0.title")}
                  </h3>
                  <p className="text-gray-300 text-sm">
                    {t("whyPartner.benefits.0.description")}
                  </p>
                </div>

                <div className="text-center">
                  <div className="w-16 h-16 bg-gradient-to-br from-green-500 to-emerald-600 rounded-full flex items-center justify-center mx-auto mb-4">
                    <svg
                      className="w-8 h-8 text-white"
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
                  </div>
                  <h3 className="text-xl font-bold text-white mb-3">
                    {t("whyPartner.benefits.1.title")}
                  </h3>
                  <p className="text-gray-300 text-sm">
                    {t("whyPartner.benefits.1.description")}
                  </p>
                </div>

                <div className="text-center">
                  <div className="w-16 h-16 bg-gradient-to-br from-orange-500 to-red-600 rounded-full flex items-center justify-center mx-auto mb-4">
                    <svg
                      className="w-8 h-8 text-white"
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
                  </div>
                  <h3 className="text-xl font-bold text-white mb-3">
                    {t("whyPartner.benefits.2.title")}
                  </h3>
                  <p className="text-gray-300 text-sm">
                    {t("whyPartner.benefits.2.description")}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>
        {/* Partnership Opportunities Section */}
        <section className="py-16 sm:py-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="bg-gradient-to-br from-purple-900/30 via-slate-800/40 to-blue-900/30 backdrop-blur-md rounded-2xl p-8 md:p-12 border border-purple-500/20">
              <div className="text-center mb-8">
                <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
                  {t("partnershipOpportunities.title")}
                </h2>
                <p className="text-gray-300 text-lg max-w-3xl mx-auto leading-relaxed">
                  {t("partnershipOpportunities.subtitle")}
                </p>
              </div>

              <div className="max-w-4xl mx-auto mb-8">
                <div className="bg-slate-800/30 rounded-xl p-8 border border-slate-700/50">
                  <p className="text-gray-300 text-base leading-relaxed mb-6">
                    {t("partnershipOpportunities.description")}
                  </p>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                    <ul className="space-y-2">
                      <li className="flex items-start">
                        <svg
                          className="w-5 h-5 text-green-400 mr-2 mt-0.5 flex-shrink-0"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M5 13l4 4L19 7"
                          />
                        </svg>
                        <span className="text-gray-300">
                          {t("partnershipOpportunities.features.0")}
                        </span>
                      </li>
                      <li className="flex items-start">
                        <svg
                          className="w-5 h-5 text-green-400 mr-2 mt-0.5 flex-shrink-0"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M5 13l4 4L19 7"
                          />
                        </svg>
                        <span className="text-gray-300">
                          {t("partnershipOpportunities.features.1")}
                        </span>
                      </li>
                      <li className="flex items-start">
                        <svg
                          className="w-5 h-5 text-green-400 mr-2 mt-0.5 flex-shrink-0"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M5 13l4 4L19 7"
                          />
                        </svg>
                        <span className="text-gray-300">
                          {t("partnershipOpportunities.features.2")}
                        </span>
                      </li>
                    </ul>
                    <ul className="space-y-2">
                      <li className="flex items-start">
                        <svg
                          className="w-5 h-5 text-green-400 mr-2 mt-0.5 flex-shrink-0"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M5 13l4 4L19 7"
                          />
                        </svg>
                        <span className="text-gray-300">
                          {t("partnershipOpportunities.features.3")}
                        </span>
                      </li>
                      <li className="flex items-start">
                        <svg
                          className="w-5 h-5 text-green-400 mr-2 mt-0.5 flex-shrink-0"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M5 13l4 4L19 7"
                          />
                        </svg>
                        <span className="text-gray-300">
                          {t("partnershipOpportunities.features.4")}
                        </span>
                      </li>
                      <li className="flex items-start">
                        <svg
                          className="w-5 h-5 text-green-400 mr-2 mt-0.5 flex-shrink-0"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M5 13l4 4L19 7"
                          />
                        </svg>
                        <span className="text-gray-300">
                          {t("partnershipOpportunities.features.5")}
                        </span>
                      </li>
                    </ul>
                  </div>
                </div>
              </div>

              <div className="text-center">
                <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
                  <a
                    href="#contact-us"
                    className="inline-flex items-center px-8 py-4 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white font-semibold rounded-lg transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-purple-500/50"
                  >
                    <svg
                      className="w-5 h-5 mr-2"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                      />
                    </svg>
                    {t("partnershipOpportunities.contactButton")}
                  </a>

                  <a
                    href="https://kubestellar.io/slack"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center px-8 py-4 bg-gradient-to-r from-gray-700 to-gray-800 hover:from-gray-600 hover:to-gray-700 text-white font-semibold rounded-lg transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-gray-500/30 border border-gray-600/50"
                  >
                    <svg
                      className="w-5 h-5 mr-2"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 60 60"
                      preserveAspectRatio="xMidYMid meet"
                    >
                      <path
                        d="M22,12 a6,6 0 1 1 6,-6 v6z M22,16 a6,6 0 0 1 0,12 h-16 a6,6 0 1 1 0,-12"
                        fill="#36C5F0"
                      />
                      <path
                        d="M48,22 a6,6 0 1 1 6,6 h-6z M32,6 a6,6 0 1 1 12,0v16a6,6 0 0 1 -12,0z"
                        fill="#2EB67D"
                      />
                      <path
                        d="M38,48 a6,6 0 1 1 -6,6 v-6z M54,32 a6,6 0 0 1 0,12 h-16 a6,6 0 1 1 0,-12"
                        fill="#ECB22E"
                      />
                      <path
                        d="M12,38 a6,6 0 1 1 -6,-6 h6z M16,38 a6,6 0 1 1 12,0v16a6,6 0 0 1 -12,0z"
                        fill="#E01E5A"
                      />
                    </svg>
                    Join the Slack
                  </a>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
      <Footer />
    </div>
  );
}
