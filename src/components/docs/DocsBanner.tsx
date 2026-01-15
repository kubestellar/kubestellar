'use client'

import Link from 'next/link'
import { useTheme } from 'next-themes'
import { useState, useEffect } from 'react'
import { getLocalizedUrl } from '@/lib/url'
import { useDocsMenu } from './DocsProvider'

export function DocsBanner() {
  const { resolvedTheme } = useTheme()
  const [mounted, setMounted] = useState(false)
  const { bannerDismissed, dismissBanner } = useDocsMenu()

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted || bannerDismissed) {
    return null
  }

  const isDark = resolvedTheme === 'dark'

  return (
    <div className={`relative ${isDark ? 'bg-neutral-900' : 'bg-blue-50'} border-b ${isDark ? 'border-gray-800' : 'border-blue-100'}`}>
      <div className="max-w-7xl mx-auto py-2 px-3 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between flex-wrap">
          <div className="flex-1 flex items-center justify-center">
            <span className={`${isDark ? 'text-gray-200' : 'text-gray-800'} text-xs md:text-sm`}>
              🚀 🚀 🚀 ATTENTION: KubeStellar needs your help - please take our 2-minute survey{' '}
              <Link 
                href={getLocalizedUrl("https://kubestellar.io/survey")}
                target="_blank" 
                rel="noopener noreferrer"
                className={isDark 
                  ? 'text-blue-400 underline hover:text-blue-300 transition-colors font-medium'
                  : 'text-blue-600 underline hover:text-blue-700 transition-colors font-medium'
                }
              >
                {getLocalizedUrl("https://kubestellar.io/survey")}
              </Link>
              {' '}🚀 🚀 🚀
            </span>
          </div>
          <button
            onClick={dismissBanner}
            className={`shrink-0 ml-3 p-1 rounded-md hover:bg-opacity-20 ${
              isDark ? 'hover:bg-gray-700' : 'hover:bg-gray-200'
            } focus:outline-none focus:ring-2 focus:ring-blue-500`}
            aria-label="Dismiss banner"
          >
            <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  )
}
