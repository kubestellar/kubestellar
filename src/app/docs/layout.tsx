import { Footer, Layout, Navbar } from 'nextra-theme-docs'
import { Banner } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import 'nextra-theme-docs/style.css'
 
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
const footer = <Footer>MIT {new Date().getFullYear()} © KubeStellar.</Footer>
 
export default async function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <Layout
      banner={banner}
      navbar={navbar}
      pageMap={await getPageMap()}
      docsRepositoryBase="https://github.com/kubestellar/kubestellar"
      footer={footer}
    >
      {children}
    </Layout>
  )
}