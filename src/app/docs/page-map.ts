import { normalizePageMap } from 'nextra/page-map'
import fs from 'fs'
import path from 'path'

// Local docs path - docs are now in this repository
export const docsContentPath = path.join(process.cwd(), 'docs', 'content')
export const basePath = 'docs'

// Strong types for page-map nodes
type MdxPageNode = { kind: 'MdxPage'; name: string; route: string }
type FolderNode = { kind: 'Folder'; name: string; route: string; children: PageMapNode[]; theme?: { collapsed?: boolean } }
type MetaNode = { kind: 'Meta'; data: Record<string, string> }
type PageMapNode = MdxPageNode | FolderNode | MetaNode

// Helper to prettify names
const pretty = (s: string) => s.charAt(0).toUpperCase() + s.slice(1).replace(/-/g, ' ')

// Recursively get all markdown files from the local docs directory
function getAllDocFiles(dir: string, baseDir: string = dir): string[] {
  const files: string[] = []
  
  if (!fs.existsSync(dir)) {
    return files
  }
  
  const entries = fs.readdirSync(dir, { withFileTypes: true })
  
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name)
    const relativePath = path.relative(baseDir, fullPath)
    
    if (entry.isDirectory()) {
      // Skip hidden directories and node_modules
      if (!entry.name.startsWith('.') && entry.name !== 'node_modules') {
        files.push(...getAllDocFiles(fullPath, baseDir))
      }
    } else if (entry.isFile() && (entry.name.endsWith('.md') || entry.name.endsWith('.mdx'))) {
      files.push(relativePath)
    }
  }
  
  return files
}

// Navigation structure based on mkdocs.yml
type NavItem = { [key: string]: string | NavItem[] | NavItem } | string

const NAV_STRUCTURE: Array<{ title: string; items: NavItem[] }> = [

  {
    title: 'What is KubeStellar?',
    items: [
      { 'Overview': 'readme.md' },
      { 'Architecture': 'direct/architecture.md' },
      {
        'Related': [
          { 'KubeStellar UI': 'direct/ui-intro.md' },
          { 'KubeFlex': 'direct/kubeflex-intro.md' },
          { 'KubeStellar Galaxy': 'direct/galaxy-intro.md' }
        ]
      },
      { 'Release Notes': 'direct/release-notes.md' },
      { 'Roadmap': 'direct/roadmap.md' }
    ]
  },
  {
    title: 'User Guide',
    items: [
      { 'Quick Start': 'direct/get-started.md' },
      { 'Guide Overview': 'direct/user-guide-intro.md' },
      { 'Observability': 'direct/observability.md' },
      { 'Getting Started': 'direct/get-started.md' },
      { 'Getting Started from OCM': 'direct/start-from-ocm.md' },
      {
        'General Setup': [
          { 'Overview': 'direct/setup-overview.md' },
          { 'Setup Limitations': 'direct/setup-limitations.md' },
          { 'Prerequisites': 'direct/pre-reqs.md' },
          {
            'KubeFlex Hosting Cluster': [
              { 'Acquire Cluster for KubeFlex Hosting': 'direct/acquire-hosting-cluster.md' },
              { 'Initialize KubeFlex Hosting Cluster': 'direct/init-hosting-cluster.md' }
            ]
          },
          {
            'Core Spaces': [
              { 'Inventory and Transport Spaces': 'direct/its.md' },
              { 'Workload Description Spaces': 'direct/wds.md' }
            ]
          },
          { 'Core Helm Chart': 'direct/core-chart.md' },
          { 'Argo CD Integration': 'direct/core-chart-argocd.md' },
          {
            'Workload Execution Clusters': [
              { 'About Workload Execution Clusters': 'direct/wec.md' },
              { 'Register a Workload Execution Cluster': 'direct/wec-registration.md' }
            ]
          }
        ]
      },
      {
        'Usage': [
          { 'Usage Limitations': 'direct/usage-limitations.md' },
          {
            'KubeStellar API': [
              { 'Overview': 'direct/control.md' },
              { 'Binding': 'direct/binding.md' },
              { 'Transforming Desired State': 'direct/transforming.md' },
              { 'Combining Reported State': 'direct/combined-status.md' },
              { 'Multi-WEC Aggregated Status': 'direct/multi-wec-aggregated-status.md' }
            ]
          },
          { 'Authorization in WECs': 'direct/authorization.md' },
          { 'Example Scenarios': 'direct/example-scenarios.md' },
          {
            'Third-party Integrations': [
              { 'ArgoCD to WDS': 'direct/argo-to-wds1.md' }
            ]
          },
          { 'Troubleshooting': 'direct/troubleshooting.md' },
          {
            'Known Issues': [
              { 'Overview': 'direct/known-issues.md' },
              { 'Hidden State in Kubeconfig': 'direct/knownissue-kflex-extension.md' },
              { 'Kind Needs OS Reconfig': 'direct/knownissue-kind-config.md' },
              { 'Helm Chart Auth Failure': 'direct/knownissue-helm-ghcr.md' },
              { 'Missing CombinedStatus Results': 'direct/knownissue-collector-miss.md' },
              { 'Kind Host Configuration': 'direct/installation-errors.md' },
              { 'Insufficient CPU': 'direct/knownissue-cpu-insufficient-for-its1.md' }
            ]
          }
        ]
      },
      { 'UI (User Interface)': 'ui-docs/ui-overview.md' },
      { 'Teardown': 'direct/teardown.md' }
    ]
  },
  {
    title: 'Contributing',
    items: [
      { 'Overview': 'direct/contribute.md' },
      { 'Code of Conduct': 'contribution-guidelines/coc-inc.md' },
      { 'Guidelines': 'contribution-guidelines/contributing-inc.md' },
      { 'Contributor Ladder': 'contribution-guidelines/contributor_ladder.md' },
      { 'License': 'contribution-guidelines/license-inc.md' },
      { 'Governance': 'contribution-guidelines/governance-inc.md' },
      { 'Onboarding': 'contribution-guidelines/onboarding-inc.md' },
      {
        'Website': [
          { 'Docs Management': 'contribution-guidelines/operations/document-management.md' },
          { 'Style Guide': 'contribution-guidelines/operations/docs-styleguide.md' },
          { 'Testing PRs': 'contribution-guidelines/operations/testing-doc-prs.md' }
        ]
      },
      {
        'Security': [
          { 'Policy': 'contribution-guidelines/security/security-inc.md' },
          { 'Contacts': 'contribution-guidelines/security/security_contacts-inc.md' }
        ]
      },
      { 'Testing': 'direct/testing.md' },
      { 'Packaging': 'direct/packaging.md' },
      { 'Release Process': 'direct/release.md' },
      { 'Release Testing': 'direct/release-testing.md' },
      { 'Sign-off': 'direct/pr-signoff.md' }
    ]
  },
  {
    title: 'Community',
    items: [
      { 'Get Involved': 'Community/_index.md' },
      {
        'Partners': [
          { 'ArgoCD': 'Community/partners/argocd.md' },
          { 'Turbonomic': 'Community/partners/turbonomic.md' },
          { 'MVI': 'Community/partners/mvi.md' },
          { 'FluxCD': 'Community/partners/fluxcd.md' },
          { 'OpenZiti': 'Community/partners/openziti.md' },
          { 'Kyverno': 'Community/partners/kyverno.md' }
        ]
      }
    ]
  }
]

