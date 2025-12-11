/**
 * Get the base URL for the application.
 * In preview/deploy contexts, uses the current host.
 * In production, uses the configured production URL.
 * 
 * This allows links to work correctly in Netlify preview deployments
 * while maintaining proper URLs in production.
 */
export function getBaseUrl(): string {
  // Server-side: use environment variable or default
  if (typeof window === 'undefined') {
    return process.env.NEXT_PUBLIC_BASE_URL || 'https://kubestellar.io';
  }
  
  // Client-side: detect if we're on a preview deployment
  const host = window.location.host;
  const protocol = window.location.protocol;
  
  // Check if we're on a Netlify preview or other non-production domain
  if (
    host.includes('netlify.app') ||
    host.includes('previews.kubestellar.io') ||
    host.includes('localhost') ||
    host.includes('127.0.0.1')
  ) {
    // Use the current host for preview/local environments
    return `${protocol}//${host}`;
  }
  
  // Default to production URL
  return process.env.NEXT_PUBLIC_BASE_URL || 'https://kubestellar.io';
}

/**
 * Convert an absolute URL to use the current base URL if it's a kubestellar.io URL.
 * External URLs are left unchanged.
 * 
 * @param url - The URL to convert (can be relative or absolute)
 * @returns The URL adjusted for the current environment
 */
export function getLocalizedUrl(url: string): string {
  // If it's already a relative URL, return as-is
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    return url;
  }
  
  // Parse the URL to check the hostname
  try {
    const urlObj = new URL(url);
    
    // If it's a kubestellar.io or docs.kubestellar.io URL, replace with current base
    if (urlObj.hostname === 'kubestellar.io' || urlObj.hostname === 'docs.kubestellar.io') {
      const baseUrl = getBaseUrl();
      return `${baseUrl}${urlObj.pathname}${urlObj.search}${urlObj.hash}`;
    }
  } catch (error) {
    // If URL parsing fails, return the original URL
    console.error('Failed to parse URL:', url, error);
    return url;
  }
  
  // Return external URLs unchanged
  return url;
}
