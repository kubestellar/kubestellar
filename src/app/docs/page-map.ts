import { convertToPageMap, normalizePageMap } from 'nextra/page-map'

export const user = 'kubestellar'
export const repo = 'kubestellar'
export const docsPath = 'docs/content/'
export const basePath = 'docs'

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

export async function buildPageMapForBranch(branch: string) {
  async function fetchDocsTree(): Promise<GitTreeResp> {
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

  // Filter out Direct folder completely
  const ROOT_FOLDERS = Array.from(new Set(allDocFiles.map(fp => fp.split('/')[0])))
  const DIRECT_ROOT = ROOT_FOLDERS.find(r => r.toLowerCase() === 'direct')
  const UI_DOCS_ROOT = ROOT_FOLDERS.find(r => r.toLowerCase() === 'ui docs' || r.toLowerCase() === 'ui-docs')

  // Strong types for page-map nodes (no `any`)
  type MdxPageNode = { kind: 'MdxPage'; name: string; route: string }
  type FolderNode = { kind: 'Folder'; name: string; route: string; children: PageMapNode[] }
  type MetaNode = { kind: 'Meta'; data: Record<string, string> }
  type PageMapNode = MdxPageNode | FolderNode | MetaNode

  const _pageMap: PageMapNode[] = []
  const aliases: Array<{ alias: string; fp: string }> = []
  const processedFiles = new Set<string>()
  const pretty = (s: string) => s.charAt(0).toUpperCase() + s.slice(1).replace(/-/g, ' ')

  type NavItem = { [key: string]: string | NavItem[] } | { file: string, root?: string };

  const CATEGORY_MAPPINGS: Array<[string, NavItem[]]> = [
    ['What is Kubestellar?', [
      { 'Overview': 'README.md' },
      { file: 'architecture.md' },
      { file: 'related-projects.md' },
      { file: 'roadmap.md' },
      { file: 'release-notes.md' }
    ]],
    ['Install & Configure', [
      // { file: 'get-started.md' },  
      { file: 'pre-reqs.md' },
      { file: 'start-from-ocm.md' },
      { file: 'setup-limitations.md' },
      {
        'KubeFlex Hosting cluster': [
          { 'Acquire cluster for KubeFlex Hosting': 'direct/acquire-hosting-cluster.md' },
          { 'Initialize KubeFlex Hosting cluster': 'direct/init-hosting-cluster.md' }
        ]
      },
      {
        'Core Spaces': [
          { 'Inventory and Transport Spaces': 'direct/its.md' },
          { 'Workload Description Spaces': 'direct/wds.md' }
        ]
      },
      {
        'Workload Execution Clusters': [
          { 'About Workload Execution Clusters': 'direct/wec.md' },
          { 'Register a Workload Execution Cluster': 'direct/wec-registration.md' }
        ]
      },
      { file: 'core-chart.md' },
      { file: 'teardown.md' }
    ]],
    ['Use & Integrate', [
      { file: 'usage-limitations.md' },
      {
        'KubeStellar API': [
          { 'Overview': 'direct/control.md' },
          { 'API reference (new tab)': 'https://pkg.go.dev/github.com/kubestellar/kubestellar/api/control/v1alpha1' },
          { 'Binding': 'direct/binding.md' },
          { 'Transforming desired state': 'direct/transforming.md' },
          { 'Combining reported state': 'direct/combined-status.md' }
        ]
      },
      { file: 'example-scenarios.md' },
      {
        'Third-party integrations': [
          { 'ArgoCD to WDS': 'direct/argo-to-wds1.md' }
        ]
      }
    ]],
    ['User Guide & Support', [
      { file: 'user-guide-intro.md' },
      { file: 'troubleshooting.md' },
      {
        'Known Issues': [
          { 'Overview': 'direct/known-issues.md' },
          { 'Hidden state in kubeconfig': 'direct/knownissue-kflex-extension.md' },
          { 'Kind needs OS reconfig': 'direct/knownissue-kind-config.md' },
          { 'Authorization failure while fetching Helm chart from ghcr.io': 'direct/knownissue-helm-ghcr.md' },
          { 'Missing results in a CombinedStatus object': 'direct/knownissue-collector-miss.md' },
          { 'Kind host not configured for more than two clusters': 'direct/installation-errors.md' },
          { 'Insufficient CPU for your clusters': 'direct/knownissue-cpu-insufficient-for-its1.md' }
        ]
      },
      { file: 'combined-status.md' }
    ]],
    ['UI & Tools', [
      // from Direct
      { file: 'ui-intro.md' },
      { file: 'plugins.md' },
      { file: 'galaxy-marketplace.md' },
      { file: 'kubeflex-intro.md' },
      { file: 'galaxy-intro.md' },
      // from UI Docs folder
      { root: UI_DOCS_ROOT, file: 'README.md' },
      { root: UI_DOCS_ROOT, file: 'ui-overview.md' },
    ]]
  ]

  function buildNavNodes(items: NavItem[], parentSlug: string): PageMapNode[] {
    const nodes: PageMapNode[] = [];
    const meta: Record<string, string> = {};

    for (const item of items) {
        let node: PageMapNode | null = null;
        let keyForMeta: string | null = null;

        if ('file' in item) {
            const root = item.root || DIRECT_ROOT;
            if (!root) continue;
            const fullPath = `${root}/${item.file}`;
            if (allDocFiles.includes(fullPath)) {
                processedFiles.add(fullPath);
                const baseName = fullPath.replace(/\.(md|mdx)$/i, '').split('/').pop()!;
                const route = `/${basePath}/${parentSlug}/${baseName}`;
                const alias = `${parentSlug}/${baseName}`;
                aliases.push({ alias, fp: fullPath });
                node = { kind: 'MdxPage' as const, name: pretty(baseName), route };
                keyForMeta = pretty(baseName);
            }
        } else { 
            const title = Object.keys(item)[0];
            const value = item[title];

            if (typeof value === 'string') { 
                 if (value.startsWith('http')) {
                    node = { kind: 'MdxPage' as const, name: title, route: value };
                    keyForMeta = title;
                 } else {
                    // Case-insensitive file search
                    const foundFile = allDocFiles.find(f => f.toLowerCase() === value.toLowerCase());
                    if (foundFile) {
                        processedFiles.add(foundFile);
                        const baseName = foundFile.replace(/\.(md|mdx)$/i, '').split('/').pop()!;
                        // Correct route for root files like README.md
                        const isRootFile = !foundFile.includes('/');
                        const route = isRootFile ? `/${basePath}` : `/${basePath}/${parentSlug}/${baseName}`;
                        const alias = isRootFile ? '' : route.replace(`/${basePath}/`, '');
                        aliases.push({ alias, fp: foundFile });
                        node = { kind: 'MdxPage' as const, name: title, route };
                        keyForMeta = title;
                    }
                }
            } else if (Array.isArray(value)) { // It's a sub-category
                const slug = title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
                const children = buildNavNodes(value, `${parentSlug}/${slug}`);
                if (children.length > 0) {
                    node = { kind: 'Folder', name: title, route: `/${basePath}/${parentSlug}/${slug}`, children };
                    keyForMeta = title;
                }
            }
        }

        if (node && keyForMeta) {
            nodes.push(node);
            meta[keyForMeta] = keyForMeta;
        }
    }

    if (Object.keys(meta).length > 0) {
      nodes.unshift({ kind: 'Meta', data: meta });
    }
    return nodes;
  }


  for (const [categoryName, fileConfigs] of CATEGORY_MAPPINGS) {
    const categorySlug = categoryName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
    const children = buildNavNodes(fileConfigs, categorySlug);

    if (!children.length) continue

    _pageMap.push({ kind: 'Folder', name: categoryName, route: `/${basePath}/${categorySlug}`, children })
  }

  const remainingFiles = allDocFiles.filter(fp => {
    if (processedFiles.has(fp)) return false
    const lower = fp.toLowerCase()
    if (
      (DIRECT_ROOT && lower.startsWith(`${DIRECT_ROOT.toLowerCase()}/`)) ||
      (UI_DOCS_ROOT && lower.startsWith(`${UI_DOCS_ROOT.toLowerCase()}/`))
    ) {
      return false
    }
    return true
  })

  const { pageMap: remainingFileNodesRaw } = convertToPageMap({ filePaths: remainingFiles })
  const remainingFileNodes = remainingFileNodesRaw as unknown as PageMapNode[]

  // Type guards so TS knows which fields exist
  const hasRoute = (n: PageMapNode): n is MdxPageNode | FolderNode => 'route' in n
  const hasChildren = (n: PageMapNode): n is FolderNode => 'children' in n
  const hasName = (n: PageMapNode): n is MdxPageNode | FolderNode => 'name' in n

  function addBasePathToRoutes(nodes: PageMapNode[]) {
    for (const node of nodes) {
      if (hasRoute(node)) node.route = `/${basePath}${node.route}`
      if (hasChildren(node)) addBasePathToRoutes(node.children)
    }
  }

  addBasePathToRoutes(remainingFileNodes)

  _pageMap.push(...remainingFileNodes)

  const meta: Record<string, string> = {}
  for (const [categoryName] of CATEGORY_MAPPINGS) {
    meta[categoryName] = categoryName
  }
  for (const item of remainingFileNodes) {
    if (hasName(item)) {
      meta[item.name] = item.name
    }
  }

  _pageMap.unshift({
    kind: 'Meta',
    data: meta
  })

  const routeMap: Record<string, string> = {}
  // Populate routeMap from all files first
  for (const fp of allDocFiles) {
    const noExt = fp.replace(/\.(md|mdx)$/i, '')
    routeMap[noExt] = fp
  }
  // Overwrite with specific aliases from our custom structure to ensure correctness
  for (const { alias, fp } of aliases) {
    routeMap[alias] = fp
  }

  // normalizePageMap has compatible types now; remove stale suppressor
  const pageMap = normalizePageMap(_pageMap)

  return { pageMap, routeMap, filePaths: allDocFiles, branch }
}