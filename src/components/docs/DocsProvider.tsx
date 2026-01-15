'use client'

import React, { createContext, useContext, useState, useRef, ReactNode } from 'react'

interface DocsContextType {
  menuOpen: boolean
  setMenuOpen: (open: boolean) => void
  toggleMenu: () => void
  sidebarCollapsed: boolean
  setSidebarCollapsed: (collapsed: boolean) => void
  toggleSidebar: () => void
  bannerDismissed: boolean
  setBannerDismissed: (dismissed: boolean) => void
  dismissBanner: () => void
  // Nav folder collapsed state - persists across navigation
  navCollapsed: Set<string>
  setNavCollapsed: React.Dispatch<React.SetStateAction<Set<string>>>
  toggleNavCollapsed: (key: string) => void
  navInitialized: React.MutableRefObject<boolean>
}

const DocsContext = createContext<DocsContextType | undefined>(undefined)

export function DocsProvider({ children }: { children: ReactNode }) {
  const [menuOpen, setMenuOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [bannerDismissed, setBannerDismissed] = useState(false)
  const [navCollapsed, setNavCollapsed] = useState<Set<string>>(new Set())
  const navInitialized = useRef(false)

  const toggleMenu = () => setMenuOpen((prev) => !prev)
  const toggleSidebar = () => setSidebarCollapsed((prev) => !prev)
  const dismissBanner = () => setBannerDismissed(true)
  const toggleNavCollapsed = (key: string) => {
    setNavCollapsed(prev => {
      const next = new Set(prev)
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
      return next
    })
  }

  return (
    <DocsContext.Provider value={{
      menuOpen,
      setMenuOpen,
      toggleMenu,
      sidebarCollapsed,
      setSidebarCollapsed,
      toggleSidebar,
      bannerDismissed,
      setBannerDismissed,
      dismissBanner,
      navCollapsed,
      setNavCollapsed,
      toggleNavCollapsed,
      navInitialized
    }}>
      {children}
    </DocsContext.Provider>
  )
}

export function useDocsMenu() {
  const context = useContext(DocsContext)
  if (context === undefined) {
    throw new Error('useDocsMenu must be used within a DocsProvider')
  }
  return context
}
