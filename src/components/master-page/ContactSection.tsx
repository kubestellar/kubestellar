"use client";

import {useState, useEffect } from "react";
import StarField from "../StarField";

export default function ContactSection() {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    subject: '',
    message: '',
    privacy: false
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.privacy) {
      alert('Please agree to the privacy policy to continue.');
      return;
    }

    setIsSubmitting(true);

    try {
      // Create mailto link for Google Groups mailing list
      const subject = encodeURIComponent(`[KubeStellar] ${formData.subject}`);
      const body = encodeURIComponent(
        `Hi KubeStellar Community,

Name: ${formData.name}
Email: ${formData.email}
Subject: ${formData.subject}

Message:
${formData.message}

Best regards,
${formData.name}

---
This message was sent via the KubeStellar website contact form.
Google Groups: https://groups.google.com/g/kubestellar-dev`
      );
      
      const mailtoLink = `mailto:kubestellar-dev@googlegroups.com?subject=${subject}&body=${body}`;
      
      // Small delay to show loading state
      await new Promise(resolve => setTimeout(resolve, 800));
      
      // Open email client
      window.location.href = mailtoLink;
      
      // Show success message
      setShowSuccess(true);
      
      // Reset form
      setFormData({
        name: '',
        email: '',
        subject: '',
        message: '',
        privacy: false
      });
      
      // Hide success message after 8 seconds
      setTimeout(() => setShowSuccess(false), 8000);
      
    } catch (error) {
      console.error('Error submitting form:', error);
      alert('There was an error opening your email client. Please try again or visit https://groups.google.com/g/kubestellar-dev directly.');
    } finally {
      setIsSubmitting(false);
    }
  };
  useEffect(() => {
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

    const gridContainer = document.getElementById("grid-lines-contact");

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
      <StarField density="medium" showComets={true} cometCount={4} />

      {/* Grid lines background */}
      <div id="grid-lines-contact" className="absolute inset-0 opacity-20"></div>

      {/* Background elements */}
      <div className="absolute inset-0 z-0 overflow-hidden">
        <div className="absolute top-0 right-0 w-1/3 h-1/3 bg-gradient-to-b from-purple-500/10 to-transparent rounded-full blur-3xl transform translate-x-1/3 -translate-y-1/3"></div>
        <div className="absolute bottom-0 left-0 w-1/3 h-1/3 bg-gradient-to-t from-blue-500/10 to-transparent rounded-full blur-3xl transform -translate-x-1/3 translate-y-1/3"></div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-extrabold text-white sm:text-4xl">
            <span className="text-gradient">Get in Touch</span>
          </h2>
          <p className="mt-4 max-w-2xl mx-auto text-xl text-gray-300">
            Have questions about KubeStellar? We&apos;re here to help!
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8 items-stretch">
          {/* Left side: Contact info cards */}
          <div className="lg:col-span-2 flex flex-col justify-between space-y-4">
            {/* Contact card 1 */}
            <a href="mailto:info@kubestellar.io" className="block bg-gray-800/50 backdrop-blur-md rounded-xl shadow-sm border border-gray-700/50 p-6 transform transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-blue-500/50 cursor-pointer">
              <div className="flex items-center">
                <div className="flex-shrink-0 h-12 w-12 rounded-full bg-blue-900/30 flex items-center justify-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-6 w-6 text-blue-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                    />
                  </svg>
                </div>
                <div className="ml-5">
                  <h3 className="text-lg font-medium text-white">Email Support</h3>
                  <p className="text-gray-300 mt-1">Get direct support from our team</p>
                  <p className="text-blue-400 mt-1 text-sm inline-flex items-center">
                    support@kubestellar.io
                    <svg className="ml-1 w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                    </svg>
                  </p>
                </div>
              </div>
            </a>

            {/* Contact card 2 */}
            <a href="https://kubestellar.slack.com" target="_blank" rel="noopener noreferrer" className="block bg-gray-800/50 backdrop-blur-md rounded-xl shadow-sm border border-transparent p-6 transform transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-purple-500/70 cursor-pointer">
              <div className="flex items-center">
                <div className="flex-shrink-0 h-12 w-12 rounded-full bg-purple-600 flex items-center justify-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-6 w-6 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                    />
                  </svg>
                </div>
                <div className="ml-5">
                  <h3 className="text-lg font-medium text-white">Community Chat</h3>
                  <p className="text-gray-300 mt-1">Join our Slack workspace for real-time support</p>
                  <p className="text-purple-400 mt-1 text-sm inline-flex items-center">
                    Join Slack
                    <svg className="ml-1 w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                    </svg>
                  </p>
                </div>
              </div>
            </a>

            {/* Contact card 3 */}
            <a href="https://github.com/kubestellar/kubestellar" target="_blank" rel="noopener noreferrer" className="block bg-gray-800/50 backdrop-blur-md rounded-xl shadow-sm border border-gray-700/50 p-6 transform transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-green-500/50 cursor-pointer">
              <div className="flex items-center">
                <div className="flex-shrink-0 h-12 w-12 rounded-full bg-green-600 flex items-center justify-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-6 w-6 text-white"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                  </svg>
                </div>
                <div className="ml-5">
                  <h3 className="text-lg font-medium text-white">GitHub</h3>
                  <p className="text-gray-300 mt-1">Contribute, report issues, or browse the source code</p>
                  <p className="text-green-400 mt-1 text-sm inline-flex items-center">
                    View Repository
                    <svg className="ml-1 w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                    </svg>
                  </p>
                </div>
              </div>
            </a>
            {/* Contact card 4 - LinkedIn */}
            <a href="https://linkedin.com/company/kubestellar" target="_blank" rel="noopener noreferrer" className="block bg-gray-800/50 backdrop-blur-md rounded-xl shadow-sm border border-gray-700/50 p-6 transform transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-blue-500/50 cursor-pointer">
              <div className="flex items-center">
                <div className="flex-shrink-0 h-12 w-12 rounded-full bg-blue-600 flex items-center justify-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-6 w-6 text-white"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
                  </svg>
                </div>
                <div className="ml-5">
                  <h3 className="text-lg font-medium text-white">LinkedIn</h3>
                  <p className="text-gray-300 mt-1">Connect with our professional community</p>
                  <p className="text-blue-400 mt-1 text-sm inline-flex items-center">
                    Follow Us
                    <svg className="ml-1 w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                    </svg>
                  </p>
                </div>
              </div>
            </a>

          </div>

          {/* Right side: Contact form */}
          <div className="lg:col-span-3 flex flex-col">
            <div className="bg-gradient-to-br from-gray-800/60 to-gray-900/60 backdrop-blur-md rounded-2xl shadow-2xl border border-gray-700/50 overflow-hidden h-full flex flex-col">
              <div className="p-6 sm:p-8 flex-1 flex flex-col">
                <h3 className="text-2xl font-bold text-white mb-6 text-center">
                  Send us a message
                </h3>

                <form onSubmit={handleSubmit} className="space-y-4 flex-1 flex flex-col">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label
                        htmlFor="name"
                        className="block text-sm font-semibold text-gray-300 mb-2"
                      >
                        Name *
                      </label>
                      <input
                        type="text"
                        id="name"
                        name="name"
                        value={formData.name}
                        onChange={handleInputChange}
                        required
                        className="w-full px-4 py-3 bg-gray-700/60 border border-gray-600/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-300 backdrop-blur-sm"
                        placeholder="Your full name"
                      />
                    </div>

                    <div>
                      <label
                        htmlFor="email"
                        className="block text-sm font-semibold text-gray-300 mb-2"
                      >
                        Email *
                      </label>
                      <input
                        type="email"
                        id="email"
                        name="email"
                        value={formData.email}
                        onChange={handleInputChange}
                        required
                        className="w-full px-4 py-3 bg-gray-700/60 border border-gray-600/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-300 backdrop-blur-sm"
                        placeholder="you@example.com"
                      />
                    </div>
                  </div>

                  <div>
                    <label
                      htmlFor="subject"
                      className="block text-sm font-semibold text-gray-300 mb-2"
                    >
                      Subject *
                    </label>
                    <div className="relative">
                      <select
                        id="subject"
                        name="subject"
                        value={formData.subject}
                        onChange={handleInputChange}
                        required
                        className="w-full px-4 py-3 pr-12 bg-gray-700/60 border border-gray-600/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-300 backdrop-blur-sm appearance-none cursor-pointer hover:border-gray-500/60 hover:bg-gray-700/70"
                      >
                        <option value="" disabled className="text-gray-400">
                          Select a subject
                        </option>
                        <option value="General Inquiry" className="bg-gray-800 text-white py-2">
                          General Inquiry
                        </option>
                        <option value="Technical Support" className="bg-gray-800 text-white py-2">
                          Technical Support
                        </option>
                        <option value="Partnership" className="bg-gray-800 text-white py-2">
                          Partnership
                        </option>
                        <option value="Documentation Feedback" className="bg-gray-800 text-white py-2">
                          Documentation Feedback
                        </option>
                        <option value="Enterprise Solutions" className="bg-gray-800 text-white py-2">
                          Enterprise Solutions
                        </option>
                        <option value="Other" className="bg-gray-800 text-white py-2">
                          Other
                        </option>
                      </select>
                      {/* Custom dropdown chevron */}
                      <div className="absolute inset-y-0 right-0 flex items-center pr-4 pointer-events-none">
                        <svg
                          className="w-5 h-5 text-gray-400 transition-transform duration-200"
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
                      </div>
                    </div>
                  </div>

                  <div className="flex-1 flex flex-col">
                    <label
                      htmlFor="message"
                      className="block text-sm font-semibold text-gray-300 mb-2"
                    >
                      Message *
                    </label>
                    <textarea
                      id="message"
                      name="message"
                      value={formData.message}
                      onChange={handleInputChange}
                      required
                      className="w-full px-4 py-3 bg-gray-700/60 border border-gray-600/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-300 backdrop-blur-sm resize-none flex-1 min-h-[120px]"
                      placeholder="Tell us about your use case and how we can help..."
                    ></textarea>
                  </div>

                  <div className="flex items-start space-x-3">
                    <input
                      id="privacy"
                      name="privacy"
                      type="checkbox"
                      checked={formData.privacy}
                      onChange={handleInputChange}
                      required
                      className="mt-1 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded transition-all duration-200"
                    />
                    <label
                      htmlFor="privacy"
                      className="text-sm text-gray-300 leading-relaxed"
                    >
                      I agree to the{" "}
                      <a
                        href="#"
                        className="text-blue-400 hover:text-blue-300 underline transition-colors duration-200"
                      >
                        privacy policy
                      </a>{" "}
                      and consent to being contacted by the KubeStellar team.
                    </label>
                  </div>

                  <div className="pt-3">
                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className="w-full py-3 px-6 bg-gradient-to-r from-blue-600 via-purple-600 to-blue-700 hover:from-blue-700 hover:via-purple-700 hover:to-blue-800 disabled:from-gray-600 disabled:to-gray-700 rounded-xl font-bold text-white shadow-lg hover:shadow-xl transition-all duration-300 transform hover:scale-[1.02] focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900"
                    >
                      {isSubmitting ? (
                        <div className="flex items-center justify-center space-x-2">
                          <div className="animate-spin h-5 w-5 border-2 border-white border-t-transparent rounded-full"></div>
                          <span>Sending...</span>
                        </div>
                      ) : (
                        <div className="flex items-center justify-center space-x-2">
                          <span>Send Message</span>
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
                              d="M14 5l7 7m0 0l-7 7m7-7H3"
                            />
                          </svg>
                        </div>
                      )}
                    </button>
                  </div>
                </form>

                {/* Success Message */}
                {showSuccess && (
                  <div className="mt-4 rounded-xl bg-green-900/30 p-4 border border-green-500/30 backdrop-blur-sm">
                    <div className="flex">
                      <div className="flex-shrink-0">
                        <svg
                          className="h-5 w-5 text-green-400"
                          xmlns="http://www.w3.org/2000/svg"
                          viewBox="0 0 20 20"
                          fill="currentColor"
                        >
                          <path
                            fillRule="evenodd"
                            d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                            clipRule="evenodd"
                          />
                        </svg>
                      </div>
                      <div className="ml-3">
                        <p className="text-sm font-medium text-green-300">
                          Your email will be sent to the KubeStellar development mailing list. 
                          Please check your email client to complete sending!
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
