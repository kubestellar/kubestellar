"use client";

import { useState } from "react";
import Link from "next/link";
import { GridLines, StarField } from "../index";
import { useTranslations } from "next-intl";
import { getLocalizedUrl } from "@/lib/url";

export default function ContactSection() {
  const t = useTranslations("contactSection");
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    subject: "",
    message: "",
    privacy: false,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);

  const handleInputChange = (
    e: React.ChangeEvent<
      HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
    >
  ) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]:
        type === "checkbox" ? (e.target as HTMLInputElement).checked : value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.privacy) {
      alert("Please agree to the privacy policy to continue.");
      return;
    }

    setIsSubmitting(true);

    try {
      const formPayload = new URLSearchParams({
        "form-name": "contact",
        name: formData.name,
        email: formData.email,
        subject: formData.subject,
        message: formData.message,
      });

      const res = await fetch("/", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: formPayload.toString(),
      });

      if (!res.ok) throw new Error("Form submission failed");

      setShowSuccess(true);
      setFormData({
        name: "",
        email: "",
        subject: "",
        message: "",
        privacy: false,
      });

      setTimeout(() => setShowSuccess(false), 8000);
    } catch (error) {
      console.error("Submission error:", error);
      alert("Submission failed. Please try again later.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <section
      id="contact"
      className="relative py-12 sm:py-16 md:py-20 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden"
    >
      {/* Dark base background */}
      <div className="absolute inset-0 bg-[#0a0a0a]"></div>

      {/* Starfield background */}
      <StarField density="medium" showComets={true} cometCount={4} />

      {/* Grid lines background */}
      <GridLines horizontalLines={21} verticalLines={15} />

      {/* Background elements */}
      <div className="absolute inset-0 z-0 overflow-hidden">
        <div className="absolute top-0 right-0 w-1/3 h-1/3 bg-gradient-to-b from-purple-500/10 to-transparent rounded-full blur-3xl transform translate-x-1/3 -translate-y-1/3"></div>
        <div className="absolute bottom-0 left-0 w-1/3 h-1/3 bg-gradient-to-t from-blue-500/10 to-transparent rounded-full blur-3xl transform -translate-x-1/3 translate-y-1/3"></div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="text-center mb-8 sm:mb-10 md:mb-12">
          <h2 className="text-3xl font-extrabold text-white sm:text-[2.4rem]">
            {t("title")}{" "}
            <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">
              {t("titleSpan")}
            </span>
          </h2>
          <p className="mt-3 sm:mt-4 max-w-2xl mx-auto text-base sm:text-lg md:text-xl text-gray-300 px-4">
            {t("subtitle")}
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6 sm:gap-8 items-stretch">
          {/* Left side: Contact info cards */}
          <div className="lg:col-span-2 flex flex-col justify-between space-y-3 sm:space-y-4">
            {/* Contact card 1 */}
            <a
              href="mailto:info@kubestellar.io"
              className="block bg-gray-800/50 backdrop-blur-md rounded-xl shadow-sm border border-gray-700/50 p-4 sm:p-6 transform transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-blue-500/50 cursor-pointer"
            >
              <div className="flex items-center">
                <div className="flex-shrink-0 h-10 w-10 sm:h-12 sm:w-12 rounded-full bg-blue-900/30 flex items-center justify-center">
                  <svg
                    className="h-6 w-6 sm:h-8 sm:w-8 text-blue-400"
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
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
                <div className="ml-3 sm:ml-5">
                  <h3 className="text-base sm:text-lg font-medium text-white">
                    {t("card1Title")}
                  </h3>
                  <p className="text-gray-300 mt-1 text-sm sm:text-base">
                    {t("card1Description")}
                  </p>
                  <p className="text-blue-400 mt-1 text-xs sm:text-sm inline-flex items-center">
                    {t("card1Link")}
                    <svg
                      className="ml-1 w-3 h-3 sm:w-4 sm:h-4"
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
                  </p>
                </div>
              </div>
            </a>

            {/* Contact card 2 */}
            <a
              href={getLocalizedUrl("https://kubestellar.io/slack")}
              target="_blank"
              rel="noopener noreferrer"
              className="block bg-gray-800/50 backdrop-blur-md rounded-xl shadow-sm border border-transparent p-4 sm:p-6 transform transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-purple-500/70 cursor-pointer"
            >
              <div className="flex items-center">
                <div className="flex-shrink-0 h-10 w-10 sm:h-12 sm:w-12 rounded-full bg-gray-800 flex items-center justify-center shadow-lg">
                  <svg
                    className="h-6 w-6 sm:h-8 sm:w-8"
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 60 60"
                    preserveAspectRatio="xMidYMid meet"
                  >
                    <path
                      d="M22,12 a6,6 0 1 1 6,-6 v6z M22,16 a6,6 0 0 1 0,12 h-16 a6,6 0 1 1 0,-12"
                      fill="#36C5F0"
                    />
                    <path
                      d="M48,22 a6,6 0 1 1 6,6 h-6z M32,6 a6,6 0 1 1 12,0v16a6,6 0 0 1 -12,0z"
                      fill="#2EB67D"
                    />
                    <path
                      d="M38,48 a6,6 0 1 1 -6,6 v-6z M54,32 a6,6 0 0 1 0,12 h-16 a6,6 0 1 1 0,-12"
                      fill="#ECB22E"
                    />
                    <path
                      d="M12,38 a6,6 0 1 1 -6,-6 h6z M16,38 a6,6 0 1 1 12,0v16a6,6 0 0 1 -12,0z"
                      fill="#E01E5A"
                    />
                  </svg>
                </div>
                <div className="ml-3 sm:ml-5">
                  <h3 className="text-base sm:text-lg font-medium text-white">
                    {t("card2Title")}
                  </h3>
                  <p className="text-gray-300 mt-1 text-sm sm:text-base">
                    {t("card2Description")}
                  </p>
                  <p className="text-purple-400 mt-1 text-xs sm:text-sm inline-flex items-center">
                    {t("card2Link")}
                    <svg
                      className="ml-1 w-3 h-3 sm:w-4 sm:h-4"
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
                  </p>
                </div>
              </div>
            </a>

            {/* Contact card 3 */}
            <a
              href="https://github.com/kubestellar/kubestellar"
              target="_blank"
              rel="noopener noreferrer"
              className="block bg-gray-800/50 backdrop-blur-md rounded-xl shadow-sm border border-gray-700/50 p-4 sm:p-6 transform transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-green-500/50 cursor-pointer"
            >
              <div className="flex items-center">
                <div className="flex-shrink-0 h-10 w-10 sm:h-12 sm:w-12 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex items-center justify-center shadow-lg border border-gray-600">
                  <svg
                    className="h-6 w-6 sm:h-8 sm:w-8 text-white"
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                  >
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                  </svg>
                </div>
                <div className="ml-3 sm:ml-5">
                  <h3 className="text-base sm:text-lg font-medium text-white">
                    {t("card3Title")}
                  </h3>
                  <p className="text-gray-300 mt-1 text-sm sm:text-base">
                    {t("card3Description")}
                  </p>
                  <p className="text-green-400 mt-1 text-xs sm:text-sm inline-flex items-center">
                    {t("card3Link")}
                    <svg
                      className="ml-1 w-3 h-3 sm:w-4 sm:h-4"
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
                  </p>
                </div>
              </div>
            </a>
            {/* Contact card 4 - LinkedIn */}
            <a
              href="https://linkedin.com/company/kubestellar"
              target="_blank"
              rel="noopener noreferrer"
              className="block bg-gray-800/50 backdrop-blur-md rounded-xl shadow-sm border border-gray-700/50 p-4 sm:p-6 transform transition-all duration-300 hover:shadow-md hover:-translate-y-1 hover:border-blue-500/50 cursor-pointer"
            >
              <div className="flex items-center">
                <div className="flex-shrink-0 h-8 w-8 sm:h-10 sm:w-10 rounded-full bg-blue-600 flex items-center justify-center shadow-lg">
                  <svg
                    className="h-4 w-4 sm:h-6 sm:w-6"
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 382 382"
                    fill="white"
                  >
                    <path d="M347.445,0H34.555C15.471,0,0,15.471,0,34.555v312.889C0,366.529,15.471,382,34.555,382h312.889 C366.529,382,382,366.529,382,347.444V34.555C382,15.471,366.529,0,347.445,0z M118.207,329.844c0,5.554-4.502,10.056-10.056,10.056 H65.345c-5.554,0-10.056-4.502-10.056-10.056V150.403c0-5.554,4.502-10.056,10.056-10.056h42.806 c5.554,0,10.056,4.502,10.056,10.056V329.844z M86.748,123.432c-22.459,0-40.666-18.207-40.666-40.666S64.289,42.1,86.748,42.1 s40.666,18.207,40.666,40.666S109.208,123.432,86.748,123.432z M341.91,330.654c0,5.106-4.14,9.246-9.246,9.246H286.73 c-5.106,0-9.246-4.14-9.246-9.246v-84.168c0-12.556,3.683-55.021-32.813-55.021c-28.309,0-34.051,29.066-35.204,42.11v97.079 c0,5.106-4.139,9.246-9.246,9.246h-44.426c-5.106,0-9.246-4.14-9.246-9.246V149.593c0-5.106,4.14-9.246,9.246-9.246h44.426 c5.106,0,9.246,4.14,9.246,9.246v15.655c10.497-15.753,26.097-27.912,59.312-27.912c73.552,0,73.131,68.716,73.131,106.472 L341.91,330.654L341.91,330.654z" />
                  </svg>
                </div>
                <div className="ml-3 sm:ml-5">
                  <h3 className="text-base sm:text-lg font-medium text-white">
                    {t("card4Title")}
                  </h3>
                  <p className="text-gray-300 mt-1 text-sm sm:text-base">
                    {t("card4Description")}
                  </p>
                  <p className="text-blue-400 mt-1 text-xs sm:text-sm inline-flex items-center">
                    {t("card4Link")}
                    <svg
                      className="ml-1 w-3 h-3 sm:w-4 sm:h-4"
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
                  </p>
                </div>
              </div>
            </a>
          </div>

          {/* Right side: Contact form */}
          <div className="lg:col-span-3 flex flex-col mt-6 lg:mt-0">
            <div className="bg-gradient-to-br from-gray-800/60 to-gray-900/60 backdrop-blur-md rounded-2xl shadow-2xl border border-gray-700/50 overflow-hidden h-full flex flex-col">
              <div className="p-4 sm:p-6 md:p-8 flex-1 flex flex-col">
                <h3 className="text-xl sm:text-2xl font-bold text-white mb-4 sm:mb-6 text-center">
                  {t("formTitle")}
                </h3>

                <form
                  name="contact"
                  method="POST"
                  data-netlify="true"
                  data-netlify-honeypot="bot-field"
                  onSubmit={handleSubmit}
                  className="space-y-3 sm:space-y-4 flex-1 flex flex-col"
                >
                  <input type="hidden" name="form-name" value="contact" />
                  <input type="hidden" name="bot-field" />
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
                    <div>
                      <label
                        htmlFor="name"
                        className="block text-sm font-semibold text-gray-300 mb-1 sm:mb-2"
                      >
                        {t("formName")}
                      </label>
                      <input
                        type="text"
                        id="name"
                        name="name"
                        value={formData.name}
                        onChange={handleInputChange}
                        required
                        className="w-full px-3 sm:px-4 py-2 sm:py-3 bg-gray-700/60 border border-gray-600/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-300 backdrop-blur-sm text-sm sm:text-base"
                        placeholder={t("formNamePlaceholder")}
                      />
                    </div>

                    <div>
                      <label
                        htmlFor="email"
                        className="block text-sm font-semibold text-gray-300 mb-1 sm:mb-2"
                      >
                        {t("formEmail")}
                      </label>
                      <input
                        type="email"
                        id="email"
                        name="email"
                        value={formData.email}
                        onChange={handleInputChange}
                        required
                        className="w-full px-3 sm:px-4 py-2 sm:py-3 bg-gray-700/60 border border-gray-600/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-300 backdrop-blur-sm text-sm sm:text-base"
                        placeholder={t("formEmailPlaceholder")}
                      />
                    </div>
                  </div>

                  <div>
                    <label
                      htmlFor="subject"
                      className="block text-sm font-semibold text-gray-300 mb-1 sm:mb-2"
                    >
                      {t("formSubject")}
                    </label>
                    <div className="relative">
                      <select
                        id="subject"
                        name="subject"
                        value={formData.subject}
                        onChange={handleInputChange}
                        required
                        className="w-full px-3 sm:px-4 py-2 sm:py-3 pr-10 sm:pr-12 bg-gray-700/60 border border-gray-600/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-300 backdrop-blur-sm appearance-none cursor-pointer hover:border-gray-500/60 hover:bg-gray-700/70 text-sm sm:text-base"
                      >
                        <option value="" disabled className="text-gray-400">
                          {t("formSubjectPlaceholder")}
                        </option>
                        <option
                          value="General Inquiry"
                          className="bg-gray-800 text-white py-2"
                        >
                          {t("formSubjectOption1")}
                        </option>
                        <option
                          value="Technical Support"
                          className="bg-gray-800 text-white py-2"
                        >
                          {t("formSubjectOption2")}
                        </option>
                        <option
                          value="Partnership"
                          className="bg-gray-800 text-white py-2"
                        >
                          {t("formSubjectOption3")}
                        </option>
                        <option
                          value="Documentation Feedback"
                          className="bg-gray-800 text-white py-2"
                        >
                          {t("formSubjectOption4")}
                        </option>
                        <option
                          value="Enterprise Solutions"
                          className="bg-gray-800 text-white py-2"
                        >
                          {t("formSubjectOption5")}
                        </option>
                        <option
                          value="Other"
                          className="bg-gray-800 text-white py-2"
                        >
                          {t("formSubjectOption6")}
                        </option>
                      </select>
                      {/* Custom dropdown chevron */}
                      <div className="absolute inset-y-0 right-0 flex items-center pr-3 sm:pr-4 pointer-events-none">
                        <svg
                          className="w-4 h-4 sm:w-5 sm:h-5 text-gray-400 transition-transform duration-200"
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
                      className="block text-sm font-semibold text-gray-300 mb-1 sm:mb-2"
                    >
                      {t("formMessage")}
                    </label>
                    <textarea
                      id="message"
                      name="message"
                      value={formData.message}
                      onChange={handleInputChange}
                      required
                      className="w-full px-3 sm:px-4 py-2 sm:py-3 bg-gray-700/60 border border-gray-600/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-300 backdrop-blur-sm resize-none flex-1 min-h-[100px] sm:min-h-[120px] text-sm sm:text-base"
                      placeholder={t("formMessagePlaceholder")}
                    ></textarea>
                  </div>

                  <div className="flex items-start space-x-2 sm:space-x-3">
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
                      className="text-xs sm:text-sm text-gray-300 leading-relaxed"
                    >
                      {t("formPrivacy")}{" "}
                      <Link
                        href="/docs/contribution-guidelines/license-inc"
                        className="text-blue-400 hover:text-blue-300 underline transition-colors duration-200"
                      >
                        {t("formPrivacyLink")}
                      </Link>{" "}
                      {t("formPrivacyCont")}
                    </label>
                  </div>

                  <div className="pt-2 sm:pt-3">
                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className="w-full py-2 sm:py-3 px-4 sm:px-6 bg-gradient-to-r from-blue-600 via-purple-600 to-blue-700 hover:from-blue-700 hover:via-purple-700 hover:to-blue-800 disabled:from-gray-600 disabled:to-gray-700 rounded-xl font-bold text-white shadow-lg hover:shadow-xl transition-all duration-300 transform hover:scale-[1.02] focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900 text-sm sm:text-base"
                    >
                      {isSubmitting ? (
                        <div className="flex items-center justify-center space-x-2">
                          <div className="animate-spin h-4 w-4 sm:h-5 sm:w-5 border-2 border-white border-t-transparent rounded-full"></div>
                          <span>{t("formSubmitting")}</span>
                        </div>
                      ) : (
                        <div className="flex items-center justify-center space-x-2">
                          <span>{t("formSubmit")}</span>
                        </div>
                      )}
                    </button>
                  </div>
                </form>

                {/* Success Message */}
                {showSuccess && (
                  <div className="mt-3 sm:mt-4 rounded-xl bg-green-900/30 p-3 sm:p-4 border border-green-500/30 backdrop-blur-sm">
                    <div className="flex">
                      <div className="flex-shrink-0">
                        <svg
                          className="h-4 w-4 sm:h-5 sm:w-5 text-green-400"
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
                      <div className="ml-2 sm:ml-3">
                        <p className="text-xs sm:text-sm font-medium text-green-300">
                          {t("formSuccess")}
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
