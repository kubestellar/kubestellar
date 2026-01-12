"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { Program } from "../programs";
import { Navbar, Footer, StarField, GridLines } from "@/components";
import { useTranslations } from "next-intl";
import Link from "next/link";

interface ProgramPageClientProps {
  program: Program;
}

// Section configuration for sidebar navigation
const sectionConfig = [
  { id: "overview", icon: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" },
  { id: "benefits", icon: "M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" },
  { id: "description", icon: "M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" },
  { id: "eligibility", icon: "M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" },
  { id: "timeline", icon: "M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" },
  { id: "structure", icon: "M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" },
  { id: "howToApply", icon: "M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" },
  { id: "resources", icon: "M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" },
];

export default function ProgramPageClient({ program }: ProgramPageClientProps) {
  const t = useTranslations("programsPage");
  const tDetails = useTranslations("programDetailsPage");
  const [activeSection, setActiveSection] = useState("overview");

  useEffect(() => {
    // Add CSS for animations and program-specific styling
    const style = document.createElement("style");
    style.textContent = `
      @keyframes gradient {
        0%, 100% { background-position: 0% 50%; }
        50% { background-position: 100% 50%; }
      }
      @keyframes fade-in {
        from { opacity: 0; transform: translateY(10px); }
        to { opacity: 1; transform: translateY(0); }
      }
      .text-gradient {
        background-clip: text;
        -webkit-background-clip: text;
        color: transparent;
        background-image: ${program.theme.gradient};
        background-size: 300% 300%;
        animation: gradient 3s ease infinite;
      }
      .section-content {
        animation: fade-in 0.4s ease-out;
      }
      .nav-item-active {
        background: linear-gradient(90deg, ${program.theme.primaryColor}20, transparent);
        border-left: 3px solid ${program.theme.primaryColor};
      }
      .nav-item:hover {
        background: linear-gradient(90deg, ${program.theme.primaryColor}10, transparent);
      }
      .prose-program h2 {
        color: ${program.theme.primaryColor};
      }
      .accent-border {
        border-left: 4px solid ${program.theme.primaryColor};
      }
      .program-link {
        color: ${program.theme.primaryColor};
      }
      .program-link:hover {
        color: ${program.theme.secondaryColor};
      }
      html {
        scroll-behavior: smooth;
      }
    `;
    document.head.appendChild(style);

    // Intersection Observer for scroll spy
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveSection(entry.target.id);
          }
        });
      },
      { rootMargin: "-20% 0px -60% 0px", threshold: 0 }
    );

    sectionConfig.forEach(({ id }) => {
      const element = document.getElementById(id);
      if (element) observer.observe(element);
    });

    return () => {
      document.head.removeChild(style);
      observer.disconnect();
    };
  }, [program]);

  const scrollToSection = (sectionId: string) => {
    const element = document.getElementById(sectionId);
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  return (
    <div className="bg-[#0a0a0a] text-white min-h-screen">
      {/* Navigation */}
      <Navbar />

      {/* Main Content with full background */}
      <main className="relative min-h-screen">
        {/* Background layers */}
        <div className="fixed inset-0 z-0">
          <div className="absolute inset-0 bg-[#0a0a0a]"></div>
          <StarField density="medium" showComets={true} cometCount={3} />
          <GridLines horizontalLines={21} verticalLines={18} />
        </div>

        {/* Hero Section */}
        <section className="relative z-20 pt-24 pb-8 border-b border-gray-800/50">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center gap-6">
              <div className="relative flex-shrink-0">
                <Image
                  src={program.logo}
                  alt={`${t(`programs.${program.id}.name`)} Logo`}
                  width={80}
                  height={80}
                  className="h-20 w-auto"
                  priority
                />
              </div>
              <div>
                <div className="flex items-center gap-3 mb-2">
                  <h1 className="text-3xl md:text-4xl font-bold text-gradient">
                    {t(`programs.${program.id}.fullName`)}
                  </h1>
                  {program.isPaid && (
                    <span className="px-3 py-1 text-xs font-semibold rounded-full bg-green-500/20 text-green-400 border border-green-500/30">
                      Paid Program
                    </span>
                  )}
                  {!program.isPaid && (
                    <span className="px-3 py-1 text-xs font-semibold rounded-full bg-blue-500/20 text-blue-400 border border-blue-500/30">
                      Volunteer Program
                    </span>
                  )}
                </div>
                <p className="text-gray-400 text-lg max-w-2xl">
                  {t(`programs.${program.id}.description`)}
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Documentation Layout */}
        <div className="relative z-20 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex gap-8">
            
            {/* Sidebar Navigation */}
            <aside className="hidden lg:block w-64 flex-shrink-0">
              <nav className="sticky top-24 space-y-1">
                <div className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-4 px-3">
                  On This Page
                </div>
                {sectionConfig.map(({ id, icon }) => (
                  <button
                    key={id}
                    onClick={() => scrollToSection(id)}
                    className={`nav-item w-full flex items-center gap-3 px-3 py-2 text-sm rounded-r-lg transition-all duration-200 text-left ${
                      activeSection === id
                        ? "nav-item-active text-white font-medium"
                        : "text-gray-400 hover:text-gray-200"
                    }`}
                  >
                    <svg
                      className="w-4 h-4 flex-shrink-0"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d={icon}
                      />
                    </svg>
                    <span>{tDetails(id)}</span>
                  </button>
                ))}
                
                {/* Back to Programs Link */}
                <div className="pt-6 mt-6 border-t border-gray-800">
                  <Link
                    href="/programs"
                    className="flex items-center gap-2 px-3 py-2 text-sm text-gray-400 hover:text-white transition-colors"
                  >
                    <svg
                      className="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M10 19l-7-7m0 0l7-7m-7 7h18"
                      />
                    </svg>
                    <span>Back to Programs</span>
                  </Link>
                </div>
              </nav>
            </aside>

            {/* Main Content */}
            <article className="flex-1 min-w-0 max-w-4xl">
              <div className="prose prose-invert prose-program max-w-none">
                
                {/* Overview Section */}
                <section id="overview" className="section-content mb-12 scroll-mt-24">
                  <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
                    <span className="w-8 h-8 rounded-lg bg-gradient-to-r from-green-500 to-teal-500 flex items-center justify-center">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </span>
                    {tDetails("overview")}
                  </h2>
                  <div className="accent-border pl-4 py-2 bg-gray-900/50 rounded-r-lg">
                    <p className="text-gray-300 leading-relaxed m-0">
                      {t(`programs.${program.id}.sections.overview`)}
                    </p>
                  </div>
                </section>

                {/* Benefits Section */}
                <section id="benefits" className="section-content mb-12 scroll-mt-24">
                  <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
                    <span className="w-8 h-8 rounded-lg bg-gradient-to-r from-orange-500 to-yellow-500 flex items-center justify-center">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
                      </svg>
                    </span>
                    {tDetails("benefits")}
                  </h2>
                  <div className="accent-border pl-4 py-2 bg-gray-900/50 rounded-r-lg">
                    <p className="text-gray-300 leading-relaxed m-0">
                      {t(`programs.${program.id}.sections.benefits`)}
                    </p>
                  </div>
                </section>

                {/* Description Section */}
                <section id="description" className="section-content mb-12 scroll-mt-24">
                  <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
                    <span className="w-8 h-8 rounded-lg bg-gradient-to-r from-blue-500 to-cyan-500 flex items-center justify-center">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </span>
                    {tDetails("description")}
                  </h2>
                  <div className="accent-border pl-4 py-2 bg-gray-900/50 rounded-r-lg">
                    <p className="text-gray-300 leading-relaxed m-0">
                      {t(`programs.${program.id}.sections.description`)}
                    </p>
                  </div>
                </section>

                {/* Eligibility Section */}
                <section id="eligibility" className="section-content mb-12 scroll-mt-24">
                  <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
                    <span className="w-8 h-8 rounded-lg bg-gradient-to-r from-purple-500 to-pink-500 flex items-center justify-center">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                    </span>
                    {tDetails("eligibility")}
                  </h2>
                  <div className="accent-border pl-4 py-2 bg-gray-900/50 rounded-r-lg">
                    <p className="text-gray-300 leading-relaxed m-0">
                      {t(`programs.${program.id}.sections.eligibility`)}
                    </p>
                  </div>
                </section>

                {/* Timeline Section */}
                <section id="timeline" className="section-content mb-12 scroll-mt-24">
                  <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
                    <span className="w-8 h-8 rounded-lg bg-gradient-to-r from-indigo-500 to-purple-500 flex items-center justify-center">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </span>
                    {tDetails("timeline")}
                  </h2>
                  <div className="accent-border pl-4 py-2 bg-gray-900/50 rounded-r-lg">
                    <p className="text-gray-300 leading-relaxed m-0">
                      {t(`programs.${program.id}.sections.timeline`)}
                    </p>
                  </div>
                </section>

                {/* Structure Section */}
                <section id="structure" className="section-content mb-12 scroll-mt-24">
                  <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
                    <span className="w-8 h-8 rounded-lg bg-gradient-to-r from-teal-500 to-cyan-500 flex items-center justify-center">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                      </svg>
                    </span>
                    {tDetails("structure")}
                  </h2>
                  <div className="accent-border pl-4 py-2 bg-gray-900/50 rounded-r-lg">
                    <p className="text-gray-300 leading-relaxed m-0">
                      {t(`programs.${program.id}.sections.structure`)}
                    </p>
                  </div>
                </section>

                {/* How to Apply Section */}
                <section id="howToApply" className="section-content mb-12 scroll-mt-24">
                  <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
                    <span className="w-8 h-8 rounded-lg bg-gradient-to-r from-red-500 to-pink-500 flex items-center justify-center">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                      </svg>
                    </span>
                    {tDetails("howToApply")}
                  </h2>
                  <div className="accent-border pl-4 py-2 bg-gray-900/50 rounded-r-lg">
                    <p className="text-gray-300 leading-relaxed m-0">
                      {t(`programs.${program.id}.sections.howToApply`)}
                    </p>
                  </div>
                </section>

                {/* Resources Section */}
                <section id="resources" className="section-content mb-12 scroll-mt-24">
                  <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
                    <span className="w-8 h-8 rounded-lg bg-gradient-to-r from-yellow-500 to-orange-500 flex items-center justify-center">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                      </svg>
                    </span>
                    {tDetails("resources")}
                  </h2>
                  <div className="accent-border pl-4 py-3 bg-gray-900/50 rounded-r-lg">
                    <ul className="space-y-3 m-0 p-0 list-none">
                      {program.sections.resources.map((resource, index) => (
                        <li key={index} className="flex items-center gap-3">
                          <svg
                            className="w-4 h-4 text-gray-500 flex-shrink-0"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth="2"
                              d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                            />
                          </svg>
                          <a
                            href={resource.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="program-link hover:underline transition-colors duration-200"
                          >
                            {resource.name}
                          </a>
                        </li>
                      ))}
                    </ul>
                  </div>
                </section>

                {/* Quick Actions */}
                <section className="mt-16 pt-8 border-t border-gray-800">
                  <h3 className="text-lg font-semibold text-white mb-6">Quick Actions</h3>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    {program.sections.resources[0] && (
                      <a
                        href={program.sections.resources[0].url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="group flex items-center gap-4 p-4 rounded-xl bg-gray-900/50 border border-gray-800 hover:border-gray-700 transition-all duration-300"
                      >
                        <div className="w-10 h-10 rounded-lg bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center flex-shrink-0">
                          <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                          </svg>
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="text-white font-medium group-hover:text-blue-400 transition-colors">
                            Visit Official Website
                          </div>
                          <div className="text-sm text-gray-500 truncate">
                            {program.sections.resources[0].url}
                          </div>
                        </div>
                        <svg className="w-5 h-5 text-gray-500 group-hover:text-white group-hover:translate-x-1 transition-all" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
                        </svg>
                      </a>
                    )}
                    <Link
                      href="/programs"
                      className="group flex items-center gap-4 p-4 rounded-xl bg-gray-900/50 border border-gray-800 hover:border-gray-700 transition-all duration-300"
                    >
                      <div className="w-10 h-10 rounded-lg bg-gradient-to-r from-gray-600 to-gray-500 flex items-center justify-center flex-shrink-0">
                        <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                        </svg>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="text-white font-medium group-hover:text-blue-400 transition-colors">
                          Explore All Programs
                        </div>
                        <div className="text-sm text-gray-500">
                          View other mentorship opportunities
                        </div>
                      </div>
                      <svg className="w-5 h-5 text-gray-500 group-hover:text-white group-hover:translate-x-1 transition-all" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </Link>
                  </div>
                </section>

              </div>
            </article>

            {/* Table of Contents - Right side (visible on xl screens) */}
            <aside className="hidden xl:block w-56 flex-shrink-0">
              <div className="sticky top-24">
                <div className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-4">
                  Program Type
                </div>
                <div className={`inline-flex items-center gap-2 px-3 py-2 rounded-lg ${
                  program.isPaid 
                    ? "bg-green-500/10 border border-green-500/20" 
                    : "bg-blue-500/10 border border-blue-500/20"
                }`}>
                  <div className={`w-2 h-2 rounded-full ${program.isPaid ? "bg-green-500" : "bg-blue-500"}`} />
                  <span className={`text-sm font-medium ${program.isPaid ? "text-green-400" : "text-blue-400"}`}>
                    {program.isPaid ? "Paid Stipend" : "Volunteer"}
                  </span>
                </div>
                
                <div className="mt-8 p-4 rounded-xl bg-gray-900/50 border border-gray-800">
                  <div className="text-sm font-medium text-white mb-2">Need Help?</div>
                  <p className="text-xs text-gray-400 mb-3">
                    Join our community to get guidance on applying to this program.
                  </p>
                  <a
                    href="https://github.com/kubestellar/kubestellar"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs program-link hover:underline"
                  >
                    Visit GitHub →
                  </a>
                </div>
              </div>
            </aside>
          </div>
        </div>
      </main>

      {/* Footer */}
      <Footer />
    </div>
  );
}
