"use client";

import { useEffect } from "react";
import Image from "next/image";
import Navigation from "../../../components/Navigation";
import Footer from "../../../components/Footer";
import StarField from "../../../components/StarField";
import { Program } from "../programs";

interface ProgramPageClientProps {
  program: Program;
}

export default function ProgramPageClient({ program }: ProgramPageClientProps) {
  useEffect(() => {
    // Add CSS for animations and program-specific styling
    const style = document.createElement("style");
    style.textContent = `
      @keyframes gradient {
        0%, 100% { background-position: 0% 50%; }
        50% { background-position: 100% 50%; }
      }
      @keyframes fade-in {
        from { opacity: 0; transform: translateY(20px); }
        to { opacity: 1; transform: translateY(0); }
      }
      @keyframes slide-up {
        from { opacity: 0; transform: translateY(40px); }
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
      .card-enhanced {
        background: linear-gradient(145deg, ${program.theme.primaryColor}1a, ${program.theme.secondaryColor}26);
        backdrop-filter: blur(20px);
        border: 1px solid ${program.theme.primaryColor}33;
        transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
        position: relative;
        overflow: hidden;
      }
      .card-enhanced::before {
        content: '';
        position: absolute;
        top: 0;
        left: -100%;
        width: 100%;
        height: 100%;
        background: linear-gradient(90deg, transparent, ${program.theme.primaryColor}1a, transparent);
        transition: left 0.6s;
      }
      .card-enhanced:hover::before {
        left: 100%;
      }
      .card-enhanced:hover {
        transform: translateY(-8px) scale(1.02);
        box-shadow: 0 25px 50px -12px ${program.theme.primaryColor}66;
        border-color: ${program.theme.primaryColor}80;
      }

      .card-content {
        animation: fade-in 1s ease-out;
      }
      .stagger-animation {
        animation: slide-up 0.8s ease-out;
      }
      .stagger-animation:nth-child(1) { animation-delay: 0.1s; }
      .stagger-animation:nth-child(2) { animation-delay: 0.2s; }
      .stagger-animation:nth-child(3) { animation-delay: 0.3s; }
      .stagger-animation:nth-child(4) { animation-delay: 0.4s; }
      .stagger-animation:nth-child(5) { animation-delay: 0.5s; }
      .stagger-animation:nth-child(6) { animation-delay: 0.6s; }
      .stagger-animation:nth-child(7) { animation-delay: 0.7s; }
      .stagger-animation:nth-child(8) { animation-delay: 0.8s; }
    `;
    document.head.appendChild(style);

    return () => {
      document.head.removeChild(style);
    };
  }, [program]);

  const IconComponent = ({ type }: { type: string }) => {
    const iconPaths = {
      benefits:
        "M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1",
      description: "M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
      overview: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z",
      eligibility:
        "M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z",
      timeline: "M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z",
      structure:
        "M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10",
      howToApply:
        "M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z",
      resources:
        "M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1",
    };

    return (
      <svg
        className="w-6 h-6 text-white"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d={iconPaths[type as keyof typeof iconPaths]}
        />
      </svg>
    );
  };

  return (
    <div className="bg-[#0a0a0a] text-white min-h-screen">
      {/* Navigation */}
      <Navigation />

      {/* Main Content with full background */}
      <main className="relative min-h-screen">
        {/* Background layers */}
        <div className="fixed inset-0 z-0">
          {/* Dark base background */}
          <div className="absolute inset-0 bg-[#0a0a0a]"></div>

          {/* Starfield background */}
          <StarField density="medium" showComets={true} cometCount={3} />

          {/* Grid lines background */}
          <div
            className="absolute inset-0 opacity-10"
            style={{
              backgroundImage: `
              linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px),
              linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)
            `,
              backgroundSize: "50px 50px",
            }}
          ></div>

          {/* Parallax Background */}
        </div>

        {/* Floating Elements */}

        <section className="py-16 relative z-20 pt-32">
          <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center mb-16 card-content">
              <div className="relative inline-block mb-6">
                <Image
                  src={program.logo}
                  alt={`${program.name} Logo`}
                  width={128}
                  height={128}
                  className="h-32 w-auto mx-auto"
                  priority
                />
              </div>
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-extrabold text-gradient mb-4 px-6 py-2 relative z-10">
                {program.fullName}
              </h1>
              <p className="text-lg md:text-xl text-gray-300 max-w-2xl mx-auto px-4 relative z-0">
                {program.description}
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {/* Benefits Card */}
              <div className="md:col-span-2 card-enhanced rounded-xl p-6 shadow-lg stagger-animation">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-orange-500 to-yellow-500 rounded-lg flex items-center justify-center mr-4">
                    <svg
                      className="w-10 h-10 text-white"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"
                      ></path>
                    </svg>
                  </div>
                  <h2 className="text-2xl font-bold text-white">Benefits</h2>
                </div>
                <p className="text-gray-300 leading-relaxed">
                  {program.sections.benefits}
                </p>
              </div>

              {/* Description Card */}
              <div className="card-enhanced rounded-xl p-6 shadow-lg stagger-animation">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-lg flex items-center justify-center mr-4">
                    <IconComponent type="description" />
                  </div>
                  <h2 className="text-2xl font-bold text-white">Description</h2>
                </div>
                <p className="text-gray-300 leading-relaxed">
                  {program.sections.description}
                </p>
              </div>

              {/* Overview Card */}
              <div className="card-enhanced rounded-xl p-6 shadow-lg stagger-animation">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-green-500 to-teal-500 rounded-lg flex items-center justify-center mr-4">
                    <IconComponent type="overview" />
                  </div>
                  <h2 className="text-2xl font-bold text-white">Overview</h2>
                </div>
                <p className="text-gray-300 leading-relaxed">
                  {program.sections.overview}
                </p>
              </div>

              {/* Eligibility Card */}
              <div className="card-enhanced rounded-xl p-6 shadow-lg stagger-animation">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded-lg flex items-center justify-center mr-4">
                    <IconComponent type="eligibility" />
                  </div>
                  <h2 className="text-2xl font-bold text-white">
                    Eligibility Criteria
                  </h2>
                </div>
                <p className="text-gray-300 leading-relaxed">
                  {program.sections.eligibility}
                </p>
              </div>

              {/* Timeline Card */}
              <div className="card-enhanced rounded-xl p-6 shadow-lg stagger-animation">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-indigo-500 to-purple-500 rounded-lg flex items-center justify-center mr-4">
                    <IconComponent type="timeline" />
                  </div>
                  <h2 className="text-2xl font-bold text-white">Timeline</h2>
                </div>
                <p className="text-gray-300 leading-relaxed">
                  {program.sections.timeline}
                </p>
              </div>

              {/* Structure Card */}
              <div className="card-enhanced rounded-xl p-6 shadow-lg stagger-animation">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-teal-500 to-cyan-500 rounded-lg flex items-center justify-center mr-4">
                    <IconComponent type="structure" />
                  </div>
                  <h2 className="text-2xl font-bold text-white">
                    Program Structure
                  </h2>
                </div>
                <p className="text-gray-300 leading-relaxed">
                  {program.sections.structure}
                </p>
              </div>

              {/* How to Apply Card */}
              <div className="card-enhanced rounded-xl p-6 shadow-lg stagger-animation">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-red-500 to-pink-500 rounded-lg flex items-center justify-center mr-4">
                    <IconComponent type="howToApply" />
                  </div>
                  <h2 className="text-2xl font-bold text-white">
                    How to Apply
                  </h2>
                </div>
                <p className="text-gray-300 leading-relaxed">
                  {program.sections.howToApply}
                </p>
              </div>

              {/* Resources Card */}
              <div className="md:col-span-2 card-enhanced rounded-xl p-6 shadow-lg stagger-animation">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-yellow-500 to-orange-500 rounded-lg flex items-center justify-center mr-4">
                    <IconComponent type="resources" />
                  </div>
                  <h2 className="text-2xl font-bold text-white">Resources</h2>
                </div>
                <ul className="space-y-3">
                  {program.sections.resources.map((resource, index) => (
                    <li key={index}>
                      <a
                        href={resource.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-400 hover:text-blue-300 transition-colors duration-300 flex items-center group"
                      >
                        <span className="group-hover:translate-x-2 transition-transform duration-300">
                          →
                        </span>
                        <span className="ml-2">{resource.name}</span>
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <Footer />
    </div>
  );
}
