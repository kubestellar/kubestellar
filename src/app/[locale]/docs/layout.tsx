import { DocsNavbar, DocsFooter, DocsBanner } from '@/components/docs/index'
import { DocsProvider } from '@/components/docs/DocsProvider'
import { Suspense } from 'react'

export const metadata = {
  title: 'KubeStellar - Multi-Cluster Kubernetes Orchestration',
  description: 'Official documentation for KubeStellar - Multi-cluster orchestration platform',
}

type Props = {
  children: React.ReactNode
}

export default async function DocsLayout({ children }: Props) {
  return (
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
  )
}
