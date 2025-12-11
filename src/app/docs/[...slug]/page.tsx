import { notFound } from 'next/navigation'
import { compileMdx } from 'nextra/compile'
import { Callout, Tabs } from 'nextra/components'
import { evaluate } from 'nextra/evaluate'
import { useMDXComponents as getMDXComponents } from '../../../../mdx-components'
import { convertHtmlScriptsToJsxComments } from '@/lib/transformMdx'
import { MermaidComponent } from '@/lib/Mermaid'
import { buildPageMapForBranch, makeGitHubHeaders, user, repo, docsPath } from '../page-map'
import { getBranchForVersion, getDefaultVersion, type VersionKey } from '@/config/versions'
import EditViewSourceButtons from '@/components/docs/EditViewSourceButtons'

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

function wrapMarkdownImagesWithFigures(markdown: string) {
  // Skip images that are part of link badges: [![...](...)](...)
  const imageRegex = /(?<!\[)!\[([^\]]*)\]\(([^)\s]+)(?:\s+"([^"]*)")?\)/g;

  return markdown.replace(imageRegex, (_match, alt = '', src, title) => {
    const captionText = title || alt;
    const titleAttr = title ? ` title="${title}"` : '';
    const figcaption = captionText ? `<figcaption>${captionText}</figcaption>` : '';

    return `
<figure class="ks-doc-figure">
  <img src="${src}" alt="${alt}"${titleAttr} />
  ${figcaption}
</figure>
`;
  });
}

