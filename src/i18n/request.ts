import { getRequestConfig } from "next-intl/server";
import { locales, type Locale } from "./settings";

export default getRequestConfig(async ({ requestLocale }) => {
  let locale = await requestLocale;

  const isLocale = (val: string): val is Locale =>
    (locales as readonly string[]).includes(val);

  if (!locale || !isLocale(locale)) {
    locale = "en";
  }

  return {
    locale,
    messages: (await import(`../../messages/${locale}.json`)).default,
  };
});
