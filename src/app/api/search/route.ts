import { NextRequest, NextResponse } from "next/server";
import fs from "fs";
import path from "path";

interface SearchResult {
  title: string;
  url: string;
  category: string;
  content: string;
  snippet: string;
  highlightedSnippet: string;
  matchType: "title" | "content" | "category";
}

// Function to extract text content from MDX
function extractTextFromMDX(content: string): string {
  // Remove code blocks
  let text = content.replace(/```[\s\S]*?```/g, "");

  // Remove inline code
  text = text.replace(/`[^`]+`/g, "");

  // Remove HTML comments
  text = text.replace(/<!--[\s\S]*?-->/g, "");

  // Remove markdown links but keep text: [text](url) -> text
  text = text.replace(/\[([^\]]+)\]\([^\)]+\)/g, "$1");

  // Remove images: ![alt](url)
  text = text.replace(/!\[([^\]]*)\]\([^\)]+\)/g, "");

  // Remove headers markup: ## Header -> Header
  text = text.replace(/^#{1,6}\s+(.+)$/gm, "$1");

  // Remove bold/italic: **text** or *text* -> text
  text = text.replace(/\*\*([^\*]+)\*\*/g, "$1");
  text = text.replace(/\*([^\*]+)\*/g, "$1");

  // Remove horizontal rules
  text = text.replace(/^---+$/gm, "");

  // Remove extra whitespace
  text = text.replace(/\n\s*\n/g, "\n");
  text = text.trim();

  return text;
}

// Function to extract title from MDX (first # heading)
function extractTitle(content: string): string {
  const match = content.match(/^#\s+(.+)$/m);
  return match ? match[1] : "Untitled";
}

// Function to read all MDX files recursively
function getAllMDXFiles(
  dir: string,
  baseDir: string = dir
): Array<{ path: string; relativePath: string }> {
  const files: Array<{ path: string; relativePath: string }> = [];

  try {
    const items = fs.readdirSync(dir);

    for (const item of items) {
      const fullPath = path.join(dir, item);
      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
        // Recursively read subdirectories
        files.push(...getAllMDXFiles(fullPath, baseDir));
      } else if (item.endsWith(".mdx") || item.endsWith(".md")) {
        const relativePath = path.relative(baseDir, fullPath);
        files.push({ path: fullPath, relativePath });
      }
    }
  } catch (error) {
    console.error("Error reading directory:", error);
  }

  return files;
}

// Convert file path to URL path
function filePathToUrl(relativePath: string): string {
  // Remove file extension
  let url = relativePath.replace(/\.(mdx|md)$/, "");

  // Remove 'page' from path (Next.js convention)
  url = url.replace(/\/page$/, "");
  url = url.replace(/^page$/, "");

  // Ensure it starts with /docs
  if (!url.startsWith("/")) {
    url = "/" + url;
  }

  // If it's empty, it's the root docs page
  if (url === "/" || url === "") {
    return "/docs";
  }

  return "/docs" + url;
}

export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams;
    const query = searchParams.get("q")?.toLowerCase() || "";

    if (!query.trim()) {
      return NextResponse.json({ results: [] });
    }

    // Path to docs directory
    const docsDir = path.join(process.cwd(), "src", "app", "docs");

    // Get all MDX files
    const mdxFiles = getAllMDXFiles(docsDir);

    const results: SearchResult[] = [];

    for (const { path: filePath, relativePath } of mdxFiles) {
      try {
        const fileContent = fs.readFileSync(filePath, "utf-8");
        const title = extractTitle(fileContent);
        const textContent = extractTextFromMDX(fileContent);
        const url = filePathToUrl(relativePath);

        // Determine category based on path
        let category = "Documentation";
        if (relativePath.includes("getting-started"))
          category = "Getting Started";
        if (relativePath.includes("tutorial")) category = "Tutorial";
        if (relativePath.includes("api")) category = "API Reference";
        if (relativePath.includes("guide")) category = "Guide";

        // Check if query matches
        const titleMatch = title.toLowerCase().includes(query);
        const contentMatch = textContent.toLowerCase().includes(query);

        if (titleMatch || contentMatch) {
          let snippet = "";
          let highlightedSnippet = "";

          if (contentMatch) {
            // Find the position of the match in content
            const contentLower = textContent.toLowerCase();
            const matchIndex = contentLower.indexOf(query);
            const start = Math.max(0, matchIndex - 60);
            const end = Math.min(
              textContent.length,
              matchIndex + query.length + 80
            );
            snippet =
              (start > 0 ? "..." : "") +
              textContent.slice(start, end) +
              (end < textContent.length ? "..." : "");

            // Highlight the matched term (case-insensitive)
            const regex = new RegExp(
              `(${query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`,
              "gi"
            );
            highlightedSnippet = snippet.replace(regex, "<mark>$1</mark>");
          } else {
            // If only title matches, show beginning of content
            snippet =
              textContent.slice(0, 120) +
              (textContent.length > 120 ? "..." : "");
            highlightedSnippet = snippet;
          }

          results.push({
            title,
            url,
            category,
            content: textContent,
            snippet,
            highlightedSnippet,
            matchType: titleMatch ? "title" : "content",
          });
        }
      } catch (error) {
        console.error(`Error processing file ${filePath}:`, error);
      }
    }

    // Sort results: title matches first, then by relevance
    results.sort((a, b) => {
      if (a.matchType === "title" && b.matchType !== "title") return -1;
      if (a.matchType !== "title" && b.matchType === "title") return 1;
      return 0;
    });

    return NextResponse.json({ results, count: results.length });
  } catch (error) {
    console.error("Search API error:", error);
    return NextResponse.json(
      { error: "Failed to perform search" },
      { status: 500 }
    );
  }
}
