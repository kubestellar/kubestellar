import { Layout } from 'nextra-theme-docs'
import { Banner } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import 'nextra-theme-docs/style.css'
import { DocsNavbar, DocsFooter } from '@/components/docs/index'
import { Inter, JetBrains_Mono } from "next/font/google";
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

export const metadata = {
  title: 'KubeStellar - Multi-Cluster Kubernetes Orchestration',
  description: 'Official documentation for KubeStellar - Multi-cluster orchestration platform',
}

const banner = <Banner storageKey="kubestellar-demo"><strong>Hacktoberfest 2025</strong> is here! Join us to learn, share, and contribute to our community🎉</Banner>
const navbar = <DocsNavbar />
const footer = <DocsFooter />
 
export default async function DocsLayout({ children }: { children: React.ReactNode }) {
  const fullPageMap = await getPageMap()
  
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const docsPageMap = fullPageMap.filter((item: any) => {
    return item.route === '/docs' || item.route?.startsWith('/docs/')
  })
  
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} ${jetbrainsMono.variable} antialiased`}>
        <Layout
          banner={banner}
          navbar={navbar}
          pageMap={docsPageMap}
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