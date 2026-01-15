/**
 * Shared cookie utilities for persisting preferences across Netlify branch deploys.
 *
 * Branch deploys share the domain *.kubestellar-docs.netlify.app, so cookies
 * set with domain=.kubestellar-docs.netlify.app are shared across all versions.
 *
 * Note: This does NOT work for kubestellar.io (production) which is a different domain.
 * Production will fall back to localStorage.
 */

// Cookie names
export const COOKIE_THEME = 'ks-theme';
export const COOKIE_BANNER_DISMISSED = 'ks-banner-dismissed';

// Max age: 1 year
const MAX_AGE = 365 * 24 * 60 * 60;

/**
 * Get the shared cookie domain for Netlify branch deploys.
 * Returns undefined for non-Netlify domains (falls back to current domain).
 */
function getSharedDomain(): string | undefined {
  if (typeof window === 'undefined') return undefined;

  const hostname = window.location.hostname;

  // Netlify branch deploys: {slug}--kubestellar-docs.netlify.app
  if (hostname.endsWith('.netlify.app') && hostname.includes('kubestellar-docs')) {
    return '.kubestellar-docs.netlify.app';
  }

  // For other domains (kubestellar.io, localhost), don't set domain
  // This lets the cookie default to the current host
  return undefined;
}

/**
 * Set a cookie with optional shared domain for Netlify deploys.
 */
export function setCookie(name: string, value: string): void {
  if (typeof document === 'undefined') return;

  const domain = getSharedDomain();
  const domainPart = domain ? `; domain=${domain}` : '';

  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${MAX_AGE}; SameSite=Lax${domainPart}`;
}

/**
 * Get a cookie value.
 */
export function getCookie(name: string): string | undefined {
  if (typeof document === 'undefined') return undefined;

  const cookies = document.cookie.split(';');
  for (const cookie of cookies) {
    const [cookieName, cookieValue] = cookie.trim().split('=');
    if (cookieName === name) {
      return decodeURIComponent(cookieValue);
    }
  }
  return undefined;
}

/**
 * Delete a cookie.
 */
export function deleteCookie(name: string): void {
  if (typeof document === 'undefined') return;

  const domain = getSharedDomain();
  const domainPart = domain ? `; domain=${domain}` : '';

  document.cookie = `${name}=; path=/; max-age=0${domainPart}`;
}
