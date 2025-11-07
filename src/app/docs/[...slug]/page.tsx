import { notFound } from 'next/navigation'
import { compileMdx } from 'nextra/compile'
import { Callout, Tabs } from 'nextra/components'
import { evaluate } from 'nextra/evaluate'
import { useMDXComponents as getMDXComponents } from '../../../../mdx-components'
import { convertHtmlScriptsToJsxComments } from '@/lib/transformMdx'
import { MermaidComponent } from '@/lib/Mermaid'
import { buildPageMapForBranch, makeGitHubHeaders, user, repo, docsPath } from '../page-map'
import { getBranchForVersion, getDefaultVersion, type VersionKey } from '@/config/versions'

export const dynamic = 'force-dynamic'

const { wrapper: Wrapper, ...components } = getMDXComponents({ $Tabs: Tabs, Callout })
const component = { ...components, Mermaid: MermaidComponent }

type PageProps = Readonly<{
  params: Promise<{ slug?: string[] }>
  searchParams: Promise<{ version?: string }>
}>

export default async function Page(props: PageProps) {
  const params = await props.params
  const searchParams = await props.searchParams
  
  // Get version from URL or use default
  const version = (searchParams.version as VersionKey) || getDefaultVersion()
  const branch = getBranchForVersion(version)
  
  // Build page map for this branch
  const { routeMap, filePaths } = await buildPageMapForBranch(branch)
  
  const route = params.slug ? params.slug.join('/') : ''

  const filePath =
    routeMap[route] ??
    [`${route}.mdx`, `${route}.md`, `${route}/README.md`, `${route}/readme.md`, `${route}/index.mdx`, `${route}/index.md`]
      .find(p => filePaths.includes(p))

  if (!filePath) notFound()

  const response = await fetch(
    `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${docsPath}${filePath}`,
    { headers: makeGitHubHeaders(), cache: 'no-store' }
  )
  if (!response.ok) notFound()

  const data = await response.text()
  const processedData = convertHtmlScriptsToJsxComments(data)
    .replace(/<br\s*\/?>/gi, '<br />')
    .replace(/align=center/g, 'align="center"')
    .replace(/frameborder="0"/g, 'frameBorder="0"')
    .replace(/allowfullscreen/g, 'allowFullScreen')
    .replace(/scrolling=no/g, 'scrolling="no"')
    .replace(/onload="[^"]*"/g, '')
    .replace(/<style\b[^>]*>[\s\S]*?<\/style>/gi, '')
    .replace(/<\/?ol>/g, '')
    .replace(/<\/?li>/g, '')

  const rawJs = await compileMdx(processedData, { filePath })
  const { default: MDXContent, toc, metadata } = evaluate(rawJs, component)

  return (
    <Wrapper toc={toc} metadata={metadata} sourceCode={rawJs}>
      <MDXContent />
    </Wrapper>
  )
}

export async function generateStaticParams() {
  const defaultVersion = getDefaultVersion()
  const branch = getBranchForVersion(defaultVersion)
  const { routeMap } = await buildPageMapForBranch(branch)
  return Object.keys(routeMap).filter(k => k !== '').map(route => ({ slug: route.split('/') }))
}