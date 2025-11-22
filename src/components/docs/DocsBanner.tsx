'use client'

import { Banner } from 'nextra/components'
import Link from 'next/link'
import { useTheme } from 'next-themes'
import { useState, useEffect } from 'react'

export function DocsBanner() {
  const { resolvedTheme } = useTheme()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return null
  }

  const isDark = resolvedTheme === 'dark'

  return (
    <Banner storageKey="kubestellar-demo" className={isDark ? '' : 'bg-blue-50'}>
      <span className={isDark ? 'text-gray-200' : 'text-gray-800'}>
        🚀 🚀 🚀 ATTENTION: KubeStellar needs your help - please take our 2-minute survey{' '}
        <Link 
          href="https://kubestellar.io/survey" 
          target="_blank" 
          rel="noopener noreferrer"
          className={isDark 
            ? 'text-blue-400 underline hover:text-blue-300 transition-colors font-medium'
            : 'text-blue-600 underline hover:text-blue-700 transition-colors font-medium'
          }
        >
          https://kubestellar.io/survey
        </Link>
        {' '}🚀 🚀 🚀
      </span>
    </Banner>
  )
}