export function buildPageMap() {
  const allDocFiles = getAllDocFiles(docsContentPath)
  const processedFiles = new Set<string>()
  const routeMap: Record<string, string> = {}
  const _pageMap: PageMapNode[] = []

  function buildNavNodes(items: NavItem[], parentSlug: string): PageMapNode[] {
    const nodes: PageMapNode[] = []
    const meta: Record<string, string> = {}

    for (const item of items) {
      if (typeof item === 'string') {
        // Simple file reference
        if (allDocFiles.includes(item)) {
          processedFiles.add(item)
          const baseName = item.replace(/\.(md|mdx)$/i, '').split('/').pop()!
          const route = `/${basePath}/${parentSlug}/${baseName}`
          routeMap[`${parentSlug}/${baseName}`] = item
          nodes.push({ kind: 'MdxPage', name: pretty(baseName), route })
          meta[pretty(baseName)] = pretty(baseName)
        }
      } else {
        // Object with title: path or title: children
        const title = Object.keys(item)[0]
        const value = (item as Record<string, string | NavItem[]>)[title]

        if (typeof value === 'string') {
          // It's a file path
          if (value.startsWith('http')) {
            // External link
            nodes.push({ kind: 'MdxPage', name: title, route: value })
            meta[title] = title
          } else if (allDocFiles.includes(value)) {
            processedFiles.add(value)
            // const baseName = value.replace(/\.(md|mdx)$/i, '').split('/').pop()!
            const slug = title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
            const route = `/${basePath}/${parentSlug ? parentSlug + '/' : ''}${slug}`
            routeMap[`${parentSlug ? parentSlug + '/' : ''}${slug}`] = value
            nodes.push({ kind: 'MdxPage', name: title, route })
            meta[title] = title
          }
        } else if (Array.isArray(value)) {
          // It's a folder with children
          const slug = title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
          const newParentSlug = parentSlug ? `${parentSlug}/${slug}` : slug
          const children = buildNavNodes(value, newParentSlug)
          if (children.length > 0) {
            nodes.push({
              kind: 'Folder',
              name: title,
              route: `/${basePath}/${newParentSlug}`,
              children
            })
            meta[title] = title
          }
        }
      }
    }

    if (Object.keys(meta).length > 0) {
      nodes.unshift({ kind: 'Meta', data: meta })
    }

    return nodes
  }

  // Build navigation from NAV_STRUCTURE
  for (const category of NAV_STRUCTURE) {
    const categorySlug = category.title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
    const children = buildNavNodes(category.items, categorySlug)

    if (children.length > 0) {
      const folderNode: FolderNode = {
        kind: 'Folder',
        name: category.title,
        route: `/${basePath}/${categorySlug}`,
        children
      }

      // Set theme for first category to be expanded
      if (category.title === 'Welcome' || category.title === 'What is KubeStellar?') {
        folderNode.theme = { collapsed: false }
      }

      _pageMap.push(folderNode)
    }
  }

  // Add top-level meta - only include our defined navigation structure
  const meta: Record<string, string> = {}
  for (const category of NAV_STRUCTURE) {
    meta[category.title] = category.title
  }
  _pageMap.unshift({ kind: 'Meta', data: meta })

  // Populate routeMap with all files for fallback resolution (needed for link rewriting)
  for (const fp of allDocFiles) {
    const noExt = fp.replace(/\.(md|mdx)$/i, '')
    if (!routeMap[noExt]) {
      routeMap[noExt] = fp
    }
  }

  const pageMap = normalizePageMap(_pageMap)

  return { pageMap, routeMap, filePaths: allDocFiles }
}

// For backwards compatibility, export a function that doesn't need branch parameter
export async function buildPageMapForBranch() {
  return buildPageMap()
}
