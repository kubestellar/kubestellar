"use client";

import { GridLines, StarField } from "../index";

const Icon = ({
  path,
  className = "h-6 w-6 text-white",
  strokeColor = "currentColor",
  strokeWidth = "2",
}: {
  path: string;
  className?: string;
  strokeColor?: string;
  strokeWidth?: string;
}) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke={strokeColor}
    strokeWidth={strokeWidth}
  >
    <path strokeLinecap="round" strokeLinejoin="round" d={path} />
  </svg>
);

export default function GetStartedSection() {
  return (
    <section
      id="get-started"
      className="relative py-16 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden"
    >
      <div className="absolute inset-0 bg-[#0a0a0a] pointer-events-none"></div>

      {/* Starfield background */}
      <StarField density="medium" showComets={true} cometCount={3} />

      {/* Gridlines background */}
      <GridLines verticalLines={20} horizontalLines={33} />

      <div className="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <div className="absolute top-2/5 left-2/11 w-[6rem] h-[6rem] bg-purple-500/10 rounded-full blur-[120px]"></div>

        <div className="absolute top-4/5 left-1/2 w-[10rem] h-[10rem] bg-purple-500/5 rounded-full blur-[180px]"></div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="text-center">
          <h2 className="text-3xl font-extrabold sm:text-4xl">
            Ready to Get Started?
          </h2>
          <p className="mt-4 max-w-2xl mx-auto text-xl text-blue-100">
            Join the growing community of KubeStellar users and contributors.
          </p>
        </div>

        <div className="mt-12 grid gap-8 md:grid-cols-3">
          {/* Installation Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-xl overflow-hidden border border-white/20 hover:bg-white/20 transition-colors duration-300">
            <div className="p-6 flex flex-col h-full">
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
                <Icon path="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </div>
              <h3 className="text-lg font-bold mb-2">Quick Installation</h3>
              <p className="text-blue-100 mb-6">
                Get up and running with KubeStellar in minutes using our streamlined
                installation guide with prerequisites and step-by-step instructions.
              </p>
              <div className="mt-auto">
                <a
                  href="/quick-installation"
                  className="inline-flex items-center justify-center w-full px-6 py-3 text-base font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors duration-200"
                >
                  Start Quick Installation
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="ml-2 h-5 w-5"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M12.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-2.293-2.293a1 1 0 010-1.414z"
                      clipRule="evenodd"
                    />
                  </svg>
                </a>
                <p className="mt-3 text-sm text-blue-200 text-center">
                  Kubernetes experience required
                </p>
              </div>
            </div>
          </div>

          {/* Community Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-xl overflow-hidden border border-white/20 hover:bg-white/20 transition-colors duration-300">
            <div className="p-6">
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
                <svg
                  width="20"
                  height="22"
                  viewBox="0 0 20 22"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    d="M14.1667 16.6666H18.3334V14.9999C18.3334 14.4804 18.1715 13.9737 17.8702 13.5504C17.5689 13.1272 17.1433 12.8082 16.6524 12.638C16.1615 12.4678 15.6298 12.4548 15.1311 12.6008C14.6325 12.7467 14.1917 13.0444 13.8701 13.4524M14.1667 16.6666H5.83341M14.1667 16.6666V14.9999C14.1667 14.4533 14.0617 13.9308 13.8701 13.4524M13.8701 13.4524C13.5606 12.6791 13.0265 12.0161 12.3367 11.5492C11.6469 11.0822 10.8331 10.8327 10.0001 10.8327C9.1671 10.8327 8.35322 11.0822 7.66342 11.5492C6.97362 12.0161 6.43954 12.6791 6.13008 13.4524M5.83341 16.6666H1.66675V14.9999C1.66679 14.4804 1.82869 13.9737 2.12997 13.5504C2.43124 13.1272 2.8569 12.8082 3.34779 12.638C3.83868 12.4678 4.3704 12.4548 4.86903 12.6008C5.36766 12.7467 5.80844 13.0444 6.13008 13.4524M5.83341 16.6666V14.9999C5.83341 14.4533 5.93841 13.9308 6.13008 13.4524M12.5001 5.83325C12.5001 6.49629 12.2367 7.13218 11.7678 7.60102C11.299 8.06986 10.6631 8.33325 10.0001 8.33325C9.33704 8.33325 8.70115 8.06986 8.23231 7.60102C7.76347 7.13218 7.50008 6.49629 7.50008 5.83325C7.50008 5.17021 7.76347 4.53433 8.23231 4.06549C8.70115 3.59664 9.33704 3.33325 10.0001 3.33325C10.6631 3.33325 11.299 3.59664 11.7678 4.06549C12.2367 4.53433 12.5001 5.17021 12.5001 5.83325ZM17.5001 8.33325C17.5001 8.77528 17.3245 9.1992 17.0119 9.51176C16.6994 9.82432 16.2754 9.99992 15.8334 9.99992C15.3914 9.99992 14.9675 9.82432 14.6549 9.51176C14.3423 9.1992 14.1667 8.77528 14.1667 8.33325C14.1667 7.89122 14.3423 7.4673 14.6549 7.15474C14.9675 6.84218 15.3914 6.66659 15.8334 6.66659C16.2754 6.66659 16.6994 6.84218 17.0119 7.15474C17.3245 7.4673 17.5001 7.89122 17.5001 8.33325ZM5.83341 8.33325C5.83341 8.77528 5.65782 9.1992 5.34526 9.51176C5.0327 9.82432 4.60878 9.99992 4.16675 9.99992C3.72472 9.99992 3.3008 9.82432 2.98824 9.51176C2.67568 9.1992 2.50008 8.77528 2.50008 8.33325C2.50008 7.89122 2.67568 7.4673 2.98824 7.15474C3.3008 6.84218 3.72472 6.66659 4.16675 6.66659C4.60878 6.66659 5.0327 6.84218 5.34526 7.15474C5.65782 7.4673 5.83341 7.89122 5.83341 8.33325Z"
                    stroke="#D1D5DB"
                    strokeWidth="1.66667"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              </div>
              <h3 className="text-lg font-bold mb-2">Join the Community</h3>
              <p className="text-blue-100 mb-4">
                Connect with other users, contributors, and maintainers through
                Slack, forums, and community calls.
              </p>
              <div className="mt-4 grid grid-cols-2 gap-2">
                <a
                  href="https://kubestellar.io/slack"
                  className="flex items-center justify-center p-2 rounded bg-blue-600 hover:bg-blue-700 text-white text-sm"
                >
                  <svg
                    className="w-4 h-4 mr-2"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z" />
                  </svg>
                  Slack
                </a>
                <a
                  href="https://github.com/kubestellar/kubestellar"
                  className="flex items-center justify-center p-2 rounded bg-gray-700 hover:bg-gray-800 text-white text-sm"
                >
                  <svg
                    className="w-4 h-4 mr-2"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                  >
                    <path
                      d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.824 9.128 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
                      clipRule="evenodd"
                    ></path>
                  </svg>
                  GitHub
                </a>
              </div>
            </div>
          </div>

          {/* Documentation Card */}
          <div className="bg-white/10 backdrop-blur-sm rounded-xl overflow-hidden border border-white/20 hover:bg-white/20 transition-colors duration-300">
            <div className="p-6">
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mb-4">
                <Icon path="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </div>
              <h3 className="text-lg font-bold mb-2">Explore Documentation</h3>
              <p className="text-blue-100 mb-4">
                Comprehensive guides, tutorials, and API references to help you
                master KubeStellar&#39;s capabilities.
              </p>
              <div className="mt-4 grid grid-cols-1 gap-2">
                <a
                  href="#"
                  className="block p-2 rounded bg-white/20 hover:bg-white/30 text-white text-sm"
                >
                  Getting Started
                </a>
                <a
                  href="#"
                  className="block p-2 rounded bg-white/20 hover:bg-white/30 text-white text-sm"
                >
                  Tutorials
                </a>
                <a
                  href="#"
                  className="block p-2 rounded bg-white/20 hover:bg-white/30 text-white text-sm"
                >
                  API Reference
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
