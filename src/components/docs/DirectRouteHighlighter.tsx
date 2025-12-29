'use client'

import { usePathname } from 'next/navigation'
import { useEffect } from 'react'

type DirectRouteHighlighterProps = {
  directRouteMap: Record<string, string>
}

export function DirectRouteHighlighter({ directRouteMap }: DirectRouteHighlighterProps) {
  const pathname = usePathname()
  
  useEffect(() => {
    // If current path is a direct route, find its canonical route
    const canonicalRoute = directRouteMap[pathname]
    
    if (canonicalRoute) {
      
      const activateLink = () => {
        // Find ALL links in the page
        const allLinks = document.querySelectorAll('a')
        
        allLinks.forEach(link => {
          const href = link.getAttribute('href')
          
          if (href === canonicalRoute) {
            // Remove inactive state classes
            link.classList.remove(
              'x:text-gray-600',
              'x:dark:text-neutral-400',
              'x:hover:text-gray-900',
              'x:dark:hover:text-gray-50',
              'x:contrast-more:text-gray-900',
              'x:contrast-more:dark:text-gray-50',
              'x:hover:bg-gray-100',
              'x:dark:hover:bg-primary-100/5',
              'x:contrast-more:border-transparent',
              'x:contrast-more:hover:border-gray-900',
              'x:contrast-more:dark:hover:border-gray-50'
            )
            
            // Add Nextra's active state classes
            link.classList.add(
              'x:bg-primary-100',
              'x:font-semibold',
              'x:text-primary-800',
              'x:dark:bg-primary-400/10',
              'x:dark:text-primary-600',
              'x:contrast-more:border-primary-500'
            )
            link.setAttribute('aria-current', 'page')
            link.setAttribute('data-active', 'true')
            
            // Expand all parent folders
            let parent = link.parentElement
            while (parent) {
              // Look for the collapse container div (has x:overflow-hidden class)
              if (parent.classList.contains('x:overflow-hidden')) {
                // Expand this folder by removing collapsed styles
                parent.classList.remove('x:opacity-0')
                parent.classList.add('x:opacity-100')
                parent.style.height = 'auto'
                
                // Find the previous sibling button (the folder toggle)
                const prevButton = parent.previousElementSibling
                if (prevButton?.tagName === 'BUTTON') {
                  prevButton.setAttribute('aria-expanded', 'true')
                  
                  // Add 'open' class to parent <li>
                  const parentLi = parent.parentElement
                  if (parentLi?.tagName === 'LI') {
                    parentLi.classList.add('open')
                  }
                  
                  // Rotate the SVG arrow
                  const svg = prevButton.querySelector('svg')
                  if (svg) {
                    svg.classList.remove('x:ltr:rotate-0', 'x:rtl:-rotate-180')
                    svg.classList.add('x:ltr:rotate-90', 'x:rtl:-rotate-270')
                  }
                }
              }
              
              // Move up the DOM tree
              parent = parent.parentElement
              
              // Stop at the sidebar container
              if (parent?.tagName === 'ASIDE' || parent?.tagName === 'NAV') break
            }
          }
        })
      }
      
      // Try multiple times with increasing delays
      setTimeout(activateLink, 0)
      setTimeout(activateLink, 100)
      setTimeout(activateLink, 300)
      setTimeout(activateLink, 500)
      
      // Watch for DOM changes
      const observer = new MutationObserver(activateLink)
      observer.observe(document.body, {
        childList: true,
        subtree: true
      })
      
      return () => observer.disconnect()
    }
  }, [pathname, directRouteMap])
  
  return null
}
