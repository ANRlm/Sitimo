import * as React from 'react'

const MOBILE_BREAKPOINT = 768

function subscribe(onStoreChange: () => void) {
  if (typeof window === 'undefined') {
    return () => {}
  }

  const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`)
  const handleChange = () => onStoreChange()

  mql.addEventListener('change', handleChange)
  return () => mql.removeEventListener('change', handleChange)
}

function getSnapshot() {
  if (typeof window === 'undefined') {
    return false
  }

  return window.innerWidth < MOBILE_BREAKPOINT
}

export function useIsMobile() {
  return React.useSyncExternalStore(subscribe, getSnapshot, () => false)
}
