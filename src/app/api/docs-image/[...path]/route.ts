import { NextRequest, NextResponse } from 'next/server'
import fs from 'fs'
import path from 'path'

const DOCS_CONTENT_PATH = path.join(process.cwd(), 'docs', 'content')

const MIME_TYPES: Record<string, string> = {
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.svg': 'image/svg+xml',
  '.webp': 'image/webp',
  '.ico': 'image/x-icon',
}

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path: pathSegments } = await params
  const imagePath = pathSegments.join('/')
  
  // Security: prevent directory traversal
  if (imagePath.includes('..')) {
    return new NextResponse('Forbidden', { status: 403 })
  }
  
  const fullPath = path.join(DOCS_CONTENT_PATH, imagePath)
  
  // Check if file exists and is within docs/content
  if (!fullPath.startsWith(DOCS_CONTENT_PATH)) {
    return new NextResponse('Forbidden', { status: 403 })
  }
  
  try {
    if (!fs.existsSync(fullPath)) {
      return new NextResponse('Not Found', { status: 404 })
    }
    
    const ext = path.extname(fullPath).toLowerCase()
    const mimeType = MIME_TYPES[ext] || 'application/octet-stream'
    
    const fileBuffer = fs.readFileSync(fullPath)
    
    return new NextResponse(fileBuffer, {
      headers: {
        'Content-Type': mimeType,
        'Cache-Control': 'public, max-age=31536000, immutable',
      },
    })
  } catch {
    return new NextResponse('Internal Server Error', { status: 500 })
  }
}
