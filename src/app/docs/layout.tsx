import { Layout, Navbar } from 'nextra-theme-docs'
import { Banner } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import 'nextra-theme-docs/style.css'
import Footer from '@/components/Footer'
 
export const metadata = {
  title: 'KubeStellar - Multi-Cluster Kubernetes Orchestration',
  description: 'Official documentation for KubeStellar - Multi-cluster orchestration platform',
}
 
const banner = <Banner storageKey="kubestellar-demo">Welcome to KubeStellar Docs - Powered by Nextra! 🎉</Banner>
const navbar = (
  <Navbar
    logo={<b>KubeStellar Docs</b>}
    projectLink="https://github.com/kubestellar/kubestellar"
  />
)
const footer = <Footer />
 
export default async function DocsLayout({ children }: { children: React.ReactNode }) {
  // Get the full pageMap and filter to only docs routes
  const fullPageMap = await getPageMap()
  
  // Create a filtered pageMap with only the docs folder content
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const docsPageMap = fullPageMap.filter((item: any) => {
    // Only include items that are within the docs route
    return item.route === '/docs' || item.route?.startsWith('/docs/')
  })
  
  return (
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
  )
}