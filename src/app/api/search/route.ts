import { NextRequest, NextResponse } from "next/server";
import { convertHtmlScriptsToJsxComments } from "@/lib/transformMdx";
import {
  buildPageMapForBranch,
  makeGitHubHeaders,
  user,
  repo,
  docsPath,
  basePath,
} from "../../docs/page-map";
import {
  getBranchForVersion,
  getDefaultVersion,
  type VersionKey,
} from "@/config/versions";

interface SearchResult {
  title: string;
  url: string;
  category: string;
  content: string;
  snippet: string;
  highlightedSnippet: string;
  matchType: "title" | "content" | "category";
}

function toPlainText(content: string): string {
  let text = content;

  // Remove code blocks, inline code, comments
  text = text.replace(/```[\s\S]*?```/g, "");
  text = text.replace(/`[^`]+`/g, "");
  text = text.replace(/<!--[\s\S]*?-->/g, "");

  // Links/images
  text = text.replace(/\[([^\]]+)\]\([^\)]+\)/g, "$1");
  text = text.replace(/!\[([^\]]*)\]\([^\)]+\)/g, "");

  // Headings -> keep text
  text = text.replace(/^#{1,6}\s+(.+)$/gm, "$1");

  // Bold/italic
  text = text.replace(/\*\*([^\*]+)\*\*/g, "$1");
  text = text.replace(/\*([^\*]+)\*/g, "$1");
  text = text.replace(/_([^_]+)_/g, "$1");

  // HR
  text = text.replace(/^---+$/gm, "");

  // Strip residual HTML tags
  text = text.replace(/<\/?[^>]+>/g, "");

  // Collapse whitespace
  text = text.replace(/\n\s*\n/g, "\n").trim();
  return text;
}

function extractTitle(md: string, fallback: string): string {
  const m = md.match(/^#\s+(.+)$/m);
  return m ? m[1].trim() : fallback;
}

function routeKeyToUrl(routeKey: string): string {
  return routeKey ? `/${basePath}/${routeKey}` : `/${basePath}`;
}

async function mapLimit<T, R>(
  items: T[],
  limit: number,
  worker: (x: T) => Promise<R>
): Promise<R[]> {
  const results: R[] = [];
  let i = 0;
  const run = async () => {
    while (i < items.length) {
      const idx = i++;
      results[idx] = await worker(items[idx]);
    }
  };
  await Promise.all(Array(Math.min(limit, items.length)).fill(0).map(run));
  return results;
}

export async function GET(request: NextRequest) {
  try {
    const sp = request.nextUrl.searchParams;
    const queryRaw = sp.get("q") || "";
    const query = queryRaw.toLowerCase().trim();
    const version =
      (sp.get("version") as VersionKey | null) || getDefaultVersion();
    if (!query) return NextResponse.json({ results: [], count: 0 });

    const branch = getBranchForVersion(version);
    const { routeMap } = await buildPageMapForBranch(branch);

    const entries = Object.entries(routeMap) as Array<[string, string]>;

    const resultsArray = await mapLimit(
      entries,
      8,
      async ([routeKey, filePath]) => {
        try {
          const res = await fetch(
            `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${docsPath}${filePath}`,
            { headers: makeGitHubHeaders(), cache: "no-store" }
          );
          if (!res.ok) return null;

          const raw = await res.text();

          const preprocessed = convertHtmlScriptsToJsxComments(raw);

          const text = toPlainText(preprocessed);

          const fallbackTitle =
            routeKey
              .split("/")
              .pop()
              ?.replace(/-/g, " ")
              .replace(/\b\w/g, c => c.toUpperCase()) || "Untitled";
          const title = extractTitle(raw, fallbackTitle);

          const hay = text.toLowerCase();
          const titleMatch = title.toLowerCase().includes(query);
          const contentMatch = hay.includes(query);
          if (!titleMatch && !contentMatch) return null;

          let snippet = "";
          let highlightedSnippet = "";
          if (contentMatch) {
            const idx = hay.indexOf(query);
            const start = Math.max(0, idx - 60);
            const end = Math.min(text.length, idx + query.length + 80);
            snippet =
              (start > 0 ? "..." : "") +
              text.slice(start, end) +
              (end < text.length ? "..." : "");
            const rx = new RegExp(
              `(${query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`,
              "gi"
            );
            highlightedSnippet = snippet.replace(rx, "<mark>$1</mark>");
          } else {
            snippet = text.slice(0, 140) + (text.length > 140 ? "..." : "");
            highlightedSnippet = snippet;
          }

          const category =
            routeKey
              .split("/")[0]
              ?.replace(/-/g, " ")
              .replace(/\b\w/g, c => c.toUpperCase()) || "Documentation";

          const url = routeKeyToUrl(routeKey);

          const result: SearchResult = {
            title,
            url,
            category,
            content: text,
            snippet,
            highlightedSnippet,
            matchType: titleMatch ? "title" : "content",
          };
          return result;
        } catch {
          return null;
        }
      }
    );

    const results = resultsArray.filter((r): r is SearchResult => !!r);

    // Sort: title matches first
    results.sort((a, b) =>
      a.matchType === "title" && b.matchType !== "title"
        ? -1
        : a.matchType === b.matchType
          ? 0
          : 1
    );

    return NextResponse.json({ results, count: results.length });
  } catch (e) {
    console.error("Search API error:", e);
    return NextResponse.json(
      { error: "Failed to perform search" },
      { status: 500 }
    );
  }
}
