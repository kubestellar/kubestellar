import { notFound } from 'next/navigation'
import { compileMdx } from 'nextra/compile'
import { Callout, Tabs } from 'nextra/components'
import { evaluate } from 'nextra/evaluate'
import { useMDXComponents as getMDXComponents } from '../../../../../../mdx-components'
import { convertHtmlScriptsToJsxComments } from '@/lib/transformMdx'
import { MermaidComponent } from '@/lib/Mermaid'
import { buildPageMap, getContentPath } from '@/app/docs/page-map'
import { CURRENT_VERSION } from '@/config/versions'
import fs from 'fs'
import path from 'path'

// This page handles kubectl-claude docs with locale support
const projectId = 'kubectl-claude' as const

export const dynamic = 'force-static'
export const revalidate = false

const { wrapper: Wrapper, ...components } = getMDXComponents({ $Tabs: Tabs, Callout })
const component = { ...components, Mermaid: MermaidComponent }

type PageProps = Readonly<{
  params: Promise<{ locale: string; slug?: string[] }>
}>

function resolvePath(baseFile: string, relativePath: string) {
  if (relativePath.startsWith('/')) return relativePath.slice(1)
  const stack = baseFile.split('/')
  stack.pop()
  const parts = relativePath.split('/')
  for (const part of parts) {
    if (part === '.') continue
    if (part === '..') {
      if (stack.length > 0) stack.pop()
    } else {
      stack.push(part)
    }
  }
  const resolved = stack.join('/')
  if (resolved === '' || resolved.startsWith('..')) {
    return parts[parts.length - 1]
  }
  return resolved
}

function wrapMarkdownImagesWithFigures(markdown: string) {
  const imageRegex = /^!\[([^\]]*)\]\(([^)\s]+)(?:\s+"([^"]*)")?\)\s*$/gm
  return markdown.replace(imageRegex, (match, alt, src, title) => {
    const safeAlt = alt || ''
    const safeSrc = src || ''
    const safeTitle = title || ''
    const captionText = safeTitle || safeAlt
    const titleAttr = safeTitle ? ` title="${safeTitle}"` : ''
    const figcaptionElement = captionText ? `\n  <figcaption>${captionText}</figcaption>` : ''
    return `<figure className="ks-doc-figure">
  <img src="${safeSrc}" alt="${safeAlt}"${titleAttr} />${figcaptionElement}
</figure>`
  })
}

function wrapBadgeLinksInGrid(markdown: string) {
  const badgePattern = /\[!\[([^\]]*)\]\(([^)]*(?:shields\.io|badge|deepwiki)[^)]*)\)\]\(([^)]*)\)/gi
  const allBadges: Array<{ fullMatch: string; startIndex: number; endIndex: number }> = []
  let match
  while ((match = badgePattern.exec(markdown)) !== null) {
    allBadges.push({
      fullMatch: match[0],
      startIndex: match.index,
      endIndex: match.index + match[0].length
    })
  }
  if (allBadges.length === 0) return markdown
  const groups: string[][] = []
  let currentGroup: string[] = []
  let lastEndIndex = -1
  for (const badge of allBadges) {
    if (currentGroup.length === 0 || badge.startIndex - lastEndIndex < 200) {
      currentGroup.push(badge.fullMatch)
    } else {
      if (currentGroup.length > 0) groups.push(currentGroup)
      currentGroup = [badge.fullMatch]
    }
    lastEndIndex = badge.endIndex
  }
  if (currentGroup.length > 0) groups.push(currentGroup)
  let result = markdown
  let offset = 0
  for (const group of groups) {
    if (group.length > 0) {
      const badgesToWrap = group.slice(0, 9)
      const firstBadge = badgesToWrap[0]
      const lastBadge = badgesToWrap[badgesToWrap.length - 1]
      const firstIndex = result.indexOf(firstBadge, offset)
      if (firstIndex !== -1) {
        const lastIndex = result.indexOf(lastBadge, firstIndex) + lastBadge.length
        const beforeSection = result.substring(0, firstIndex)
        const afterSection = result.substring(lastIndex)
        const wrapped = `<div className="badge-grid-container">\n${badgesToWrap.map(b => `  <p>${b}</p>`).join('\n')}\n</div>`
        result = beforeSection + wrapped + afterSection
        offset = beforeSection.length + wrapped.length
      }
    }
  }
  return result
}

