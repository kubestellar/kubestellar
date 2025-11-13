"use client";

import React, { useState, useMemo, useRef, useEffect } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { Navbar, Footer } from "@/components";
import { GridLines, StarField } from "@/components/";
import { usePlugins } from "./plugins";

export default function MarketplacePage() {
  const t = useTranslations("marketplace");
  const plugins = usePlugins();
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All");
  const [selectedPricing, setSelectedPricing] = useState("All");
  const [isCategoryOpen, setIsCategoryOpen] = useState(false);
  const [isPricingOpen, setIsPricingOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const categoryRef = useRef<HTMLDivElement>(null);
  const pricingRef = useRef<HTMLDivElement>(null);
  const PLUGINS_PER_PAGE = 12;

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        categoryRef.current &&
        !categoryRef.current.contains(event.target as Node)
      ) {
        setIsCategoryOpen(false);
      }
      if (
        pricingRef.current &&
        !pricingRef.current.contains(event.target as Node)
      ) {
        setIsPricingOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Extract unique categories
  const categories = useMemo(() => {
    const cats = new Set(plugins.map((p) => p.category));
    return ["All", ...Array.from(cats)];
  }, []);

  // Filter plugins
  const filteredPlugins = useMemo(() => {
    return plugins.filter((plugin) => {
      const matchesSearch =
        plugin.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        plugin.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        plugin.tags.some((tag) =>
          tag.toLowerCase().includes(searchQuery.toLowerCase())
        );

      const matchesCategory =
        selectedCategory === "All" || plugin.category === selectedCategory;

      const matchesPricing =
        selectedPricing === "All" || plugin.pricing.type === selectedPricing;

      return matchesSearch && matchesCategory && matchesPricing;
    });
  }, [searchQuery, selectedCategory, selectedPricing]);

  // Pagination
  const totalPages = Math.ceil(filteredPlugins.length / PLUGINS_PER_PAGE);
  const paginatedPlugins = useMemo(() => {
    const startIndex = (currentPage - 1) * PLUGINS_PER_PAGE;
    return filteredPlugins.slice(startIndex, startIndex + PLUGINS_PER_PAGE);
  }, [filteredPlugins, currentPage]);

  // Reset to page 1 when filters change
  useEffect(() => {
    setCurrentPage(1);
  }, [searchQuery, selectedCategory, selectedPricing]);

  return (
    <main className="min-h-screen bg-[#0a0a0a]">
      <Navbar />

      {/* Hero Section */}
      <section className="relative pt-40 pb-32 overflow-hidden">
        {/* Background Effects */}
        <div className="absolute inset-0 z-0">
          <StarField density="medium" showComets={true} cometCount={3} />
          <GridLines />
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="text-center mb-16">
            <h1 className="text-5xl md:text-7xl font-bold mb-8">
              <span className="bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent animate-gradient-x">
                KubeStellar Galaxy
              </span>
              <br />
              <span className="text-white">Marketplace</span>
            </h1>
            <p className="text-xl md:text-2xl text-gray-300 max-w-3xl mx-auto mb-8">
              Extend your KubeStellar deployment with powerful plugins and
              tools. From free community projects to enterprise solutions.
            </p>
            
            {/* Stats */}
            <div className="flex flex-wrap justify-center gap-8 mt-12">
              <div className="text-center">
                <div className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-purple-400 to-pink-500 bg-clip-text text-transparent mb-2">
                  {plugins.length}+
                </div>
                <div className="text-gray-400 text-sm md:text-base">Plugins Available</div>
              </div>
              <div className="text-center">
                <div className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent mb-2">
                  {plugins.filter(p => p.pricing.type === 'free').length}
                </div>
                <div className="text-gray-400 text-sm md:text-base">Free Plugins</div>
              </div>
              <div className="text-center">
                <div className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-green-400 to-blue-500 bg-clip-text text-transparent mb-2">
                  {Math.floor(plugins.reduce((sum, p) => sum + p.downloads, 0) / 100000)}K+
                </div>
                <div className="text-gray-400 text-sm md:text-base">Total Downloads</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Featured & Most Popular Plugins Carousel */}
      <section className="relative py-12 overflow-hidden">
        <div className="absolute inset-0 z-0">
          <StarField density="low" showComets={false} cometCount={0} />
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10 mb-12">
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-2 text-center">
            <span className="bg-gradient-to-r from-purple-400 to-pink-500 bg-clip-text text-transparent">
              {t("featured.title")}
            </span>{" "}
            {t("featured.titleSuffix")}
          </h2>
          <p className="text-gray-400 text-center mb-8">
            {t("featured.subtitle")}
          </p>
        </div>

        {/* Desktop Sliding View */}
        <div className="hidden lg:block relative">
          <div className="overflow-hidden">
            <div className="flex gap-6 animate-slide-partners">
              {/* Get top 6 plugins by downloads and triple them for seamless loop */}
              {[
                ...plugins.slice().sort((a, b) => b.downloads - a.downloads).slice(0, 6),
                ...plugins.slice().sort((a, b) => b.downloads - a.downloads).slice(0, 6),
                ...plugins.slice().sort((a, b) => b.downloads - a.downloads).slice(0, 6),
              ].map((plugin, index) => (
                <div
                  key={`${plugin.id}-${index}`}
                  className="flex-shrink-0 w-[400px] group/card"
                  onMouseEnter={(e) => {
                    e.currentTarget
                      .closest(".animate-slide-partners")
                      ?.classList.add("pause-animation");
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget
                      .closest(".animate-slide-partners")
                      ?.classList.remove("pause-animation");
                  }}
                >
                  <Link
                    href={`/marketplace/${plugin.slug}`}
                    className="relative block bg-gray-800/50 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-8 h-80 overflow-hidden transition-all duration-300 hover:shadow-2xl hover:shadow-purple-500/30 hover:border-purple-500/50 hover:-translate-y-1"
                  >
                    {/* Badge for pricing */}
                    <div className="absolute top-4 right-4">
                      {plugin.pricing.type === "free" ? (
                        <span className="px-3 py-1 bg-green-500/20 text-green-300 text-xs font-semibold rounded-full">
                          FREE
                        </span>
                      ) : (
                        <span className="px-3 py-1 bg-purple-500/20 text-purple-300 text-xs font-semibold rounded-full">
                          ${plugin.pricing.amount}
                        </span>
                      )}
                    </div>

                    <div className="transition-all duration-300 group-hover/card:-translate-y-2 h-full flex flex-col">
                      {/* Icon */}
                      <div className="text-6xl mb-4 group-hover/card:scale-110 transition-transform duration-300">
                        {plugin.icon}
                      </div>

                      {/* Title */}
                      <h3 className="text-2xl font-bold text-white mb-3 group-hover/card:text-purple-400 transition-colors">
                        {plugin.name}
                      </h3>

                      {/* Description */}
                      <p className="text-gray-300 leading-relaxed text-sm mb-4 flex-grow line-clamp-3">
                        {plugin.tagline}
                      </p>

                      {/* Stats */}
                      <div className="flex items-center gap-4 text-sm text-gray-400 mb-4">
                        <div className="flex items-center gap-1">
                          <svg
                            className="w-4 h-4 text-yellow-500"
                            fill="currentColor"
                            viewBox="0 0 20 20"
                          >
                            <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                          </svg>
                          <span>{plugin.rating}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <svg
                            className="w-4 h-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                            />
                          </svg>
                          <span>{plugin.downloads.toLocaleString()}</span>
                        </div>
                      </div>

                      {/* Category Badge */}
                      <div>
                        <span className="inline-block px-3 py-1 bg-blue-500/20 text-blue-300 text-xs rounded-full">
                          {plugin.category}
                        </span>
                      </div>
                    </div>

                    {/* Learn More */}
                    <div className="absolute bottom-6 right-6 opacity-0 group-hover/card:opacity-100 transition-opacity duration-300">
                      <span className="text-purple-400 font-semibold flex items-center gap-2">
                        View Details
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M9 5l7 7-7 7"
                          />
                        </svg>
                      </span>
                    </div>
                  </Link>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Mobile/Tablet Grid View */}
        <div className="lg:hidden max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {plugins
              .slice()
              .sort((a, b) => b.downloads - a.downloads)
              .slice(0, 6)
              .map((plugin) => (
                <Link
                  key={plugin.id}
                  href={`/marketplace/${plugin.slug}`}
                  className="relative group bg-gray-800/50 backdrop-blur-sm border border-gray-700/50 rounded-xl p-6 overflow-hidden transition-all duration-300 hover:shadow-2xl hover:shadow-purple-500/30 hover:border-purple-500/50"
                >
                  {/* Badge for pricing */}
                  <div className="absolute top-4 right-4">
                    {plugin.pricing.type === "free" ? (
                      <span className="px-3 py-1 bg-green-500/20 text-green-300 text-xs font-semibold rounded-full">
                        {t("plugin.badge.free")}
                      </span>
                    ) : (
                      <span className="px-3 py-1 bg-purple-500/20 text-purple-300 text-xs font-semibold rounded-full">
                        ${plugin.pricing.amount}
                      </span>
                    )}
                  </div>

                  <div className="transition-all duration-300 group-hover:-translate-y-2">
                    {/* Icon */}
                    <div className="text-5xl mb-3 group-hover:scale-110 transition-transform duration-300">
                      {plugin.icon}
                    </div>

                    {/* Title */}
                    <h3 className="text-xl font-bold text-white mb-2 group-hover:text-purple-400 transition-colors">
                      {plugin.name}
                    </h3>

                    {/* Description */}
                    <p className="text-gray-300 text-sm mb-3 line-clamp-2">
                      {plugin.tagline}
                    </p>

                    {/* Stats */}
                    <div className="flex items-center gap-3 text-xs text-gray-400 mb-3">
                      <div className="flex items-center gap-1">
                        <svg
                          className="w-3 h-3 text-yellow-500"
                          fill="currentColor"
                          viewBox="0 0 20 20"
                        >
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                        <span>{plugin.rating}</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <svg
                          className="w-3 h-3"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                          />
                        </svg>
                        <span>{plugin.downloads.toLocaleString()}</span>
                      </div>
                    </div>

                    {/* Category Badge */}
                    <div>
                      <span className="inline-block px-2 py-1 bg-blue-500/20 text-blue-300 text-xs rounded-full">
                        {plugin.category}
                      </span>
                    </div>
                  </div>
                </Link>
              ))}
          </div>
        </div>
      </section>

      {/* Browse All Plugins Section */}
      <section className="relative pt-12 pb-20 overflow-hidden">
        <div className="absolute inset-0 z-0">
          <StarField density="medium" showComets={true} cometCount={3} />
          <GridLines />
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="text-center mb-8">
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-2">
              {t("browse.title")}
            </h2>
            <p className="text-gray-400">
              {t("browse.subtitle")}
            </p>
          </div>

          {/* Search and Filters */}
          <div className="mb-12">
            <div className="flex flex-col md:flex-row gap-4">
              {/* Search Bar */}
              <div className="flex-1 relative">
                <input
                  type="text"
                  placeholder={t("browse.searchPlaceholder")}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full px-6 py-4 bg-gray-800/50 border border-gray-700 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent backdrop-blur-sm"
                />
                <svg
                  className="absolute right-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                  />
                </svg>
              </div>

              {/* Category Filter */}
              <div className="relative" ref={categoryRef}>
                <button
                  onClick={() => {
                    setIsCategoryOpen(!isCategoryOpen);
                    setIsPricingOpen(false);
                  }}
                  className="w-full md:w-48 px-6 py-4 bg-gray-800/90 backdrop-blur-md border border-gray-700/50 rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent cursor-pointer hover:bg-gray-700/90 transition-all duration-200 shadow-lg flex items-center justify-between"
                >
                  <span>{selectedCategory}</span>
                  <svg
                    className={`w-5 h-5 transition-transform duration-200 ${
                      isCategoryOpen ? "rotate-180" : ""
                    }`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                {isCategoryOpen && (
                  <div className="absolute z-50 mt-2 w-full md:w-64 bg-gray-800/95 backdrop-blur-md rounded-xl shadow-2xl py-2 ring-1 ring-gray-700/50 max-h-96 overflow-y-auto scrollbar-hide">
                    {categories.map((cat) => (
                      <button
                        key={cat}
                        onClick={() => {
                          setSelectedCategory(cat);
                          setIsCategoryOpen(false);
                        }}
                        className={`w-full text-left px-4 py-2 text-sm transition-all duration-200 ${
                          selectedCategory === cat
                            ? "bg-purple-600/30 text-purple-300"
                            : "text-gray-300 hover:bg-gray-700/50 hover:text-white"
                        }`}
                      >
                        {cat}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              {/* Pricing Filter */}
              <div className="relative" ref={pricingRef}>
                <button
                  onClick={() => {
                    setIsPricingOpen(!isPricingOpen);
                    setIsCategoryOpen(false);
                  }}
                  className="w-full md:w-48 px-6 py-4 bg-gray-800/90 backdrop-blur-md border border-gray-700/50 rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent cursor-pointer hover:bg-gray-700/90 transition-all duration-200 shadow-lg flex items-center justify-between"
                >
                  <span>
                    {selectedPricing === "All"
                      ? t("browse.pricingFilter.all")
                      : selectedPricing === "free"
                      ? t("browse.pricingFilter.free")
                      : selectedPricing === "monthly"
                      ? t("browse.pricingFilter.monthly")
                      : t("browse.pricingFilter.oneTime")}
                  </span>
                  <svg
                    className={`w-5 h-5 transition-transform duration-200 ${
                      isPricingOpen ? "rotate-180" : ""
                    }`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                {isPricingOpen && (
                  <div className="absolute z-50 mt-2 w-full bg-gray-800/95 backdrop-blur-md rounded-xl shadow-2xl py-2 ring-1 ring-gray-700/50">
                    {[
                      { value: "All", label: "All Pricing" },
                      { value: "free", label: "Free" },
                      { value: "monthly", label: "Monthly" },
                      { value: "one-time", label: "One-time" },
                    ].map((option) => (
                      <button
                        key={option.value}
                        onClick={() => {
                          setSelectedPricing(option.value);
                          setIsPricingOpen(false);
                        }}
                        className={`w-full text-left px-4 py-2 text-sm transition-all duration-200 ${
                          selectedPricing === option.value
                            ? "bg-purple-600/30 text-purple-300"
                            : "text-gray-300 hover:bg-gray-700/50 hover:text-white"
                        }`}
                      >
                        {option.label}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Results Count */}
          <div className="mb-6">
            <p className="text-gray-400">
              {t("browse.showing")} {filteredPlugins.length} {t("browse.of")} {plugins.length} {t("browse.plugins")}
            </p>
          </div>

          {/* Plugin Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
            {paginatedPlugins.map((plugin) => (
              <div
                key={plugin.id}
                className="group bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-xl overflow-hidden hover:border-purple-500/50 transition-all duration-300 hover:shadow-2xl hover:shadow-purple-500/20 hover:-translate-y-1"
              >
                <div className="p-6">
                  {/* Plugin Icon & Name */}
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center gap-3">
                      <div className="text-5xl group-hover:scale-110 transition-transform duration-300">
                        {plugin.icon}
                      </div>
                      <div>
                        <h3 className="text-xl font-bold text-white group-hover:text-purple-400 transition-colors">
                          {plugin.name}
                        </h3>
                        <p className="text-sm text-gray-400">
                          v{plugin.version}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Tagline */}
                  <p className="text-gray-300 mb-4 line-clamp-2">
                    {plugin.tagline}
                  </p>

                  {/* Category Badge */}
                  <div className="mb-4">
                    <span className="inline-block px-3 py-1 bg-purple-500/20 text-purple-300 text-xs rounded-full">
                      {plugin.category}
                    </span>
                  </div>

                  {/* Stats */}
                  <div className="flex items-center gap-4 mb-4 text-sm text-gray-400">
                    <div className="flex items-center gap-1">
                      <svg
                        className="w-4 h-4 text-yellow-500"
                        fill="currentColor"
                        viewBox="0 0 20 20"
                      >
                        <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                      </svg>
                      <span>{plugin.rating}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <svg
                        className="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                        />
                      </svg>
                      <span>{plugin.downloads.toLocaleString()}</span>
                    </div>
                  </div>

                  {/* Pricing & CTA */}
                  <div className="flex items-center justify-between pt-4 border-t border-gray-700/50">
                    <div>
                      {plugin.pricing.type === "free" ? (
                        <span className="text-green-400 font-semibold">
                          {t("plugin.free")}
                        </span>
                      ) : (
                        <div>
                          <span className="text-white font-semibold text-lg">
                            ${plugin.pricing.amount}
                          </span>
                          <span className="text-gray-400 text-sm ml-1">
                            {plugin.pricing.type === "monthly"
                              ? t("plugin.monthly")
                              : t("plugin.oneTime")}
                          </span>
                        </div>
                      )}
                    </div>
                    <Link
                      href={`/marketplace/${plugin.slug}`}
                      className="px-4 py-2 bg-gradient-to-r from-purple-600 to-pink-600 text-white rounded-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 hover:shadow-lg hover:shadow-purple-500/50"
                    >
                      {t("plugin.viewDetails")}
                    </Link>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Pagination */}
          {filteredPlugins.length > 0 && totalPages > 1 && (
            <div className="flex justify-center items-center gap-2 mt-12">
              <button
                onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                disabled={currentPage === 1}
                className="px-4 py-2 bg-gray-800/50 border border-gray-700/50 rounded-lg text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-700/50 transition-all duration-300"
              >
                {t("browse.pagination.previous")}
              </button>
              
              <div className="flex gap-2">
                {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
                  <button
                    key={page}
                    onClick={() => setCurrentPage(page)}
                    className={`px-4 py-2 rounded-lg transition-all duration-300 ${
                      currentPage === page
                        ? 'bg-gradient-to-r from-purple-600 to-pink-600 text-white'
                        : 'bg-gray-800/50 border border-gray-700/50 text-gray-300 hover:bg-gray-700/50'
                    }`}
                  >
                    {page}
                  </button>
                ))}
              </div>

              <button
                onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                disabled={currentPage === totalPages}
                className="px-4 py-2 bg-gray-800/50 border border-gray-700/50 rounded-lg text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-700/50 transition-all duration-300"
              >
                {t("browse.pagination.next")}
              </button>
            </div>
          )}

          {/* No Results */}
          {filteredPlugins.length === 0 && (
            <div className="text-center py-20">
              <div className="text-6xl mb-4">🔍</div>
              <h3 className="text-2xl font-bold text-white mb-2">
                {t("browse.noResults.title")}
              </h3>
              <p className="text-gray-400">
                {t("browse.noResults.subtitle")}
              </p>
            </div>
          )}
        </div>
      </section>

      <Footer />
    </main>
  );
}
