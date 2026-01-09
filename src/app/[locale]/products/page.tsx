"use client";

import Image from "next/image";
import { useEffect, useState } from "react";
import {
  Navbar,
  Footer,
  GridLines,
  StarField,
  PageActions,
} from "@/components";
import { useTranslations } from "next-intl";

interface Product {
  id: string;
  logo: string;
  website: string;
  repository: string;
  name: string;
  fullName: string;
  description: string;
  demoVideo?: string;
  hasDemo?: boolean;
}

export default function ProductsPage() {
  const t = useTranslations("productsPage");
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);

  // Product data with translatable strings
  const products: Product[] = [
    {
      id: "kubestellar",
      logo: "/products/kubestellar.png",
      website: "https://kubestellar.io",
      repository: "https://github.com/kubestellar/kubestellar",
      name: t("products.kubestellar.name"),
      fullName: t("products.kubestellar.fullName"),
      description: t("products.kubestellar.description"),
    },
    {
      id: "kubestellar-ui",
      logo: "/products/ui.png",
      website: "https://ui.kubestellar.io",
      repository: "https://github.com/kubestellar/ui",
      name: t("products.kubestellarUI.name"),
      fullName: t("products.kubestellarUI.fullName"),
      description: t("products.kubestellarUI.description"),
    },
    {
      id: "kubeflex",
      logo: "/products/kubeflex.png",
      website: "https://kubestellar.io/docs/what-is-kubestellar/related/kubeflex",
      repository: "https://github.com/kubestellar/kubeflex",
      name: t("products.kubeflex.name"),
      fullName: t("products.kubeflex.fullName"),
      description: t("products.kubeflex.description"),
    },
    {
      id: "a2a",
      logo: "/products/a2a.png",
      website: "https://kubestellar.github.io/a2a/",
      repository: "https://github.com/kubestellar/a2a",
      name: t("products.a2a.name"),
      fullName: t("products.a2a.fullName"),
      description: t("products.a2a.description"),
    },
    {
      id: "kubectl-multi",
      logo: "/products/kubectl-multi.png",
      website: "https://github.com/kubestellar/kubectl-multi-plugin",
      repository: "https://github.com/kubestellar/kubectl-multi-plugin",
      name: t("products.kubectlMulti.name"),
      fullName: t("products.kubectlMulti.fullName"),
      description: t("products.kubectlMulti.description"),
      hasDemo: true,
      demoVideo:
        "https://www.youtube.com/embed/YtocfNSKqgI?si=SJc798MuZ2o9LeP_",
    },
    {
      id: "galaxy-marketplace",
      logo: "/products/galaxy.png",
      website: "https://galaxy.kubestellar.io",
      repository: "https://github.com/kubestellar/ui-plugins",
      name: t("products.galaxyMarketplace.name"),
      fullName: t("products.galaxyMarketplace.fullName"),
      description: t("products.galaxyMarketplace.description"),
    },
  ];

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
      .product-card:hover {
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
      <Navbar />
      <PageActions position="fixed" />

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
        <div className="relative z-10 text-center px-2 sm:px-4 pt-16 sm:pt-20 pb-2">
          <h1 className="text-3xl sm:text-4xl md:text-6xl font-extrabold tracking-tighter text-shadow-lg">
            {t("title")} <span className="text-gradient">{t("titleSpan")}</span>
          </h1>
          <p className="mt-4 max-w-2xl mx-auto text-base sm:text-lg md:text-xl text-gray-300">
            {t("subtitle")}
          </p>
        </div>
      </section>

      {/* Products Section */}
      <section
        id="products"
        className="relative pt-6 sm:pt-8 pb-16 sm:pb-24 z-10"
      >
        <div className="max-w-7xl mx-auto px-2 sm:px-4 lg:px-8 relative">
          <div className="grid gap-6 sm:gap-10 grid-cols-1 md:grid-cols-2 lg:grid-cols-2">
            {products.map(product => {
              return (
                <div
                  key={product.id}
                  className={`product-card bg-gray-800/50 backdrop-blur-md rounded-xl shadow-lg border border-gray-700/50 p-4 sm:p-8 flex flex-col text-left transition-all duration-300 hover:shadow-2xl hover:-translate-y-2 hover:border-blue-500/50`}
                >
                  {/* Top section: Logo on left, Title on right */}
                  <div className="flex flex-col sm:flex-row items-center mb-4 sm:mb-6">
                    {/* Logo on the left */}
                    <div
                      className={`relative ${product.id === "kubestellar" ? "w-full h-20 sm:h-24" : product.id === "a2a" ? "w-64 sm:w-96 h-20 sm:h-24" : product.id === "galaxy-marketplace" ? "w-28 sm:w-40 h-20 sm:h-28" : product.id === "kubectl-multi" ? "w-28 sm:w-40 h-20 sm:h-28" : "w-24 sm:w-32 h-16 sm:h-24"} ${product.id === "kubestellar" ? "" : "sm:mr-6"} flex items-center justify-center flex-shrink-0 mb-4 sm:mb-0`}
                    >
                      <Image
                        src={product.logo}
                        alt={`${product.name} Logo`}
                        fill
                        className="object-contain"
                        sizes={
                          product.id === "kubestellar"
                            ? "100vw"
                            : product.id === "a2a"
                              ? "(min-width: 640px) 384px, 256px"
                              : product.id === "galaxy-marketplace" ||
                                  product.id === "kubectl-multi"
                                ? "(min-width: 640px) 160px, 112px"
                                : "(min-width: 640px) 128px, 96px"
                        }
                        priority
                      />
                    </div>

                    {/* Title on the right - hidden for kubestellar */}
                    {product.id !== "kubestellar" && (
                      <div className="flex-1 text-center sm:text-left">
                        <h3 className="text-lg sm:text-2xl font-bold text-white">
                          {product.fullName}
                        </h3>
                      </div>
                    )}
                  </div>

                  {/* Bottom section: Description and buttons */}
                  <div className="flex-1">
                    <p className="text-gray-400 mb-4 sm:mb-6 text-sm sm:text-base">
                      {product.description}
                    </p>

                    {/* Action Buttons */}
                    <div className="flex flex-col sm:flex-row gap-3">
                      <a
                        href={product.repository}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex-1 inline-flex items-center justify-center px-4 py-2 border border-gray-600 rounded-lg text-sm font-medium text-gray-300 bg-gray-700/50 hover:bg-gray-600/50 hover:text-white transition-all duration-200 mb-2 sm:mb-0"
                      >
                        <svg
                          className="w-4 h-4 mr-2"
                          fill="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"></path>
                        </svg>
                        {t("repoButton")}
                      </a>
                      {product.hasDemo ? (
                        <button
                          onClick={() => setSelectedProduct(product)}
                          className="flex-1 inline-flex items-center justify-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-all duration-200"
                        >
                          <svg
                            className="w-4 h-4 mr-2"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                            strokeWidth="2"
                          >
                            <polygon points="5 3 19 12 5 21 5 3"></polygon>
                          </svg>
                          {t("watchDemoButton")}
                        </button>
                      ) : (
                        <a
                          href={product.website}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex-1 inline-flex items-center justify-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-all duration-200"
                        >
                          {product.id === "kubestellar-ui" ||
                          product.id === "galaxy-marketplace" ? (
                            <>
                              <svg
                                className="w-4 h-4 mr-2"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                                strokeWidth="2"
                              >
                                <polygon points="5 3 19 12 5 21 5 3"></polygon>
                              </svg>
                              {t("watchDemoButton")}
                            </>
                          ) : (
                            <>
                              <svg
                                className="w-4 h-4 mr-2"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                                strokeWidth="2"
                              >
                                <circle cx="12" cy="12" r="10" />
                                <path d="M2 12h20" />
                                <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                              </svg>
                              {t("websiteButton")}
                            </>
                          )}
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Video Modal */}
      {selectedProduct && selectedProduct.demoVideo && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
          onClick={() => setSelectedProduct(null)}
        >
          <div
            className="relative w-full max-w-5xl bg-gray-900 rounded-2xl overflow-hidden shadow-2xl border border-blue-500/30"
            onClick={e => e.stopPropagation()}
          >
            {/* Close button */}
            <button
              onClick={() => setSelectedProduct(null)}
              className="absolute top-4 right-4 z-10 w-10 h-10 flex items-center justify-center bg-gray-800/90 hover:bg-gray-700 rounded-full transition-colors duration-200 text-white"
            >
              <svg
                className="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                strokeWidth="2"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>

            {/* Video content */}
            <div className="aspect-video">
              <iframe
                className="w-full h-full"
                src={selectedProduct.demoVideo}
                title={`${selectedProduct.fullName} Demo`}
                frameBorder="0"
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                referrerPolicy="strict-origin-when-cross-origin"
                allowFullScreen
              />
            </div>

            {/* Modal footer */}
            <div className="p-6 bg-gray-800/50 border-t border-gray-700/50">
              <h3 className="text-xl font-bold text-white mb-2">
                {selectedProduct.fullName} Demo
              </h3>
              <p className="text-gray-400 text-sm">
                {selectedProduct.description}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Footer */}
      <Footer />
    </div>
  );
}