function removeCommentPatterns(content: string): string {
  let cleaned = content
  cleaned = cleaned.replace(/<!--[\s\S]*?-->/g, '')
  cleaned = cleaned.replace(/\{#[\s\S]*?#\}/g, '')
  cleaned = cleaned.replace(/\{\/[^}]*\/\}/g, '')
  return cleaned
}

function sanitizeHtmlForMdx(content: string): string {
  let sanitized = content
  sanitized = sanitized.replace(/<table>[\s\S]*?<\/table>/gi, (tableMatch) => {
    const contributors: Array<{ name: string; github: string; avatar: string; profileUrl: string }> = []
    const tdRegex = /<td[^>]*>[\s\S]*?<a href="([^"]+)"[^>]*><img src="([^"]+)"[^>]*\/?><br\s*\/?><sub><b>([^<]+)<\/b><\/sub><\/a>[\s\S]*?<\/td>/gi
    let tdMatch
    while ((tdMatch = tdRegex.exec(tableMatch)) !== null) {
      const profileUrl = tdMatch[1]
      const avatar = tdMatch[2]
      const name = tdMatch[3]
      const githubMatch = profileUrl.match(/github\.com\/([^\/]+)/)
      const github = githubMatch ? githubMatch[1] : ''
      if (name && avatar) {
        contributors.push({ name, github, avatar, profileUrl })
      }
    }
    if (contributors.length === 0) return ''
    const cards = contributors.map(c =>
      `<a href="${c.profileUrl}" className="contributor-card" target="_blank" rel="noopener noreferrer">
        <img src="${c.avatar}" alt="${c.name}" />
        <span>${c.name}</span>
      </a>`
    ).join('\n')
    return `<div className="contributors-grid">\n${cards}\n</div>`
  })
  sanitized = sanitized.replace(/<tr>[\s\S]*?<\/tr>/gi, '')
  sanitized = sanitized.replace(/<td[^>]*>[\s\S]*?<\/td>/gi, '')
  sanitized = sanitized.replace(/<iframe[\s\S]*?<\/iframe>/gi, '')
  sanitized = sanitized.replace(/<img\s+([^>]*?)\/?>/gi, (match, attrs) => {
    const srcMatch = attrs.match(/src=["']([^"']+)["']/i)
    const altMatch = attrs.match(/alt=["']([^"']*)["']/i)
    const titleMatch = attrs.match(/title=["']([^"']*)["']/i)
    const src = srcMatch ? srcMatch[1] : ''
    const alt = altMatch ? altMatch[1] : ''
    const title = titleMatch ? ` title="${titleMatch[1]}"` : ''
    if (!src) return ''
    return `<img src="${src}" alt="${alt}"${title} />`
  })
  sanitized = sanitized.replace(/<br\s*\/?>/gi, '<br />')
  sanitized = sanitized.replace(/<hr\s*\/?>/gi, '<hr />')
  sanitized = sanitized.replace(/\s+on\w+\s*=\s*["'][^"']*["']/gi, '')
  sanitized = sanitized.replace(/\s+on\w+\s*=\s*"[^"]*"/gi, '')
  sanitized = sanitized.replace(/\s+on\w+\s*=\s*'[^']*'/gi, '')
  sanitized = sanitized.replace(/\s+on\w+\s*=\s*["'][\s\S]*?["'](?=\s|>)/gi, '')
  sanitized = sanitized.replace(/\s+align=["']?[^"'\s>]+["']?/gi, '')
  sanitized = sanitized.replace(/\s+width=["']?[^"'\s>]+["']?/gi, '')
  sanitized = sanitized.replace(/\s+height=["']?[^"'\s>]+["']?/gi, '')
  sanitized = sanitized.replace(/\s+frameborder=["']?[^"'\s>]+["']?/gi, '')
  sanitized = sanitized.replace(/\s+allowfullscreen(?:=["'][^"']*["'])?/gi, '')
  sanitized = sanitized.replace(/\s+scrolling=["']?[^"'\s>]+["']?/gi, '')
  sanitized = sanitized.replace(/\sclass=/gi, ' className=')
  sanitized = sanitized.replace(/\s+style=["'][^"']*["']/gi, '')
  sanitized = sanitized.replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '')
  sanitized = sanitized.replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '')
  sanitized = sanitized.replace(/<sub>/gi, '')
  sanitized = sanitized.replace(/<\/sub>/gi, '')
  return sanitized
}

function replaceTemplateVariables(content: string): string {
  const version = CURRENT_VERSION as string
  const versionBranch = version === '0.29.0' ? 'main' : `release-${version}`
  const versionTag = version === '0.29.0' ? 'latest' : `v${version}`
  const vars: Record<string, string> = {
    'config.ks_branch': versionBranch,
    'config.ks_tag': versionTag,
    'config.ks_latest_release': CURRENT_VERSION,
    'config.ks_latest_regular_release': CURRENT_VERSION,
    'config.docs_url': 'https://docs.kubestellar.io',
    'config.repo_url': 'https://github.com/kubestellar/kubestellar',
    'config.site_url': 'https://docs.kubestellar.io'
  }
  let result = content
  for (const [key, value] of Object.entries(vars)) {
    result = result.replace(new RegExp(`\\{\\{\\s*${key.replace('.', '\\.')}\\s*\\}\\}`, 'g'), value)
  }
  result = result.replace(/\{\{[^}]+\}\}/g, '')
  return result
}

