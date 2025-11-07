import { convertToPageMap, normalizePageMap } from 'nextra/page-map'

export const user = 'kubestellar'
export const repo = 'kubestellar'
export const branch = 'main'
export const docsPath = 'docs/content/'
export const basePath = 'docs'
const INCLUDE_PREFIXES: string[] = []

export function makeGitHubHeaders(): Record<string, string> {
  const token = process.env.GITHUB_TOKEN || process.env.GH_TOKEN || process.env.GITHUB_PAT
  const h: Record<string, string> = {
    'User-Agent': 'kubestellar-docs',
    'Accept': 'application/vnd.github+json',
    'X-GitHub-Api-Version': '2022-11-28'
  }
  if (token) h.Authorization = `Bearer ${token}`
  return h
}

type GitTreeItem = { path: string; type: 'blob' | 'tree' }
type GitTreeResp = { tree?: GitTreeItem[] }

async function fetchDocsTree(): Promise<GitTreeResp> {
  // Resolve branch to commit SHA then fetch tree
  const refUrl = `https://api.github.com/repos/${user}/${repo}/git/refs/heads/${encodeURIComponent(branch)}`
  let sha: string | undefined
  const refRes = await fetch(refUrl, { headers: makeGitHubHeaders(), cache: 'no-store' })
  if (refRes.ok) {
    const ref = await refRes.json()
    sha = ref?.object?.sha
  }
  const treeUrl = sha
    ? `https://api.github.com/repos/${user}/${repo}/git/trees/${sha}?recursive=1`
    : `https://api.github.com/repos/${user}/${repo}/git/trees/${encodeURIComponent(branch)}?recursive=1`
  const treeRes = await fetch(treeUrl, { headers: makeGitHubHeaders(), cache: 'no-store' })
  if (!treeRes.ok) {
    const body = await treeRes.text().catch(() => '')
    throw new Error(`GitHub tree fetch failed: ${treeRes.status} ${treeRes.statusText} ${body}`)
  }
  return treeRes.json()
}

const treeData = await fetchDocsTree()

const allDocFiles =
  treeData.tree?.filter(
    t =>
      t.type === 'blob' &&
      t.path.startsWith(docsPath) &&
      (t.path.endsWith('.md') || t.path.endsWith('.mdx'))
  ).map(t => t.path.slice(docsPath.length)) ?? []

export const filePaths = INCLUDE_PREFIXES.length
  ? allDocFiles.filter(fp =>
      INCLUDE_PREFIXES.some(prefix => fp === prefix || fp.startsWith(prefix + '/'))
    )
  : allDocFiles

// Detect actual "Direct" folder name (case-insensitive)
const ROOT_FOLDERS = Array.from(new Set(filePaths.map(fp => fp.split('/')[0])))
const DIRECT_ROOT = ROOT_FOLDERS.find(r => r.toLowerCase() === 'direct')

// Normalize: drop extension + trailing index/readme
function noExt(fp: string) {
  return fp
    .replace(/\.(md|mdx)$/i, '')
    .replace(/\/(readme|index)$/i, '')
    .replace(/^(readme|index)$/i, '')
}

// HIDE the Direct folder AND every single page under it
const meta: Record<string, any> = {}
if (DIRECT_ROOT) {
  meta[DIRECT_ROOT] = false
  // Also hide every Direct/* page so they don't auto-promote
  for (const fp of filePaths) {
    if (fp.toLowerCase().startsWith(`${DIRECT_ROOT.toLowerCase()}/`)) {
      meta[noExt(fp)] = false
    }
  }
}

const filteredFilePaths = filePaths.filter(
  fp => !(DIRECT_ROOT && fp.toLowerCase().startsWith(`${DIRECT_ROOT.toLowerCase()}/`))
)

const { pageMap: baseMap } = convertToPageMap({
  filePaths: filteredFilePaths,
  basePath,
  meta
})

type PageMapNode = { kind: 'Folder'; name: string; route: string; children?: any[] } | any

