export const locales = ["en", "hi", "ja", "es", "de", "fr", "it", "SC", "TC", "pt"] as const;
export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = "en";

export const localeNames: Record<Locale, string> = {
  en: "English",
  hi: "हिन्दी",
  ja: "日本語",
  es: "Español",
  de: "Deutsch",
  fr: "Français",
  it: "Italiano",
  SC: "简体中文",
  TC: "繁體中文",
  pt: "Português",
};
