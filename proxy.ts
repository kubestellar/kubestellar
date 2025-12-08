import createMiddleware from "next-intl/middleware";
import { NextRequest } from "next/server";

export default async function proxy(request: NextRequest) {
  // Step 1: Use the incoming request (example)
  const defaultLocale = request.headers.get("x-your-custom-locale") || "en";

  // Step 2: Create and call the next-intl middleware (example)
  const handleI18nRouting = createMiddleware({
    locales: ["en", "de"],
    defaultLocale: "en",
  });
  const response = handleI18nRouting(request);

  // Step 3: Alter the response (example)
  response.headers.set("x-your-custom-locale", defaultLocale);

  return response;
}

export const config = {
  // Match only internationalized pathnames
  matcher: [
    "/((?!docs|api|_next|_vercel|agenda|blog|code|community|drive|infomercial|join_us|joinus|ladder|ladder_stats|linkedin|quickstart|slack|survey|tv|youtube|.*\\..*).*)",
  ],
};
