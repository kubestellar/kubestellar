import type { NextConfig } from "next";
import nextra from "nextra";

import createNextIntlPlugin from "next-intl/plugin";

const withNextra = nextra({
  latex: true,
  search: {
    codeblocks: false,
  },
});

const nextConfig: NextConfig = {
  output: "standalone",
  images: {
    unoptimized: true,
  },
  experimental: {
    optimizePackageImports: ["@/components"],
  },
  pageExtensions: ["js", "jsx", "ts", "tsx", "md", "mdx"],
};

const configWithNextra = withNextra(nextConfig);

// Note: Route-level exclusion is handled in src/middleware.ts (matcher excludes /docs)
const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");

export default withNextIntl(configWithNextra);
