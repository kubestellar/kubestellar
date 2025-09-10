"use client";

import { useEffect } from "react";

export default function ContactSection() {
  useEffect(() => {
    // Create starfield and grid for Contact section
    const createStarfield = (container: HTMLElement) => {
      if (!container) return;
      container.innerHTML = "";

      for (let layer = 1; layer <= 3; layer++) {
        const layerDiv = document.createElement("div");
        layerDiv.className = `star-layer layer-${layer}`;
        layerDiv.style.position = "absolute";
        layerDiv.style.inset = "0";
        layerDiv.style.zIndex = layer.toString();

        const starCount = layer === 1 ? 50 : layer === 2 ? 30 : 15;

        for (let i = 0; i < starCount; i++) {
          const star = document.createElement("div");
          star.style.position = "absolute";
          star.style.width = `${Math.random() * 2 + 1}px`;
          star.style.height = star.style.width;
          star.style.backgroundColor = "white";
          star.style.borderRadius = "50%";
          star.style.top = `${Math.random() * 100}%`;
          star.style.left = `${Math.random() * 100}%`;
          star.style.opacity = Math.random().toString();
          star.style.animation = `twinkle ${Math.random() * 3 + 2}s infinite alternate`;
          star.style.animationDelay = `${Math.random() * 2}s`;
          layerDiv.appendChild(star);
        }

        container.appendChild(layerDiv);
      }
    };

    const createGrid = (container: HTMLElement) => {
      if (!container) return;
      container.innerHTML = "";

      const gridSvg = document.createElementNS(
        "http://www.w3.org/2000/svg",
        "svg"
      );
      gridSvg.setAttribute("width", "100%");
      gridSvg.setAttribute("height", "100%");
      gridSvg.style.position = "absolute";
      gridSvg.style.top = "0";
      gridSvg.style.left = "0";

      for (let i = 0; i < 8; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", "0");
        line.setAttribute("y1", `${i * 12}%`);
        line.setAttribute("x2", "100%");
        line.setAttribute("y2", `${i * 12}%`);
        line.setAttribute("stroke", "#6366F1");
        line.setAttribute("stroke-width", "0.5");
        line.setAttribute("stroke-opacity", "0.3");
        gridSvg.appendChild(line);
      }

      for (let i = 0; i < 8; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", `${i * 12}%`);
        line.setAttribute("y1", "0");
        line.setAttribute("x2", `${i * 12}%`);
        line.setAttribute("y2", "100%");
        line.setAttribute("stroke", "#6366F1");
        line.setAttribute("stroke-width", "0.5");
        line.setAttribute("stroke-opacity", "0.3");
        gridSvg.appendChild(line);
      }

      container.appendChild(gridSvg);
    };

    const starsContainer = document.getElementById("stars-container-contact");
    const gridContainer = document.getElementById("grid-lines-contact");

    if (starsContainer) createStarfield(starsContainer);
    if (gridContainer) createGrid(gridContainer);
  }, []);

  return (
    <section
      id="contact"
      className="relative py-20 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden"
    >
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]"></div>

      {/* Starfield background */}
      <div
        id="stars-container-contact"
        className="absolute inset-0 overflow-hidden"
      ></div>

      {/* Grid lines background */}
      <div
        id="grid-lines-contact"
        className="absolute inset-0 opacity-20"
      ></div>

      {/* Background elements */}
      <div className="absolute inset-0 z-0 overflow-hidden">
        <div className="absolute top-0 right-0 w-1/3 h-1/3 bg-gradient-to-b from-purple-500/10 to-transparent rounded-full blur-3xl transform translate-x-1/3 -translate-y-1/3"></div>
        <div className="absolute bottom-0 left-0 w-1/3 h-1/3 bg-gradient-to-t from-blue-500/10 to-transparent rounded-full blur-3xl transform -translate-x-1/3 translate-y-1/3"></div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-extrabold text-white sm:text-4xl">
            Get in <span className="text-gradient">Touch</span>
          </h2>
          <p className="mt-4 max-w-2xl mx-auto text-xl text-gray-300">
            Ready to transform your multi-cluster operations? Let&apos;s connect
            and explore how KubeStellar can help.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8 items-start">
          {/* Contact Information */}
          <div className="lg:col-span-2 space-y-8">
            <div className="space-y-6">
              <div className="flex items-start space-x-4 group">
                <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                  <svg
                    className="w-6 h-6 text-white"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                    />
                  </svg>
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-white">
                    Community Chat
                  </h3>
                  <p className="text-gray-300">
                    Join our Slack workspace for real-time support
                  </p>
                  <a
                    href="#"
                    className="inline-flex items-center mt-2 text-blue-400 hover:text-blue-300 transition-colors"
                  >
                    Join Slack
                    <svg
                      className="w-4 h-4 ml-1"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                      />
                    </svg>
                  </a>
                </div>
              </div>

              <div className="flex items-start space-x-4 group">
                <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-green-500 to-green-600 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                  <svg
                    className="w-6 h-6 text-white"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M12 0C5.374 0 0 5.373 0 12 0 17.302 3.438 21.8 8.207 23.387c.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.300 24 12c0-6.627-5.373-12-12-12z" />
                  </svg>
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-white">GitHub</h3>
                  <p className="text-gray-300">
                    Contribute, report issues, or browse the source code
                  </p>
                  <a
                    href="#"
                    className="inline-flex items-center mt-2 text-green-400 hover:text-green-300 transition-colors"
                  >
                    View Repository
                    <svg
                      className="w-4 h-4 ml-1"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                      />
                    </svg>
                  </a>
                </div>
              </div>

              <div className="flex items-start space-x-4 group">
                <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                  <svg
                    className="w-6 h-6 text-white"
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
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-white">
                    Email Support
                  </h3>
                  <p className="text-gray-300">
                    Get direct support from our team
                  </p>
                  <a
                    href="mailto:support@kubestellar.io"
                    className="inline-flex items-center mt-2 text-purple-400 hover:text-purple-300 transition-colors"
                  >
                    support@kubestellar.io
                    <svg
                      className="w-4 h-4 ml-1"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                      />
                    </svg>
                  </a>
                </div>
              </div>
            </div>

            {/* Community Stats */}
            <div className="bg-gradient-to-br from-gray-800/50 to-gray-900/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50">
              <h3 className="text-lg font-semibold text-white mb-4">
                Community
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center">
                  <div
                    className="text-2xl font-bold text-blue-400 counter"
                    data-target="2500"
                  >
                    0
                  </div>
                  <div className="text-sm text-gray-400">GitHub Stars</div>
                </div>
                <div className="text-center">
                  <div
                    className="text-2xl font-bold text-green-400 counter"
                    data-target="150"
                  >
                    0
                  </div>
                  <div className="text-sm text-gray-400">Contributors</div>
                </div>
                <div className="text-center">
                  <div
                    className="text-2xl font-bold text-purple-400 counter"
                    data-target="500"
                  >
                    0
                  </div>
                  <div className="text-sm text-gray-400">Slack Members</div>
                </div>
                <div className="text-center">
                  <div
                    className="text-2xl font-bold text-yellow-400 counter"
                    data-target="50"
                  >
                    0
                  </div>
                  <div className="text-sm text-gray-400">Releases</div>
                </div>
              </div>
            </div>
          </div>

          {/* Contact Form */}
          <div className="lg:col-span-3">
            <div className="bg-gradient-to-br from-gray-800/50 to-gray-900/50 backdrop-blur-sm rounded-xl p-8 border border-gray-700/50">
              <h3 className="text-xl font-semibold text-white mb-6">
                Send us a message
              </h3>
              <form className="space-y-6">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <div>
                    <label
                      htmlFor="first-name"
                      className="block text-sm font-medium text-gray-300 mb-2"
                    >
                      First Name
                    </label>
                    <input
                      type="text"
                      id="first-name"
                      name="first-name"
                      className="w-full px-4 py-3 bg-gray-700/50 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-300"
                      placeholder="John"
                    />
                  </div>
                  <div>
                    <label
                      htmlFor="last-name"
                      className="block text-sm font-medium text-gray-300 mb-2"
                    >
                      Last Name
                    </label>
                    <input
                      type="text"
                      id="last-name"
                      name="last-name"
                      className="w-full px-4 py-3 bg-gray-700/50 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-300"
                      placeholder="Doe"
                    />
                  </div>
                </div>

                <div>
                  <label
                    htmlFor="email"
                    className="block text-sm font-medium text-gray-300 mb-2"
                  >
                    Email
                  </label>
                  <input
                    type="email"
                    id="email"
                    name="email"
                    className="w-full px-4 py-3 bg-gray-700/50 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-300"
                    placeholder="john@example.com"
                  />
                </div>

                <div>
                  <label
                    htmlFor="company"
                    className="block text-sm font-medium text-gray-300 mb-2"
                  >
                    Company (Optional)
                  </label>
                  <input
                    type="text"
                    id="company"
                    name="company"
                    className="w-full px-4 py-3 bg-gray-700/50 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-300"
                    placeholder="Your Company"
                  />
                </div>

                <div>
                  <label
                    htmlFor="subject"
                    className="block text-sm font-medium text-gray-300 mb-2"
                  >
                    Subject
                  </label>
                  <select
                    id="subject"
                    name="subject"
                    className="w-full px-4 py-3 bg-gray-700/50 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-300"
                  >
                    <option value="">Select a subject</option>
                    <option value="general">General Inquiry</option>
                    <option value="technical">Technical Support</option>
                    <option value="partnership">Partnership</option>
                    <option value="enterprise">Enterprise Solutions</option>
                    <option value="other">Other</option>
                  </select>
                </div>

                <div>
                  <label
                    htmlFor="message"
                    className="block text-sm font-medium text-gray-300 mb-2"
                  >
                    Message
                  </label>
                  <textarea
                    id="message"
                    name="message"
                    rows={5}
                    className="w-full px-4 py-3 bg-gray-700/50 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-300 resize-none"
                    placeholder="Tell us about your use case and how we can help..."
                  ></textarea>
                </div>

                <button
                  type="submit"
                  className="w-full px-8 py-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 rounded-lg font-semibold text-lg transition-all duration-300 transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900"
                >
                  Send Message
                </button>
              </form>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
