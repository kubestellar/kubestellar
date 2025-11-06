import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";

import { NextIntlClientProvider } from "next-intl";
import { getMessages } from "next-intl/server";
import type { ReactNode } from "react";
import { notFound } from "next/navigation";
import { locales, type Locale } from "@/i18n/settings";
import { PageTransitionLoader } from "@/components/animations/loader";
import "../globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
});

const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "KubeStellar - Multi-Cluster Kubernetes Orchestration",
  description:
    "Simplify multi-cluster Kubernetes operations with intelligent workload distribution, unified management, and seamless orchestration across any infrastructure.",
};

type Props = {
  children: ReactNode;
  params: Promise<{ locale: string }>;
};

export default async function RootLayout({ children, params }: Props) {
  const { locale } = await params;

  const isLocale = (val: string): val is Locale =>
    (locales as readonly string[]).includes(val);

  if (!isLocale(locale)) {
    notFound();
  }

  const messages = await getMessages();

  return (
    <html lang={locale} suppressHydrationWarning>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `
              try {
                const theme = localStorage.getItem('theme') || 'dark';
                document.documentElement.classList.toggle('dark', theme === 'dark');
              } catch (e) {}
            `,
          }}
        />
      </head>
      <body
        className={`${inter.variable} ${jetbrainsMono.variable} antialiased`}
      >
        <NextIntlClientProvider messages={messages}>
          <PageTransitionLoader>{children}</PageTransitionLoader>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
