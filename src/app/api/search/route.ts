import { NextRequest, NextResponse } from "next/server"
import { convertHtmlScriptsToJsxComments } from "@/lib/transformMdx"
import { buildPageMap, docsContentPath, basePath } from "../../docs/page-map"
import fs from 'fs'
import path from 'path'

interface SearchResult {
  title: string
  url: string
  category: string
  content: string
  snippet: string
  highlightedSnippet: string
  matchType: "title" | "content" | "category"
}

function toPlainText(content: string): string {
  let text = content

  // Remove code blocks, inline code, comments
  text = text.replace(/```[\s\S]*?```/g, "")
  text = text.replace(/`[^`]+`/g, "")
  text = text.replace(/<!--[\s\S]*?-->/g, "")

  // Links/images
  text = text.replace(/\[([^\]]+)\]\([^\)]+\)/g, "$1")
  text = text.replace(/!\[([^\]]*)\]\([^\)]+\)/g, "")

  // Headings -> keep text
  text = text.replace(/^#{1,6}\s+(.+)$/gm, "$1")

  // Bold/italic
  text = text.replace(/\*\*([^\*]+)\*\*/g, "$1")
  text = text.replace(/\*([^\*]+)\*/g, "$1")
  text = text.replace(/_([^_]+)_/g, "$1")

  // HR
  text = text.replace(/^---+$/gm, "")

  // Strip residual HTML tags
  text = text.replace(/<\/?[^>]+>/g, "")

  // Collapse whitespace
  text = text.replace(/\n\s*\n/g, "\n").trim()
  return text
}

function extractTitle(md: string, fallback: string): string {
  const m = md.match(/^#\s+(.+)$/m)
  return m ? m[1].trim() : fallback
}

function routeKeyToUrl(routeKey: string): string {
  return routeKey ? `/${basePath}/${routeKey}` : `/${basePath}`
}

function readLocalFile(filePath: string): string | null {
  const fullPath = path.join(docsContentPath, filePath)
  try {
    if (fs.existsSync(fullPath)) {
      return fs.readFileSync(fullPath, 'utf-8')
    }
  } catch {
    // File doesn't exist
  }
  return null
}

export async function GET(request: NextRequest) {
  try {
    const sp = request.nextUrl.searchParams
    const queryRaw = sp.get("q") || ""
    const query = queryRaw.toLowerCase().trim()
    if (!query) return NextResponse.json({ results: [], count: 0 })

    const { routeMap } = buildPageMap()

    const entries = Object.entries(routeMap) as Array<[string, string]>

    const results: SearchResult[] = []

    for (const [routeKey, filePath] of entries) {
      const raw = readLocalFile(filePath)
      if (!raw) continue

      const preprocessed = convertHtmlScriptsToJsxComments(raw)
      const text = toPlainText(preprocessed)

      const fallbackTitle =
        routeKey
          .split("/")
          .pop()
          ?.replace(/-/g, " ")
          .replace(/\b\w/g, c => c.toUpperCase()) || "Untitled"
      const title = extractTitle(raw, fallbackTitle)

      const hay = text.toLowerCase()
      const titleMatch = title.toLowerCase().includes(query)
      const contentMatch = hay.includes(query)
      if (!titleMatch && !contentMatch) continue

      let snippet = ""
      let highlightedSnippet = ""
      if (contentMatch) {
        const idx = hay.indexOf(query)
        const start = Math.max(0, idx - 60)
        const end = Math.min(text.length, idx + query.length + 80)
        snippet =
          (start > 0 ? "..." : "") +
          text.slice(start, end) +
          (end < text.length ? "..." : "")
        const rx = new RegExp(
          `(${query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`,
          "gi"
        )
        highlightedSnippet = snippet.replace(rx, "<mark>$1</mark>")
      } else {
        snippet = text.slice(0, 140) + (text.length > 140 ? "..." : "")
        highlightedSnippet = snippet
      }

      const category =
        routeKey
          .split("/")[0]
          ?.replace(/-/g, " ")
          .replace(/\b\w/g, c => c.toUpperCase()) || "Docs"

      results.push({
        title,
        url: routeKeyToUrl(routeKey),
        category,
        content: text.slice(0, 500),
        snippet,
        highlightedSnippet,
        matchType: titleMatch ? "title" : "content",
      })
    }

    // Sort: title matches first, then by title alphabetically
    results.sort((a, b) => {
      if (a.matchType === "title" && b.matchType !== "title") return -1
      if (a.matchType !== "title" && b.matchType === "title") return 1
      return a.title.localeCompare(b.title)
    })

    return NextResponse.json({
      results: results.slice(0, 20),
      count: results.length,
    })
  } catch (error) {
    console.error("Search error:", error)
    return NextResponse.json(
      { error: "Search failed", results: [], count: 0 },
      { status: 500 }
    )
  }
}
