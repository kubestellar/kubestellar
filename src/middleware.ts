import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import createMiddleware from "next-intl/middleware";
import { locales, defaultLocale } from "./i18n/settings";

const intlMiddleware = createMiddleware({
  locales,
  defaultLocale,
  localePrefix: "always",
});

export default function middleware(request: NextRequest) {
  // Redirect docs.kubestellar.io before any other processing
  if (request.nextUrl.hostname === "docs.kubestellar.io") {
    return NextResponse.redirect("https://kubestellar.io/docs", 301);
  }

  // Run the i18n middleware for everything else
  return intlMiddleware(request);
}

export const config = {
  matcher: [
    "/((?!docs|api|_next|_vercel|agenda|blog|code|community|drive|infomercial|join_us|joinus|ladder|ladder_stats|linkedin|quickstart|slack|survey|tv|youtube|.*\\..*).*)",
    "/",
  ],
};
