import { DocsNavbar, DocsFooter, DocsBanner } from '@/components/docs/index'
import { DocsProvider } from '@/components/docs/DocsProvider'
import { Inter, JetBrains_Mono } from "next/font/google"
import { Suspense } from 'react'
import { ThemeProvider } from "next-themes"
import "../globals.css"
import { buildPageMap } from './page-map'

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
})

const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
})

export const metadata = {
  title: 'KubeStellar - Multi-Cluster Kubernetes Orchestration',
  description: 'Official documentation for KubeStellar - Multi-cluster orchestration platform',
}

type Props = {
  children: React.ReactNode
}

export default async function DocsLayout({ children }: Props) {
  // Build page map from local docs
  buildPageMap();
  
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} ${jetbrainsMono.variable} antialiased`}>
        <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
          <DocsProvider>
            <div className="flex flex-col min-h-screen">
              <DocsBanner />
              <Suspense fallback={<div className="h-16" />}>
                <DocsNavbar />
              </Suspense>
              <main className="flex-1">
                {children}
              </main>
              <DocsFooter />
            </div>
          </DocsProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}