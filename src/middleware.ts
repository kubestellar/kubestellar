import createMiddleware from "next-intl/middleware";
import { locales, defaultLocale } from "./i18n/settings";

export default createMiddleware({
  locales,
  defaultLocale,
  localePrefix: "always",
});

export const config = {
  matcher: ["/((?!docs|api|_next|_vercel|.*\\..*).*)", "/"],
};
