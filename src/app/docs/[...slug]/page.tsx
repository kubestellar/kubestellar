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

function resolvePath(baseFile: string, relativePath: string) {
  if (relativePath.startsWith('/')) return relativePath.slice(1);
  const stack = baseFile.split('/');
  stack.pop(); // Remove current filename
  const parts = relativePath.split('/');
  for (const part of parts) {
    if (part === '.') continue;
    if (part === '..') {
      if (stack.length > 0) stack.pop();
    } else {
      stack.push(part);
    }
  }
  return stack.join('/');
}

export default async function Page(props: PageProps) {
  const params = await props.params
  const searchParams = await props.searchParams
  
  const version = (searchParams.version as VersionKey) || getDefaultVersion()
  const branch = getBranchForVersion(version)
  
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

  const rawText = await response.text()

  let contentWithIncludes = rawText;
  const includeRegex = /{%\s*include\s+["']([^"']+)["']\s*%}/g;
  const includeMatches = Array.from(rawText.matchAll(includeRegex));

  if (includeMatches.length > 0) {
    const uniqueIncludes = [...new Set(includeMatches.map(m => m[1]))];
    
    const includeContents = await Promise.all(uniqueIncludes.map(async (relativePath) => {
      const resolvedPath = resolvePath(filePath, relativePath);
      const url = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${docsPath}${resolvedPath}`;
      
      try {
        const res = await fetch(url, { headers: makeGitHubHeaders(), cache: 'no-store' });
        if (res.ok) {
            return { path: relativePath, text: await res.text() };
        }
        
        const rootUrl = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${resolvedPath}`;
        const rootRes = await fetch(rootUrl, { headers: makeGitHubHeaders(), cache: 'no-store' });
        if (rootRes.ok) {
            return { path: relativePath, text: await rootRes.text() };
        }

        return { path: relativePath, text: `> **Error**: Could not include \`${relativePath}\` (File not found)` };
      } catch {
        return { path: relativePath, text: `> **Error**: Failed to fetch \`${relativePath}\`` };
      }
    }));

    includeContents.forEach(({ path, text }) => {
      const escapedPath = path.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const pattern = new RegExp(`{%\\s*include\\s+["']${escapedPath}["']\\s*%}`, 'g');
      contentWithIncludes = contentWithIncludes.replace(pattern, () => text);
    });
  }

  const filePathToRoute = new Map<string, string>();
  Object.entries(routeMap).forEach(([r, fp]) => filePathToRoute.set(fp, r));

  let rewrittenText = contentWithIncludes.replace(/(!?\[.*?\])\((.*?)\)/g, (match, label, link) => {
    if (/^(http|https|mailto:|#)/.test(link)) return match;

    const isImage = label.startsWith('!');
    const [linkUrl, linkHash] = link.split('#');
    
    const resolvedPath = resolvePath(filePath, linkUrl);
    
    if (isImage) {
       const rawUrl = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${docsPath}${resolvedPath}`;
       return `${label}(${rawUrl})`;
    } else {
       let targetRoute = filePathToRoute.get(resolvedPath);
       if (!targetRoute) targetRoute = filePathToRoute.get(resolvedPath + '.md');
       if (!targetRoute) targetRoute = filePathToRoute.get(resolvedPath + '.mdx');
       
       if (targetRoute) {
         return `${label}(/docs/${targetRoute}${linkHash ? '#' + linkHash : ''})`;
       }
       
       return `${label}(https://raw.githubusercontent.com/${user}/${repo}/${branch}/${docsPath}${resolvedPath})`;
    }
  });

  rewrittenText = rewrittenText.replace(/<img\s+([^>]*?)src=["']([^"']+)["']([^>]*?)>/gi, (match, pre, src, post) => {
    if (/^(http|https|mailto:|#|data:)/.test(src)) return match;

    const resolvedPath = resolvePath(filePath, src);
    const rawUrl = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${docsPath}${resolvedPath}`;
    
    return `<img ${pre}src="${rawUrl}"${post}>`;
  });

  const processedData = convertHtmlScriptsToJsxComments(rewrittenText)
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