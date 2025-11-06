import { getRequestConfig } from "next-intl/server";
import { locales, type Locale, defaultLocale } from "./settings";
import { notFound } from "next/navigation";

// A type guard to check if the locale is valid
const isLocale = (val: string): val is Locale =>
  (locales as readonly string[]).includes(val);

export default getRequestConfig(async ({ requestLocale }) => {
  // Await the requestLocale promise
  const locale = await requestLocale;

  // Validate that the incoming `requestLocale` is a valid locale
  if (!locale || !isLocale(locale)) {
    notFound();
  }

  // Load the default locale's messages (English) as a base
  const defaultMessages = (await import(`../../messages/${defaultLocale}.json`))
    .default;

  // If the requested locale is the default, just return those messages
  if (locale === defaultLocale) {
    return {
      locale: locale,
      messages: defaultMessages,
    };
  }

  // If the requested locale is different, load its messages
  let localeMessages;
  try {
    localeMessages = (await import(`../../messages/${locale}.json`)).default;
  } catch {
    // This handles cases where a locale is valid but its JSON file is missing.
    // We'll log a warning and fall back to only English.
    console.warn(
      `Could not load messages for locale: ${locale}. Falling back to '${defaultLocale}'.`
    );
    return {
      locale: locale,
      messages: defaultMessages,
    };
  }

  // Return the merged messages.
  // Keys from `localeMessages` will override keys from `defaultMessages`.
  return {
    locale: locale,
    messages: {
      ...defaultMessages,
      ...localeMessages,
    },
  };
});
