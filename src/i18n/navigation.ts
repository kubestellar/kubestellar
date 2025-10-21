import { createNavigation } from "next-intl/navigation";
import { locales } from "./settings";

export const { Link, redirect, usePathname, useRouter } = createNavigation({
  locales,
  localePrefix: "always",
});
