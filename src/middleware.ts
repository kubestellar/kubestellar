import createMiddleware from "next-intl/middleware";
import { locales, defaultLocale } from "./i18n/settings";

export default createMiddleware({
  locales,
  defaultLocale,
  localePrefix: "always",
});

export const config = {
  matcher: [
    "/((?!docs|api|_next|_vercel|agenda|blog|code|community|drive|infomercial|join_us|joinus|ladder|ladder_stats|linkedin|quickstart|slack|survey|tv|youtube|.*\\..*).*)",
    "/",
  ],
};