function readLocalFile(filePath: string, contentPath: string): string | null {
  const fullPath = path.join(contentPath, filePath)
  try {
    if (fs.existsSync(fullPath)) {
      return fs.readFileSync(fullPath, 'utf-8')
    }
  } catch {
    // File doesn't exist
  }
  return null
}

export default async function Page(props: PageProps) {
  const params = await props.params
  const { locale } = params
  const slug = params.slug ?? []
  const route = slug.join('/')

  const contentPath = getContentPath(projectId)
  const { routeMap, filePaths, pageMap } = buildPageMap(projectId)

  const filePath =
    routeMap[route] ??
    [`${route}.mdx`, `${route}.md`, `${route}/README.md`, `${route}/readme.md`, `${route}/index.mdx`, `${route}/index.md`]
      .find(p => filePaths.includes(p))

  if (!filePath) notFound()

  const rawText = readLocalFile(filePath, contentPath)
  if (!rawText) notFound()

  let processedContent = removeCommentPatterns(rawText)

  // Process Jekyll-style includes
  const includeRegex = /{%\s*include\s+["']([^"']+)["']\s*%}/g
  const includeMatches = Array.from(processedContent.matchAll(includeRegex))
  if (includeMatches.length > 0) {
    for (const match of includeMatches) {
      const [fullMatch, relativePath] = match
      const resolvedPath = resolvePath(filePath, relativePath)
      const includeContent = readLocalFile(resolvedPath, contentPath)
      if (includeContent) {
        processedContent = processedContent.replace(fullMatch, removeCommentPatterns(includeContent))
      } else {
        processedContent = processedContent.replace(fullMatch, `> **Note**: Include file \`${relativePath}\` not found`)
      }
    }
  }

  // Process full markdown includes
  const fullIncludeMarkdownRegex = /{%-?\s*include-markdown\s+["']([^"']+)["']\s*-?%}/g
  const fullIncludeMarkdownMatches = Array.from(processedContent.matchAll(fullIncludeMarkdownRegex))
  if (fullIncludeMarkdownMatches.length > 0) {
    for (const match of fullIncludeMarkdownMatches) {
      const [fullMatch, relativePath] = match
      if (fullMatch.includes('start=') || fullMatch.includes('end=')) continue
      const resolvedPath = resolvePath(filePath, relativePath)
      const includeContent = readLocalFile(resolvedPath, contentPath)
      if (includeContent) {
        processedContent = processedContent.replace(fullMatch, removeCommentPatterns(includeContent))
      } else {
        processedContent = processedContent.replace(fullMatch, `> **Note**: Include file \`${relativePath}\` not found`)
      }
    }
  }

  // Process partial includes
  const includeMarkdownRegex = /{%-?\s*include-markdown\s+["']([^"']+)["']\s+start=["']([^"']+)["']\s+end=["']([^"']+)["']\s*-?%}/g
  const includeMarkdownMatches = Array.from(processedContent.matchAll(includeMarkdownRegex))
  if (includeMarkdownMatches.length > 0) {
    for (const match of includeMarkdownMatches) {
      const [fullMatch, relativePath, startMarker, endMarker] = match
      const resolvedPath = resolvePath(filePath, relativePath)
      const includeContent = readLocalFile(resolvedPath, contentPath)
      if (includeContent) {
        const startIndex = includeContent.indexOf(startMarker)
        const endIndex = includeContent.indexOf(endMarker)
        if (startIndex !== -1 && endIndex !== -1) {
          const extractedContent = includeContent.substring(startIndex + startMarker.length, endIndex).trim()
          processedContent = processedContent.replace(fullMatch, removeCommentPatterns(extractedContent))
        } else {
          processedContent = processedContent.replace(fullMatch, `> **Note**: Markers not found in \`${relativePath}\``)
        }
      } else {
        processedContent = processedContent.replace(fullMatch, `> **Note**: Include file \`${relativePath}\` not found`)
      }
    }
  }

  const filePathToRoute = new Map<string, string>()
  Object.entries(routeMap).forEach(([r, fp]) => {
    if (!filePathToRoute.has(fp)) {
      filePathToRoute.set(fp, r)
    }
  })

  // Link base path includes locale
  const linkBasePath = `/${locale}/docs/kubectl-claude`

  // Rewrite Markdown links/images
  let rewrittenText = processedContent.replace(/(!?\[.*?\])\((.*?)\)/g, (match, label, link) => {
    if (/^(http|https|mailto:|#)/.test(link)) return match
    const isImage = label.startsWith('!')
    const [linkUrl, linkHash] = link.split('#')
    const resolvedPath = resolvePath(filePath, linkUrl)
    if (isImage) {
      const isImageFile = /\.(png|jpg|jpeg|gif|svg|webp|ico)$/i.test(resolvedPath)
      if (isImageFile) {
        const imgPath = `/docs-images/kubectl-claude/${resolvedPath}`
        return `${label}(${imgPath})`
      }
      return match
    } else {
      let targetRoute = filePathToRoute.get(resolvedPath)
      if (!targetRoute) targetRoute = filePathToRoute.get(resolvedPath + '.md')
      if (!targetRoute) targetRoute = filePathToRoute.get(resolvedPath + '.mdx')
      if (targetRoute) {
        return `${label}(${linkBasePath}/${targetRoute}${linkHash ? '#' + linkHash : ''})`
      }
      return match
    }
  })

  rewrittenText = rewrittenText.replace(/<img\s+([^>]*?)src=["']([^"']+)["']([^>]*?)\/?>/gi, (match, pre, src, post) => {
    if (/^(http|https|mailto:|#|data:)/.test(src)) return match
    const resolvedPath = resolvePath(filePath, src)
    const isImageFile = /\.(png|jpg|jpeg|gif|svg|webp|ico)$/i.test(resolvedPath)
    if (isImageFile) {
      const imgPath = `/docs-images/kubectl-claude/${resolvedPath}`
      const altMatch = (pre + post).match(/alt=["']([^"']*)["']/i)
      const alt = altMatch ? altMatch[1] : ''
      return `<img src="${imgPath}" alt="${alt}" />`
    }
    return match
  })

  rewrittenText = wrapMarkdownImagesWithFigures(rewrittenText)
  rewrittenText = wrapBadgeLinksInGrid(rewrittenText)

  let preProcessedText = replaceTemplateVariables(rewrittenText)
  preProcessedText = preProcessedText.replace(/```\s*{([^}]+)}\s*\n/g, (_match, attrs) => {
    const normalizedAttrs = attrs.replace(/^\./, '').replace(/\s+\./g, ' ')
    return `\`\`\`${normalizedAttrs}\n`
  })

  let processedData = sanitizeHtmlForMdx(preProcessedText)
  processedData = convertHtmlScriptsToJsxComments(processedData)

  const rawJs = await compileMdx(processedData, { filePath })
  const { default: MDXContent, toc, metadata } = evaluate(rawJs, component)

  return (
    <Wrapper toc={toc} metadata={metadata} sourceCode={rawJs} pageMap={pageMap} filePath={filePath} projectId={projectId}>
      <MDXContent />
    </Wrapper>
  )
}

export async function generateStaticParams() {
  const locales = ['en', 'ja']
  const allParams: { locale: string; slug: string[] }[] = []

  const kubectlClaudeMap = buildPageMap('kubectl-claude')
  for (const route of Object.keys(kubectlClaudeMap.routeMap)) {
    if (route !== '') {
      for (const locale of locales) {
        allParams.push({ locale, slug: route.split('/') })
      }
    }
  }

  return allParams
}
