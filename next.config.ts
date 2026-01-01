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
  images: {
    unoptimized: true,
  },
  experimental: {
    optimizePackageImports: ["@/components"],
  },
  pageExtensions: ["js", "jsx", "ts", "tsx", "md", "mdx"],
  async rewrites() {
    return [
      {
        // Serve docs images from docs/content folder
        source: "/docs-images/:path*",
        destination: "/api/docs-image/:path*",
      },
    ];
  },
  async redirects() {
    return [
      {
        source: "/agenda",
        destination:
          "https://docs.google.com/document/d/1XppfxSOD7AOX1lVVVIPWjpFkrxakfBfVzcybRg17-PM/edit?usp=share_link",
        permanent: true,
      },
      {
        source: "/blog",
        destination:
          "https://medium.com/@kubestellar/list/predefined:e785a0675051:READING_LIST",
        permanent: true,
      },
      {
        source: "/code",
        destination: "https://github.com/kubestellar/kubestellar",
        permanent: true,
      },
      {
        source: "/community",
        destination: "https://docs.kubestellar.io/stable/Community/_index/",
        permanent: true,
      },
      {
        source: "/drive",
        destination:
          "https://drive.google.com/drive/u/1/folders/1p68MwkX0sYdTvtup0DcnAEsnXElobFLS",
        permanent: true,
      },
      {
        source: "/infomercial",
        destination: "https://youtu.be/rCjQAdwvZjk",
        permanent: true,
      },
      {
        source: "/join_us",
        destination: "https://groups.google.com/g/kubestellar-dev",
        permanent: true,
      },
      {
        source: "/joinus",
        destination: "https://groups.google.com/g/kubestellar-dev",
        permanent: true,
      },
      {
        source: "/ladder",
        destination: "https://kubestellar.io/en/ladder",
        permanent: true,
      },
      {
        source: "/ladder_stats",
        destination:
          "https://docs.google.com/spreadsheets/d/16CxUk2tNbTB-Si0qRVwIrI_f19t9HwSby9C1djMN7Sc/edit?usp=sharing",
        permanent: true,
      },
      {
        source: "/linkedin",
        destination:
          "https://www.linkedin.com/feed/hashtag/?keywords=kubestellar",
        permanent: true,
      },
      {
        source: "/quickstart",
        destination: "https://kubestellar.io/en/quick-installation",
        permanent: true,
      },
      {
        source: "/slack",
        destination: "https://cloud-native.slack.com/archives/C097094RZ3M",
        permanent: true,
      },
      {
        source: "/survey",
        destination: "https://forms.gle/WJ7N6ZVtp44D9NK79",
        permanent: true,
      },
      {
        source: "/tv",
        destination: "https://youtube.com/@kubestellar",
        permanent: true,
      },
      {
        source: "/youtube",
        destination: "https://youtube.com/@kubestellar",
        permanent: true,
      },
    ];
  },
};

const configWithNextra = withNextra(nextConfig);

// Note: Route-level exclusion is handled in src/middleware.ts (matcher excludes /docs)
const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");

export default withNextIntl(configWithNextra);
