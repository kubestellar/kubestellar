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

  // Explicitly handle root path to ensure consistent redirect to /en
  // This helps override any cached redirects in browsers like Safari
  if (request.nextUrl.pathname === "/") {
    const url = request.nextUrl.clone();
    url.pathname = `/${defaultLocale}`;
    return NextResponse.redirect(url, 307); // Use 307 to avoid aggressive caching
  }

  // Run the i18n middleware for everything else
  return intlMiddleware(request);
}

export const config = {
  matcher: [
    // Use negative lookahead to exclude docs EXCEPT kubectl-claude which gets i18n
    "/((?!docs(?!/kubectl-claude)|api|_next|_vercel|agenda|blog|code|community|drive|infomercial|join_us|joinus|ladder|ladder_stats|linkedin|quickstart|slack|survey|tv|youtube|.*\\..*).*)",
    "/",
  ],
};
