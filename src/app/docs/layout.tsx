import { Layout } from 'nextra-theme-docs'
import 'nextra-theme-docs/style.css'
import { DocsNavbar, DocsFooter, DocsBanner } from '@/components/docs/index'
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

const banner = <DocsBanner />
const navbar = (
  <Suspense fallback={<div style={{ height: '4rem' }} />}>
    <DocsNavbar />
  </Suspense>
)
const footer = <DocsFooter />

type Props = {
  children: React.ReactNode
}

export default async function DocsLayout({ children }: Props) {
  // Build page map from local docs
  const { pageMap } = buildPageMap()
  
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} ${jetbrainsMono.variable} antialiased`}>
        <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
          <Layout
            banner={banner}
            navbar={navbar}
            pageMap={pageMap}
            docsRepositoryBase="https://github.com/kubestellar/docs/edit/main/docs/content"
            footer={footer}
            darkMode={true}
            sidebar={{
              defaultMenuCollapseLevel: 1,
              toggleButton: true
            }}
            toc={{
              float: true,
              title: "On This Page"
            }}
          >
            {children}
          </Layout>
        </ThemeProvider>
      </body>
    </html>
  )
}