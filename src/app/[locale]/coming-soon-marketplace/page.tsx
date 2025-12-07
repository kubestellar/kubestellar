"use client";

import { useTranslations } from "next-intl";
import {
  Navbar,
  Footer,
  GridLines,
  StarField,
  ComingSoonCTA,
} from "@/components";

export default function ComingSoonPage() {
  const t = useTranslations("comingSoonPage");

  return (
    <div className="min-h-screen bg-black text-white relative overflow-hidden">
      <div className="absolute inset-0 z-0">
        <GridLines />
        <StarField />
      </div>

      <div className="relative z-10">
        <Navbar />

        {/* Hero Section */}
        <section className="px-4 py-32 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-4xl text-center">
            <div className="inline-flex items-center px-4 py-2 rounded-full bg-blue-500/20 border border-blue-500/30 text-blue-300 text-sm font-medium mb-8">
              <div className="w-2 h-2 bg-blue-500 rounded-full mr-2 animate-pulse"></div>
              {t("statusBadge")}
            </div>

            <h1 className="text-5xl sm:text-6xl lg:text-7xl font-bold mb-6">
              <span className="text-gradient">{t("title")}</span>
              <span className="block text-gradient-animated">
                {t("titleSpan")}
              </span>
            </h1>

            <p className="text-xl sm:text-2xl text-gray-300 mb-8 max-w-3xl mx-auto leading-relaxed">
              {t("subtitle")}
            </p>

            <p className="text-lg text-gray-400 max-w-2xl mx-auto mb-16">
              {t("description")}
            </p>
          </div>
        </section>

        {/* Demo Video Section */}
        <section className="px-4 py-16 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-5xl">
            <div className="text-center mb-12">
              <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold mb-4">
                <span className="text-gradient">See It In Action</span>
              </h2>
              <p className="text-lg text-gray-400 max-w-2xl mx-auto">
                Watch our marketplace demo to get a glimpse of what&apos;s coming
              </p>
            </div>

            <div className="relative rounded-2xl overflow-hidden shadow-2xl border border-blue-500/30 bg-gray-900/50 backdrop-blur-sm">
              <div className="aspect-video">
                <iframe
                  className="w-full h-full"
                  src="https://www.youtube.com/embed/ZmuvEQJn98U?si=87ivMCi58ZKY3TBh"
                  title="Marketplace Demo"
                  frameBorder="0"
                  allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                  referrerPolicy="strict-origin-when-cross-origin"
                  allowFullScreen
                />
              </div>
            </div>
          </div>
        </section>

        {/* Call to Action */}
        <ComingSoonCTA />

        <Footer />
      </div>
    </div>
  );
}
