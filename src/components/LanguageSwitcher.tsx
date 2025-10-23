"use client";

import React, { useState, useRef, useEffect } from "react";
import { useLocale, useTranslations } from "next-intl";
import { usePathname, useRouter } from "@/i18n/navigation";
import { locales, localeNames, type Locale } from "@/i18n/settings";

interface LanguageSwitcherProps {
  className?: string;
  showLabel?: boolean;
  variant?: "dropdown" | "minimal";
}

export default function LanguageSwitcher({
  className = "",
  showLabel = true,
  variant = "dropdown",
}: LanguageSwitcherProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [isTransitioning, setIsTransitioning] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const locale = useLocale() as Locale;
  const pathname = usePathname();
  const router = useRouter();
  const t = useTranslations("navigation");

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  // Close dropdown on escape key
  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setIsOpen(false);
        buttonRef.current?.focus();
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleEscape);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
    };
  }, [isOpen]);

  const handleLanguageChange = async (newLocale: Locale) => {
    if (newLocale === locale) {
      setIsOpen(false);
      return;
    }

    setIsTransitioning(true);

    try {
      // Add a small delay for smooth transition
      await new Promise(resolve => setTimeout(resolve, 150));

      // Navigate to the new locale
      router.replace(pathname, { locale: newLocale });

      // Close dropdown after successful navigation
      setIsOpen(false);
    } catch (error) {
      console.error("Failed to change language:", error);
    } finally {
      setIsTransitioning(false);
    }
  };

  const currentLanguage = localeNames[locale];

  if (variant === "minimal") {
    return (
      <div className={`relative ${className}`}>
        <button
          ref={buttonRef}
          onClick={() => setIsOpen(!isOpen)}
          disabled={isTransitioning}
          className="text-sm font-medium text-gray-300 hover:text-pink-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-pink-500/10 hover:shadow-lg hover:shadow-pink-500/20 disabled:opacity-50"
          aria-label={`Current language: ${currentLanguage}. Click to change language`}
          aria-expanded={isOpen}
          aria-haspopup="listbox"
        >
          <span className="text-xs">{locale.toUpperCase()}</span>
          <svg
            className={`w-3 h-3 transition-transform duration-200 ${isOpen ? "rotate-180" : ""}`}
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

        {isOpen && (
          <div
            ref={dropdownRef}
            className="absolute right-0 mt-2 w-48 bg-gray-800/95 backdrop-blur-sm rounded-lg shadow-xl border border-gray-700/50 py-2 z-50 animate-in fade-in-0 zoom-in-95 duration-200"
            role="listbox"
            aria-label="Select language"
          >
            {locales.map(loc => (
              <button
                key={loc}
                onClick={() => handleLanguageChange(loc)}
                disabled={isTransitioning}
                className={`w-full text-left px-4 py-2 text-sm transition-all duration-200 hover:bg-pink-500/20 hover:text-pink-300 disabled:opacity-50 flex items-center justify-between ${
                  loc === locale
                    ? "bg-pink-500/10 text-pink-300 font-medium"
                    : "text-gray-300"
                }`}
                role="option"
                aria-selected={loc === locale}
              >
                <span>{localeNames[loc]}</span>
                {loc === locale && (
                  <svg
                    className="w-4 h-4"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                      clipRule="evenodd"
                    />
                  </svg>
                )}
              </button>
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className={`relative ${className}`}>
      <button
        ref={buttonRef}
        onClick={() => setIsOpen(!isOpen)}
        disabled={isTransitioning}
        className="text-sm font-medium text-gray-300 hover:text-pink-400 transition-all duration-300 flex items-center space-x-1 px-3 py-2 rounded-lg hover:bg-pink-500/10 hover:shadow-lg hover:shadow-pink-500/20 hover:scale-100 transform nav-link-hover disabled:opacity-50"
        aria-label={`Current language: ${currentLanguage}. Click to change language`}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
      >
        <svg
          className="w-4 h-4 mr-2"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129"
          />
        </svg>
        {showLabel && <span>{currentLanguage}</span>}
        <svg
          className={`w-4 h-4 ml-1 transition-transform duration-200 ${isOpen ? "rotate-180" : ""}`}
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

      {isOpen && (
        <div
          ref={dropdownRef}
          className="absolute right-0 mt-2 w-56 bg-gray-800/95 backdrop-blur-md rounded-xl shadow-2xl py-2 ring-1 ring-gray-700/50 transition-all duration-200 z-50 animate-in fade-in-0 zoom-in-95"
          role="listbox"
          aria-label="Select language"
        >
          <div className="px-3 py-2 text-xs font-semibold text-gray-400 uppercase tracking-wider border-b border-gray-700/50 mb-1">
            {t("selectLanguage") || "Select Language"}
          </div>

          {locales.map(loc => (
            <button
              key={loc}
              onClick={() => handleLanguageChange(loc)}
              disabled={isTransitioning}
              className={`w-full text-left px-4 py-3 text-sm transition-all duration-200 hover:bg-pink-500/20 hover:text-pink-300 disabled:opacity-50 flex items-center justify-between group ${
                loc === locale
                  ? "bg-pink-500/10 text-pink-300 font-medium"
                  : "text-gray-300"
              }`}
              role="option"
              aria-selected={loc === locale}
            >
              <div className="flex items-center space-x-3">
                <span className="text-xs font-mono opacity-60 group-hover:opacity-100 transition-opacity">
                  {loc.toUpperCase()}
                </span>
                <span>{localeNames[loc]}</span>
              </div>

              {loc === locale && (
                <svg
                  className="w-4 h-4 text-pink-400"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                  aria-hidden="true"
                >
                  <path
                    fillRule="evenodd"
                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                    clipRule="evenodd"
                  />
                </svg>
              )}
            </button>
          ))}

          {isTransitioning && (
            <div className="absolute inset-0 bg-gray-800/80 backdrop-blur-sm rounded-xl flex items-center justify-center">
              <div className="flex items-center space-x-2 text-pink-400">
                <svg
                  className="w-4 h-4 animate-spin"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                <span className="text-sm">Switching...</span>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// Export different variants for convenience
export const LanguageSwitcherMinimal = (
  props: Omit<LanguageSwitcherProps, "variant">
) => <LanguageSwitcher {...props} variant="minimal" />;

export const LanguageSwitcherFull = (
  props: Omit<LanguageSwitcherProps, "variant">
) => <LanguageSwitcher {...props} variant="dropdown" />;
