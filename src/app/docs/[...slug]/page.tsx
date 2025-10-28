import { notFound } from 'next/navigation'
import { compileMdx } from 'nextra/compile'
import { Callout, Tabs } from 'nextra/components'
import { evaluate } from 'nextra/evaluate'
import {
  convertToPageMap,
  mergeMetaWithPageMap,
  normalizePageMap
} from 'nextra/page-map'
import { useMDXComponents as getMDXComponents } from '../../../../mdx-components'

const user = 'kubestellar'
const repo = 'kubestellar'
const branch = 'main'
const docsPath = 'docs/content/'
// const filePaths = [
//   'configs.mdx',
//   'custom-rules.mdx',
//   'getting-started.mdx',
//   'getting-started/parser-options.mdx',
//   'getting-started/parser.mdx',
//   'index.mdx'
// ]

const INCLUDE_PREFIXES: string[] = [

]
const basePath = 'docs'
const headers: Record<string, string> = {
  'User-Agent': 'kubestellar-docs-dev',
  'Accept': 'application/vnd.github+json'
}
if (process.env.GITHUB_TOKEN) {
  headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`
}

type GitTreeItem = { path: string; type: 'blob' | 'tree' }
type GitTreeResp = { tree?: GitTreeItem[] }

const treeUrl = `https://api.github.com/repos/${user}/${repo}/git/trees/${encodeURIComponent(
  branch
)}?recursive=1`

const treeData: GitTreeResp = await fetch(treeUrl, { headers, cache: 'force-cache' }).then(r =>
  r.ok ? r.json() : Promise.reject(new Error(`GitHub tree fetch failed: ${r.status}`))
)


const allDocFiles = treeData.tree?.filter(t => t.type === 'blob' && t.path.startsWith(docsPath) && (t.path.endsWith('.md') || t.path.endsWith('.mdx'))).map(t => t.path.slice(docsPath.length)) ?? []

const filePaths = INCLUDE_PREFIXES.length
  ? allDocFiles.filter(fp =>
    INCLUDE_PREFIXES.some(prefix => fp === prefix || fp.startsWith(prefix + '/'))
  )
  : allDocFiles

const { mdxPages, pageMap: _pageMap } = convertToPageMap({
  filePaths,
  basePath
})

function normalizeRoute(noExtPath: string) {
  let r = noExtPath;
  // strip folder index files
  r = r.replace(/\/(readme|index)$/i, '');
  r = r.replace(/^(readme|index)$/i, '');
  return r;
}

const routeMap: Record<string, string> = {};
for (const fp of filePaths) {
  const noExt = fp.replace(/\.(md|mdx)$/i, '');
  const norm = normalizeRoute(noExt);

  routeMap[noExt] = fp;
  if (!noExt.startsWith('content/')) {
    routeMap[`content/${noExt}`] = fp;
  }

  // clean route without "content/" and without README/index
  const isIndex = /\/(readme|index)$/i.test(noExt) || /^(readme|index)$/i.test(noExt);
  if (!routeMap[norm] || isIndex) routeMap[norm] = fp;

  // expose normalized route with 'content/' prefix 
  if (norm !== '' && !norm.startsWith('content/')) {
    const contentNorm = `content/${norm}`;
    if (!routeMap[contentNorm] || isIndex) routeMap[contentNorm] = fp;
  }

}

export const pageMap = normalizePageMap(_pageMap)

const { wrapper: Wrapper, ...components } = getMDXComponents({
  $Tabs: Tabs,
  Callout
})

type PageProps = Readonly<{
  params: Promise<{
    slug?: string[]
  }>
}>

export default async function Page(props: PageProps) {
  const params = await props.params
  const route = params.slug ? params.slug.join('/') : ''


  console.log(route);

  // Use normalized map instead of mdxPages
  const filePath =
    routeMap[route] ??
    [`${route}.mdx`, `${route}.md`, `${route}/README.md`, `${route}/readme.md`, `${route}/index.mdx`, `${route}/index.md`]
      .find(p => filePaths.includes(p))

  if (!filePath) {
    notFound()
  }

  const response = await fetch(
    `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${docsPath}${filePath}`,
    { headers, cache: 'force-cache' }
  )
  if (!response.ok) notFound()

  const data = await response.text()
  const processedData = data
    // Convert HTML comments to JSX comments
    .replace(/<!--/g, '{/*')
    .replace(/-->/g, '*/}')
    .replace(/<br\s*\/?>/gi, '<br />')
    .replace(/\{\{.*?\}\}/g, '')
    .replace(/style="border: 0"/g, 'style={{border:"0"}}')
    .replace(/align=center/g, 'align="center"')
    .replace(/frameborder="0"/g, 'frameBorder="0"')
    .replace(/allowfullscreen/g, 'allowFullScreen')
    .replace(/scrolling=no/g, 'scrolling="no"')
    .replace(/onload="[^"]*"/g, '')
    .replace(/<style\b[^>]*>[\s\S]*?<\/style>/gi, '')
    .replace(/<\/?ol>/g, '')
    .replace(/<\/?li>/g, '')
  const rawJs = await compileMdx(processedData, { filePath })
  const { default: MDXContent, toc, metadata } = evaluate(rawJs, components)

  return (
    <Wrapper toc={toc} metadata={metadata} sourceCode={rawJs}>
      <MDXContent />
    </Wrapper>
  )
}

export function generateStaticParams() {
  return Object.keys(routeMap)
    .filter(k => k !== '')
    .map(route => ({ slug: route.split('/') }))
}