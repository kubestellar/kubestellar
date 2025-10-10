"use client";

import { getAllPrograms } from "./programs";
import Image from "next/image";
import Link from "next/link";
import { useEffect } from "react";
import Navigation from "../../components/Navigation";
import Footer from "../../components/Footer";
import { GridLines, StarField } from "@/components";

export default function ProgramsPage() {
  const programs = getAllPrograms();

  useEffect(() => {
    // Add CSS for animations
    const style = document.createElement("style");
    style.textContent = `
      @keyframes twinkle {
        0%, 100% { opacity: 0.2; }
        50% { opacity: 1; }
      }
      .text-gradient {
        background-clip: text;
        -webkit-background-clip: text;
        color: transparent;
        background-image: linear-gradient(to right, #8B5CF6, #3B82F6);
      }
      .program-card:hover {
        transform: translateY(-0.5rem);
        box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
      }
      .background-grid {
        background-image: 
          linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px),
          linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px);
        background-size: 50px 50px;
      }
    `;
    document.head.appendChild(style);

    return () => {
      document.head.removeChild(style);
    };
  }, []);

  return (
    <div className="bg-[#0a0a0a] text-white overflow-x-hidden min-h-screen">
      {/* Navigation */}
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

      {/* Hero Section */}
      <section className="relative min-h-[40vh] flex items-center justify-center z-10">
        <div className="relative z-10 text-center px-4 pt-20 pb-2">
          <h1 className="text-4xl md:text-6xl font-extrabold tracking-tighter text-shadow-lg">
            Join Our <span className="text-gradient">Mission</span>
          </h1>
          <p className="mt-4 max-w-2xl mx-auto text-lg md:text-xl text-gray-300">
            Discover meaningful opportunities to contribute to KubeStellar and
            advance your career in open source development.
          </p>
        </div>
      </section>

      {/* Programs Section */}
      <section id="programs" className="relative pt-8 pb-24 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
          <div className="grid gap-10 md:grid-cols-2 lg:grid-cols-2">
            {programs.map(program => (
              <Link
                key={program.id}
                href={`/programs/${program.id}`}
                className={`program-card bg-gray-800/50 backdrop-blur-md rounded-xl shadow-lg border border-gray-700/50 p-8 flex flex-col items-center text-center transition-all duration-300 hover:shadow-2xl hover:-translate-y-2 ${
                  program.isPaid
                    ? "hover:border-blue-500/50"
                    : "hover:border-purple-500/50"
                }`}
              >
                <div className="relative w-32 h-24 mb-6 flex items-center justify-center">
                  <Image
                    src={program.logo}
                    alt={`${program.name} Logo`}
                    fill
                    className="object-contain"
                    sizes="128px"
                    priority
                  />
                </div>
                <h3 className="text-2xl font-bold text-white mb-2">
                  {program.fullName}
                </h3>
                <p className="text-gray-400 mb-4 flex-grow">
                  {program.description}
                </p>
                <span
                  className={`inline-block text-xs font-semibold px-3 py-1 rounded-full ${
                    program.isPaid
                      ? "bg-green-500/20 text-green-300"
                      : "bg-purple-500/20 text-purple-300"
                  }`}
                >
                  {program.isPaid ? "Paid Program" : "Unpaid Internship"}
                </span>
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <Footer />
    </div>
  );
}
