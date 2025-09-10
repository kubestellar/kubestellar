import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";

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

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <head>
        <script src="https://cdn.tailwindcss.com"></script>
        <script src="https://unpkg.com/framer-motion@10.16.4/dist/framer-motion.js"></script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/gsap/3.11.4/gsap.min.js"></script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/three.js/r128/three.min.js"></script>
      </head>
      <body
        className={`${inter.variable} ${jetbrainsMono.variable} bg-space-dark text-white overflow-x-hidden dark antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
