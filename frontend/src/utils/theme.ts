const darkModeQuery = '(prefers-color-scheme: dark)'

function getDarkModeMediaQuery(): MediaQueryList {
  return window.matchMedia(darkModeQuery)
}

export function isSystemDark(): boolean {
  if (typeof window.matchMedia !== 'function') {
    return false
  }

  return getDarkModeMediaQuery().matches
}

export function applySystemTheme(): boolean {
  localStorage.removeItem('theme')
  const shouldUseDark = isSystemDark()
  document.documentElement.classList.toggle('dark', shouldUseDark)
  return shouldUseDark
}

export function watchSystemTheme(onChange?: (isDark: boolean) => void): () => void {
  if (typeof window.matchMedia !== 'function') {
    const isDark = applySystemTheme()
    onChange?.(isDark)
    return () => {}
  }

  const mediaQuery = getDarkModeMediaQuery()
  const syncTheme = () => {
    const isDark = applySystemTheme()
    onChange?.(isDark)
  }

  syncTheme()
  mediaQuery.addEventListener('change', syncTheme)

  return () => mediaQuery.removeEventListener('change', syncTheme)
}
