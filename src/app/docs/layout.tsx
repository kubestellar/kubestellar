import { Layout } from 'nextra-theme-docs'
import 'nextra-theme-docs/style.css'
import { DocsNavbar, DocsFooter, DocsBanner } from '@/components/docs/index'
import { Inter, JetBrains_Mono } from "next/font/google"
import "../globals.css"
import { buildPageMapForBranch } from './page-map'
import { getDefaultVersion, getBranchForVersion } from '@/config/versions'

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
const navbar = <DocsNavbar />
const footer = <DocsFooter />

type Props = {
  children: React.ReactNode
}

export default async function DocsLayout({ children }: Props) {
  // Always use default version for initial layout
  // The page component will handle version-specific content
  const defaultVersion = getDefaultVersion()
  const branch = getBranchForVersion(defaultVersion)
  
  // Build page map for the default version
  const { pageMap } = await buildPageMapForBranch(branch)
  
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} ${jetbrainsMono.variable} antialiased`}>
        <Layout
          banner={banner}
          navbar={navbar}
          pageMap={pageMap}
          docsRepositoryBase="https://github.com/kubestellar/kubestellar"
          footer={footer}
          darkMode={true}
        >
          {children}
        </Layout>
      </body>
    </html>
  )
}