const _pageMap: PageMapNode[] = (baseMap as any).filter(
  (n: any) =>
    !(
      n?.kind === 'Folder' &&
      DIRECT_ROOT &&
      typeof n.name === 'string' &&
      n.name.toLowerCase() === DIRECT_ROOT.toLowerCase()
    )
)


const aliases: Array<{ alias: string; fp: string }> = []

const CATEGORY_MAPPINGS: Record<string, string[]> = {
  'What is Kubestellar?': [
    'overview.md',
    'architecture.md',
    'related-projects.md',
    'roadmap.md',
    'release-notes.md'
  ],
  'Install & Configure': [
    '.get-started.md',
    'start-from-ocm.md',
    'setup-limitations.md',
    'acquire-hosting-cluster.md',
    'init-hosting-cluster.md',
    'core-specs/inventory-and-transport.md',
    'core-specs/workload-description.md',
    'workload-execution-cluster/about.md',
    'workload-execution-cluster/register.md',
    'core-chart.md',
    'teardown.md'
  ],
  'UI & Tools': [
    'ui-intro.md',
    'plugins.md',
    'galaxy-marketplace.md',
    'kubeflex-intro.md',
    'galaxy-intro.md'
  ],
  'Use & Integrate': [
    'usage-limitations.md',
    'binding.md',
    'transforming.md',
    'combined-status.md',
    'example-scenarios.md',
    'argo-to-wds1.md'
  ],
  'User Guide & Support': [
    'user-guide-intro.md',
    'troubleshooting.md',
    'known-issues.md',
    'knownissue-collector-miss.md',
    'knownissue-helm-ghcr.md',
    'knownissue-kind-config.md',
    'knownissue-cpu-insufficient-for-its1.md',
    'knownissue-kflex-extension.md',
    'combined-status.md'
  ]
}

const pretty = (s: string) => s.charAt(0).toUpperCase() + s.slice(1).replace(/-/g, ' ')

for (const [categoryName, relFiles] of Object.entries(CATEGORY_MAPPINGS)) {
  if (!DIRECT_ROOT) continue
  const fulls = relFiles.map(rel => `${DIRECT_ROOT}/${rel}`).filter(full => filePaths.includes(full))
  if (!fulls.length) continue

  const categorySlug = categoryName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
  const children = fulls.map(full => {
    const base = full.replace(/\.(md|mdx)$/i, '').split('/').pop()!
    const route = `/${basePath}/${categorySlug}/${base}`
    const alias = `${categorySlug}/${base}`
    aliases.push({ alias, fp: full })
    return { kind: 'MdxPage' as const, name: pretty(base), route }
  })

  _pageMap.push({ kind: 'Folder', name: categoryName, route: `/${basePath}/${categorySlug}`, children })
}

// Build routeMap to resolve requests -> actual files
function normalizeRoute(noExtPath: string) {
  let r = noExtPath
  r = r.replace(/\/(readme|index)$/i, '')
  r = r.replace(/^(readme|index)$/i, '')
  return r
}

export const routeMap: Record<string, string> = {}
for (const fp of filePaths) {
  const noExt = fp.replace(/\.(md|mdx)$/i, '')
  const norm = normalizeRoute(noExt)
  routeMap[noExt] = fp
  if (!noExt.startsWith('content/')) routeMap[`content/${noExt}`] = fp
  const isIndex = /\/(readme|index)$/i.test(noExt) || /^(readme|index)$/i.test(noExt)
  if (!routeMap[norm] || isIndex) routeMap[norm] = fp
  if (norm !== '' && !norm.startsWith('content/')) {
    const contentNorm = `content/${norm}`
    if (!routeMap[contentNorm] || isIndex) routeMap[contentNorm] = fp
  }
}
for (const { alias, fp } of aliases) {
  routeMap[alias] = fp
  if (!alias.startsWith('content/')) routeMap[`content/${alias}`] = fp
}

// @ts-expect-error nextra typing mismatch
export const pageMap = normalizePageMap(_pageMap)