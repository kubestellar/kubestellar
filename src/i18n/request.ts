import { getRequestConfig } from "next-intl/server";
import { locales, type Locale, defaultLocale } from "./settings";
import { notFound } from "next/navigation";

const isLocale = (val: string): val is Locale =>
  (locales as readonly string[]).includes(val);

function deepMerge(
  target: Record<string, unknown>,
  source: Record<string, unknown>
): Record<string, unknown> {
  const output = { ...target };
  for (const key of Object.keys(source)) {
    if (
      typeof source[key] === "object" &&
      source[key] !== null &&
      !Array.isArray(source[key])
    ) {
      output[key] = deepMerge(
        (target[key] as Record<string, unknown>) || {},
        source[key] as Record<string, unknown>
      );
    } else {
      output[key] = source[key];
    }
  }
  return output;
}

export default getRequestConfig(async ({ requestLocale }) => {
  const locale = await requestLocale;

  if (!locale || !isLocale(locale)) {
    notFound();
  }

  const defaultMessages = (await import(`../../messages/${defaultLocale}.json`))
    .default;

  if (locale === defaultLocale) {
    return { locale, messages: defaultMessages };
  }

  let localeMessages;
  try {
    localeMessages = (await import(`../../messages/${locale}.json`)).default;
  } catch {
    console.warn(
      `Could not load messages for locale: ${locale}. Falling back to '${defaultLocale}'.`
    );
    return { locale, messages: defaultMessages };
  }

  const mergedMessages = deepMerge(defaultMessages, localeMessages);

  return {
    locale,
    messages: mergedMessages,
  };
});