function wrapBadgeLinksInGrid(markdown: string) {
  const badgePattern = /\[!\[([^\]]*)\]\(([^)]*(?:shields\.io|badge|deepwiki)[^)]*)\)\]\(([^)]*)\)/gi;

  const allBadges: Array<{ fullMatch: string; startIndex: number; endIndex: number }> = [];
  let match;

  while ((match = badgePattern.exec(markdown)) !== null) {
    allBadges.push({
      fullMatch: match[0],
      startIndex: match.index,
      endIndex: match.index + match[0].length
    });
  }

  if (allBadges.length === 0) return markdown;

  const groups: string[][] = [];
  let currentGroup: string[] = [];
  let lastEndIndex = -1;

  for (const badge of allBadges) {
    if (currentGroup.length === 0 || badge.startIndex - lastEndIndex < 200) {
      currentGroup.push(badge.fullMatch);
    } else {
      if (currentGroup.length > 0) groups.push(currentGroup);
      currentGroup = [badge.fullMatch];
    }
    lastEndIndex = badge.endIndex;
  }
  if (currentGroup.length > 0) groups.push(currentGroup);

  let result = markdown;
  let offset = 0;

  for (const group of groups) {
    if (group.length > 0) {
      const badgesToWrap = group.slice(0, 9);
      const firstBadge = badgesToWrap[0];
      const lastBadge = badgesToWrap[badgesToWrap.length - 1];
      const firstIndex = result.indexOf(firstBadge, offset);

      if (firstIndex !== -1) {
        const lastIndex = result.indexOf(lastBadge, firstIndex) + lastBadge.length;
        const beforeSection = result.substring(0, firstIndex);
        const afterSection = result.substring(lastIndex);
        const wrapped = `<div class="badge-grid-container">\n${badgesToWrap.map(b => `  <p>${b}</p>`).join('\n')}\n</div>`;

        result = beforeSection + wrapped + afterSection;
        offset = beforeSection.length + wrapped.length;
      }
    }
  }

  return result;
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

  function removeCommentPatterns(content: string): string {
    let cleaned = content;
    cleaned = cleaned.replace(/\{\/Note[^}]*\/\}/g, '');
    cleaned = cleaned.replace(/\{\/ALL-CONTRIBUTORS-LIST[^}]*\/\}/gi, '');
    cleaned = cleaned.replace(/\{\/prettier-ignore[^}]*\/\}/gi, '');
    cleaned = cleaned.replace(/\{\/markdownlint[^}]*\/\}/gi, '');
    cleaned = cleaned.replace(/<!--[\s\S]*?(?:Note that this repo|ALL-CONTRIBUTORS-LIST|prettier-ignore|markdownlint)[\s\S]*?-->/gi, '');

    // Remove specific unwanted headings/comments
    cleaned = cleaned.replace(/\{\/Included in website\. Edit CONTRIBUTING\.md for GitHub\.\/\}/gi, '');
    cleaned = cleaned.replace(/<!--\s*Included in website\. Edit CONTRIBUTING\.md for GitHub\.\s*-->/gi, '');

    cleaned = cleaned.replace(/\{\/Canonical GitHub version\. Edit contributing-inc\.md for website\.\/\}/gi, '');
    cleaned = cleaned.replace(/<!--\s*Canonical GitHub version\. Edit contributing-inc\.md for website\.\s*-->/gi, '');

    cleaned = cleaned.replace(/\{\/A wrapper file to include the GOVERNANCE file from the repository root\/\}/gi, '');
    cleaned = cleaned.replace(/<!--\s*A wrapper file to include the GOVERNANCE file from the repository root\s*-->/gi, '');

    cleaned = cleaned.replace(/\{\/Code management[\s\S]*?Quay\.io[\s\S]*?\/\}/gi, '');
    cleaned = cleaned.replace(/<!--\s*Code management[\s\S]*?Quay\.io[\s\S]*?-->/gi, '');

    return cleaned;
  }

  // --- START PROCESSING INCLUDES ---
  let processedContent = removeCommentPatterns(rawText);

  // 1. Process Jekyll-style includes: {% include "path" %}
  const includeRegex = /{%\s*include\s+["']([^"']+)["']\s*%}/g;
  const includeMatches = Array.from(processedContent.matchAll(includeRegex));

  if (includeMatches.length > 0) {
    const uniqueIncludes = [...new Set(includeMatches.map(m => m[1]))];
    const includeContents = await Promise.all(uniqueIncludes.map(async (relativePath) => {
      const resolvedPath = resolvePath(filePath, relativePath);
      const url = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${docsPath}${resolvedPath}`;
      try {
        const res = await fetch(url, { headers: makeGitHubHeaders(), cache: 'no-store' });
        if (res.ok) return { path: relativePath, text: removeCommentPatterns(await res.text()) };
        const rootUrl = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${resolvedPath}`;
        const rootRes = await fetch(rootUrl, { headers: makeGitHubHeaders(), cache: 'no-store' });
        if (rootRes.ok) return { path: relativePath, text: removeCommentPatterns(await rootRes.text()) };

        // Suppress error for coming-soon.md
        if (relativePath.includes('coming-soon.md')) {
          return { path: relativePath, text: '' };
        }

        return { path: relativePath, text: `> **Error**: Could not include \`${relativePath}\` (File not found)` };
      } catch {
        return { path: relativePath, text: `> **Error**: Failed to fetch \`${relativePath}\`` };
      }
    }));
    includeContents.forEach(({ path, text }) => {
      const escapedPath = path.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const pattern = new RegExp(`{%\\s*include\\s+["']${escapedPath}["']\\s*%}`, 'g');
      processedContent = processedContent.replace(pattern, () => removeCommentPatterns(text));
    });
  }

  // 2. Process full markdown includes (without start/end): {% include-markdown "path" %}
  const fullIncludeMarkdownRegex = /{%-?\s*include-markdown\s+["']([^"']+)["']\s*-?%}/g;
  // This regex is designed to not match the version with start/end attributes.
  // We'll process these matches and remove them from the main string before the next step.
  const fullIncludeMarkdownMatches = Array.from(processedContent.matchAll(fullIncludeMarkdownRegex));

  if (fullIncludeMarkdownMatches.length > 0) {
    for (const match of fullIncludeMarkdownMatches) {
      const [fullMatch, relativePath] = match;
      // Skip if it's the more complex version
      if (fullMatch.includes('start=') || fullMatch.includes('end=')) continue;

      const resolvedPath = resolvePath(filePath, relativePath);
      const url = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${resolvedPath}`;
      try {
        const res = await fetch(url, { headers: makeGitHubHeaders(), cache: 'no-store' });
        if (res.ok) {
          const fileContent = await res.text();
          processedContent = processedContent.replace(fullMatch, () => removeCommentPatterns(fileContent));
        } else {
          // Suppress error for coming-soon.md
          if (relativePath.includes('coming-soon.md')) {
            processedContent = processedContent.replace(fullMatch, '');
          } else {
            processedContent = processedContent.replace(fullMatch, `> **Error**: Could not include \`${relativePath}\` (File not found)`);
          }
        }
      } catch {
        processedContent = processedContent.replace(fullMatch, `> **Error**: Failed to fetch \`${relativePath}\``);
      }
    }
  }

  // 3. Process partial includes: {% include-markdown "path" start="..." end="..." %}
  const includeMarkdownRegex = /{%-?\s*include-markdown\s+["']([^"']+)["']\s+start=["']([^"']+)["']\s+end=["']([^"']+)["']\s*-?%}/g;
  const includeMarkdownMatches = Array.from(processedContent.matchAll(includeMarkdownRegex));

  if (includeMarkdownMatches.length > 0) {
    for (const match of includeMarkdownMatches) {
      const [fullMatch, relativePath, startMarker, endMarker] = match;
      const resolvedPath = resolvePath(filePath, relativePath);
      const url = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${resolvedPath}`;
      try {
        const res = await fetch(url, { headers: makeGitHubHeaders(), cache: 'no-store' });
        if (res.ok) {
          const fileContent = await res.text();
          const startIndex = fileContent.indexOf(startMarker);
          const endIndex = fileContent.indexOf(endMarker);
          if (startIndex !== -1 && endIndex !== -1) {
            const extractedContent = fileContent.substring(startIndex + startMarker.length, endIndex).trim();
            processedContent = processedContent.replace(fullMatch, removeCommentPatterns(extractedContent));
          } else {
            processedContent = processedContent.replace(fullMatch, `> **Error**: Markers not found in \`${relativePath}\``);
          }
        } else {
          // Suppress error for coming-soon.md
          if (relativePath.includes('coming-soon.md')) {
            processedContent = processedContent.replace(fullMatch, '');
          } else {
            processedContent = processedContent.replace(fullMatch, `> **Error**: Could not include \`${relativePath}\` (File not found)`);
          }
        }
      } catch {
        processedContent = processedContent.replace(fullMatch, `> **Error**: Failed to fetch \`${relativePath}\``);
      }
    }
  }
  // --- END PROCESSING INCLUDES ---

  const filePathToRoute = new Map<string, string>();
  Object.entries(routeMap).forEach(([r, fp]) => filePathToRoute.set(fp, r));

  // Rewrite Markdown links/images using the fully processed content
  let rewrittenText = processedContent.replace(/(!?\[.*?\])\((.*?)\)/g, (match, label, link) => {
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

  rewrittenText = wrapMarkdownImagesWithFigures(rewrittenText);

  // Wrap badges in grid container before HTML processing
  rewrittenText = wrapBadgeLinksInGrid(rewrittenText);

  // 3. Pre-process Jinja and Pymdown syntax before MDX compilation
  const preProcessedText = rewrittenText
    // Replace Jinja-style variables {{ config.var_name }} with a placeholder.
    // This prevents MDX from trying to parse it as a JSX expression.
    .replace(/{{\s*config\.([\w_]+)\s*}}/g, (match, varName) => `[${varName}]`)
    // Convert Pymdown code block attributes into standard MDX attributes.
    // e.g., ``` {.bash .no-copy} -> ```bash .no-copy
    // e.g., ``` title="file.sh" -> ```sh title="file.sh"
    .replace(/```\s*{([^}]+)}\s*\n/g, (match, attrs) => {
      // Normalize attributes: remove leading dot, handle multiple attrs.
      const normalizedAttrs = attrs.replace(/^\./, '').replace(/\s+\./g, ' ');
      return `\`\`\`${normalizedAttrs}\n`;
    });


  const processedData = convertHtmlScriptsToJsxComments(preProcessedText)
    .replace(/<br\s*\/?>/gi, '<br />')
    .replace(/align=center/g, 'align="center"')
    .replace(/frameborder="0"/g, 'frameBorder="0"')
    .replace(/allowfullscreen/g, 'allowFullScreen')
    .replace(/scrolling=no/g, 'scrolling="no"')
    .replace(/onload="[^"]*"/g, '')
    .replace(/<style\b[^>]*>[\s\S]*?<\/style>/gi, '');

  const rawJs = await compileMdx(processedData, { filePath })
  const { default: MDXContent, toc, metadata } = evaluate(rawJs, component)

  return (
    <Wrapper toc={toc} metadata={metadata} sourceCode={rawJs}>
      <EditViewSourceButtons
        filePath={filePath}
        user={user}
        repo={repo}
        branch={branch}
        docsPath={docsPath}
      />
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