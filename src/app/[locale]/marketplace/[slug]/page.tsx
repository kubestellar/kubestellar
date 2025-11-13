"use client";

import React, { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Navbar, Footer } from "@/components";
import { GridLines, StarField } from "@/components/";
import { plugins } from "../plugins";

export default function PluginDetailPage() {
  const params = useParams();
  const router = useRouter();
  const slug = params.slug as string;

  const plugin = plugins.find((p) => p.slug === slug);
  const [showInstallModal, setShowInstallModal] = useState(false);
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [installedSuccess, setInstalledSuccess] = useState(false);
  const [paymentSuccess, setPaymentSuccess] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);

  if (!plugin) {
    return (
      <main className="min-h-screen bg-[#0a0a0a] flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-white mb-4">
            Plugin Not Found
          </h1>
          <Link
            href="/marketplace"
            className="text-purple-400 hover:text-purple-300"
          >
            ← Back to Marketplace
          </Link>
        </div>
      </main>
    );
  }

  const handleInstall = () => {
    // For paid plugins, show payment modal first
    if (plugin.pricing.type !== "free") {
      setShowPaymentModal(true);
    } else {
      // For free plugins, install directly
      setShowInstallModal(true);
      setTimeout(() => {
        setInstalledSuccess(true);
      }, 2000);
    }
  };

  const handlePayment = () => {
    setIsProcessing(true);
    // Simulate payment processing
    setTimeout(() => {
      setPaymentSuccess(true);
      setIsProcessing(false);
      // After payment, start installation
      setTimeout(() => {
        setShowPaymentModal(false);
        setShowInstallModal(true);
        setTimeout(() => {
          setInstalledSuccess(true);
        }, 2000);
      }, 1500);
    }, 3000);
  };

  return (
    <main className="min-h-screen bg-[#0a0a0a]">
      <Navbar />

      {/* Hero Section */}
      <section className="relative pt-32 pb-12 overflow-hidden">
        <div className="absolute inset-0 z-0">
          <StarField density="low" showComets={true} cometCount={2} />
          <GridLines />
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          {/* Breadcrumb */}
          <div className="mb-8">
            <Link
              href="/marketplace"
              className="text-gray-400 hover:text-purple-400 transition-colors inline-flex items-center gap-2"
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
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
              Back to Marketplace
            </Link>
          </div>

          {/* Plugin Header */}
          <div className="bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-8 mb-8">
            <div className="flex flex-col md:flex-row gap-8">
              {/* Icon */}
              <div className="flex-shrink-0">
                <div className="w-32 h-32 flex items-center justify-center bg-gradient-to-br from-purple-500/20 to-pink-500/20 rounded-2xl text-7xl border border-purple-500/30">
                  {plugin.icon}
                </div>
              </div>

              {/* Info */}
              <div className="flex-1">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h1 className="text-4xl font-bold text-white mb-2">
                      {plugin.name}
                    </h1>
                    <p className="text-xl text-gray-300 mb-3">
                      {plugin.tagline}
                    </p>
                    <div className="flex items-center gap-4 text-sm">
                      <span className="px-3 py-1 bg-purple-500/20 text-purple-300 rounded-full">
                        {plugin.category}
                      </span>
                      <span className="text-gray-400">v{plugin.version}</span>
                      <span className="text-gray-400">
                        by {plugin.author}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Stats */}
                <div className="flex items-center gap-6 mb-6">
                  <div className="flex items-center gap-2">
                    <svg
                      className="w-5 h-5 text-yellow-500"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                    </svg>
                    <span className="text-white font-semibold">
                      {plugin.rating}
                    </span>
                    <span className="text-gray-400">/5.0</span>
                  </div>
                  <div className="flex items-center gap-2 text-gray-300">
                    <svg
                      className="w-5 h-5"
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
                    <span>{plugin.downloads.toLocaleString()} downloads</span>
                  </div>
                </div>

                {/* Pricing and Install */}
                <div className="flex items-center gap-4">
                  <div className="flex-1">
                    {plugin.pricing.type === "free" ? (
                      <div className="text-2xl font-bold text-green-400">
                        Free
                      </div>
                    ) : (
                      <div>
                        <span className="text-3xl font-bold bg-gradient-to-r from-purple-400 to-pink-400 bg-clip-text text-transparent">
                          ${plugin.pricing.amount}
                        </span>
                        <span className="text-gray-400 text-lg ml-2">
                          {plugin.pricing.type === "monthly"
                            ? "/month"
                            : "one-time"}
                        </span>
                      </div>
                    )}
                  </div>
                  <button
                    onClick={handleInstall}
                    className="group relative px-8 py-4 bg-gradient-to-r from-purple-600/80 to-pink-600/80 backdrop-blur-xl text-white font-semibold rounded-xl border border-purple-500/30 hover:from-purple-600 hover:to-pink-600 transition-all duration-300 hover:shadow-2xl hover:shadow-purple-500/50 hover:scale-105 hover:border-purple-400/50 overflow-hidden"
                  >
                    <div className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/20 to-white/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700"></div>
                    <span className="relative flex items-center gap-2">
                      {plugin.pricing.type === "free" ? (
                        <>
                          <svg
                            className="w-5 h-5"
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
                          Install Plugin
                        </>
                      ) : (
                        <>
                          <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
                            />
                          </svg>
                          Pay & Install
                        </>
                      )}
                    </span>
                  </button>
                  {plugin.github && (
                    <a
                      href={plugin.github}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="px-6 py-4 bg-gray-700/50 backdrop-blur-sm text-white rounded-xl hover:bg-gray-700 transition-all duration-300 flex items-center gap-2 border border-gray-600/50"
                    >
                      <svg
                        className="w-5 h-5"
                        fill="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
                      </svg>
                      GitHub
                    </a>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Content Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Main Content */}
            <div className="lg:col-span-2 space-y-8">
              {/* Description */}
              <div className="bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-8">
                <h2 className="text-2xl font-bold text-white mb-4">
                  About this plugin
                </h2>
                <div className="prose prose-invert max-w-none">
                  <p className="text-gray-300 whitespace-pre-line leading-relaxed">
                    {plugin.longDescription}
                  </p>
                </div>
              </div>

              {/* Features */}
              <div className="bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-8">
                <h2 className="text-2xl font-bold text-white mb-4">
                  Key Features
                </h2>
                <ul className="space-y-3">
                  {plugin.features.map((feature, index) => (
                    <li key={index} className="flex items-start gap-3">
                      <svg
                        className="w-6 h-6 text-green-400 flex-shrink-0 mt-0.5"
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
                      <span className="text-gray-300">{feature}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>

            {/* Sidebar */}
            <div className="space-y-6">
              {/* Requirements */}
              <div className="bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-6">
                <h3 className="text-xl font-bold text-white mb-4">
                  Requirements
                </h3>
                <ul className="space-y-2">
                  {plugin.requirements.map((req, index) => (
                    <li
                      key={index}
                      className="text-gray-300 text-sm flex items-start gap-2"
                    >
                      <span className="text-purple-400 mt-1">•</span>
                      <span>{req}</span>
                    </li>
                  ))}
                </ul>
              </div>

              {/* Compatibility */}
              <div className="bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-6">
                <h3 className="text-xl font-bold text-white mb-4">
                  Compatibility
                </h3>
                <div className="flex flex-wrap gap-2">
                  {plugin.compatibility.map((platform, index) => (
                    <span
                      key={index}
                      className="px-3 py-1 bg-blue-500/20 text-blue-300 text-sm rounded-full"
                    >
                      {platform}
                    </span>
                  ))}
                </div>
              </div>

              {/* Maintainers */}
              <div className="bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-6">
                <h3 className="text-xl font-bold text-white mb-4">Maintainers</h3>
                <ul className="space-y-2">
                  <li className="text-gray-300 text-sm flex items-start gap-2">
                    <span className="text-purple-400 mt-1">•</span>
                    <span>
                      Andy Anderson -{" "}
                      <a
                        href="https://github.com/pdettori"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-purple-400 hover:text-purple-300 transition-colors underline"
                      >
                        GitHub
                      </a>
                    </span>
                  </li>
                  {/* <li className="text-gray-300 text-sm flex items-start gap-2">
                    <span className="text-purple-400 mt-1">•</span>
                    <span>
                      Mike Spreitzer -{" "}
                      <a
                        href="https://github.com/MikeSpreitzer"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-purple-400 hover:text-purple-300 transition-colors underline"
                      >
                        GitHub
                      </a>
                    </span>
                  </li> */}
                </ul>
              </div>

              {/* Tags */}
              <div className="bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-6">
                <h3 className="text-xl font-bold text-white mb-4">Tags</h3>
                <div className="flex flex-wrap gap-2">
                  {plugin.tags.map((tag, index) => (
                    <span
                      key={index}
                      className="px-3 py-1 bg-gray-700/50 text-gray-300 text-sm rounded-full hover:bg-gray-600/50 transition-colors cursor-pointer"
                    >
                      #{tag}
                    </span>
                  ))}
                </div>
              </div>

              {/* Links */}
              <div className="bg-gray-800/30 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-6">
                <h3 className="text-xl font-bold text-white mb-4">Links</h3>
                <div className="space-y-3">
                  <a
                    href={plugin.documentation}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 text-purple-400 hover:text-purple-300 transition-colors"
                  >
                    <svg
                      className="w-5 h-5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      />
                    </svg>
                    Documentation
                  </a>
                  {plugin.github && (
                    <a
                      href={plugin.github}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 text-purple-400 hover:text-purple-300 transition-colors"
                    >
                      <svg
                        className="w-5 h-5"
                        fill="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
                      </svg>
                      GitHub Repository
                    </a>
                  )}
                  {plugin.website && (
                    <a
                      href={plugin.website}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 text-purple-400 hover:text-purple-300 transition-colors"
                    >
                      <svg
                        className="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
                        />
                      </svg>
                      Official Website
                    </a>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <Footer />

      {/* Payment Modal */}
      {showPaymentModal && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-gray-800/90 backdrop-blur-xl border border-gray-700/50 rounded-2xl p-8 max-w-md w-full shadow-2xl">
            {!paymentSuccess ? (
              <div>
                <div className="text-center mb-6">
                  <div className="w-16 h-16 bg-gradient-to-br from-purple-500/20 to-pink-500/20 rounded-full flex items-center justify-center mx-auto mb-4 border border-purple-500/30">
                    <svg
                      className="w-8 h-8 text-purple-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
                      />
                    </svg>
                  </div>
                  <h3 className="text-2xl font-bold text-white mb-2">
                    Complete Payment
                  </h3>
                  <p className="text-gray-400">
                    Purchase {plugin.name} to get started
                  </p>
                </div>

                {/* Payment Details */}
                <div className="bg-gray-900/50 rounded-xl p-4 mb-6 border border-gray-700/50">
                  <div className="flex justify-between items-center mb-3">
                    <span className="text-gray-400">Plugin</span>
                    <span className="text-white font-semibold">{plugin.name}</span>
                  </div>
                  <div className="flex justify-between items-center mb-3">
                    <span className="text-gray-400">License Type</span>
                    <span className="text-white capitalize">{plugin.pricing.type}</span>
                  </div>
                  <div className="border-t border-gray-700/50 my-3"></div>
                  <div className="flex justify-between items-center">
                    <span className="text-white font-semibold">Total</span>
                    <span className="text-2xl font-bold bg-gradient-to-r from-purple-400 to-pink-400 bg-clip-text text-transparent">
                      ${plugin.pricing.amount}
                    </span>
                  </div>
                </div>

                {/* Mock Payment Form */}
                <div className="space-y-4 mb-6">
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">Card Number</label>
                    <input
                      type="text"
                      placeholder="1234 5678 9012 3456"
                      className="w-full px-4 py-3 bg-gray-900/50 border border-gray-700/50 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                      disabled={isProcessing}
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">Expiry</label>
                      <input
                        type="text"
                        placeholder="MM/YY"
                        className="w-full px-4 py-3 bg-gray-900/50 border border-gray-700/50 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                        disabled={isProcessing}
                      />
                    </div>
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">CVV</label>
                      <input
                        type="text"
                        placeholder="123"
                        className="w-full px-4 py-3 bg-gray-900/50 border border-gray-700/50 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                        disabled={isProcessing}
                      />
                    </div>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex gap-3">
                  <button
                    onClick={() => setShowPaymentModal(false)}
                    className="flex-1 px-6 py-3 bg-gray-700/50 text-white rounded-lg hover:bg-gray-700 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={isProcessing}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handlePayment}
                    className="flex-1 px-6 py-3 bg-gradient-to-r from-purple-600 to-pink-600 text-white font-semibold rounded-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                    disabled={isProcessing}
                  >
                    {isProcessing ? (
                      <>
                        <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                        Processing...
                      </>
                    ) : (
                      <>
                        <svg
                          className="w-5 h-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                          />
                        </svg>
                        Pay Now
                      </>
                    )}
                  </button>
                </div>

                <p className="text-xs text-gray-500 text-center mt-4">
                  🔒 Secure payment powered by KubeStellar Gateway
                </p>
              </div>
            ) : (
              <div className="text-center">
                <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg
                    className="w-10 h-10 text-green-400"
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
                </div>
                <h3 className="text-xl font-bold text-white mb-2">
                  Payment Successful!
                </h3>
                <p className="text-gray-400 mb-4">
                  Starting installation...
                </p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Install Modal */}
      {showInstallModal && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-gray-800/90 backdrop-blur-xl border border-gray-700/50 rounded-2xl p-8 max-w-md w-full shadow-2xl">
            {!installedSuccess ? (
              <div className="text-center">
                <div className="w-16 h-16 border-4 border-purple-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
                <h3 className="text-xl font-bold text-white mb-2">
                  Installing {plugin.name}
                </h3>
                <p className="text-gray-400">Please wait...</p>
              </div>
            ) : (
              <div className="text-center">
                <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-4 border border-green-500/30">
                  <svg
                    className="w-10 h-10 text-green-400"
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
                </div>
                <h3 className="text-xl font-bold text-white mb-2">
                  Successfully Installed!
                </h3>
                <p className="text-gray-400 mb-6">
                  {plugin.name} has been installed to your KubeStellar
                  deployment.
                </p>
                <div className="bg-gray-900/50 rounded-lg p-4 mb-6 text-left border border-gray-700/50">
                  <p className="text-sm text-gray-400 mb-2">
                    Run this command to get started:
                  </p>
                  <code className="text-purple-400 text-sm block break-all">
                    kubectl ks plugin enable {plugin.slug}
                  </code>
                </div>
                <button
                  onClick={() => {
                    setShowInstallModal(false);
                    setInstalledSuccess(false);
                    setPaymentSuccess(false);
                  }}
                  className="w-full px-6 py-3 bg-gradient-to-r from-purple-600 to-pink-600 text-white font-semibold rounded-lg hover:from-purple-700 hover:to-pink-700 transition-all duration-300 hover:shadow-lg hover:shadow-purple-500/50"
                >
                  Close
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </main>
  );
}
