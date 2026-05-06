import { afterEach, describe, expect, it, vi } from 'vitest'
import { applySystemTheme, isSystemDark, watchSystemTheme } from '../theme'

function mockMatchMedia(matches: boolean) {
  const listeners = new Set<(event: MediaQueryListEvent) => void>()

  const mediaQuery = {
    matches,
    media: '(prefers-color-scheme: dark)',
    onchange: null,
    addEventListener: vi.fn((_event: string, listener: (event: MediaQueryListEvent) => void) => {
      listeners.add(listener)
    }),
    removeEventListener: vi.fn((_event: string, listener: (event: MediaQueryListEvent) => void) => {
      listeners.delete(listener)
    }),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  } as unknown as MediaQueryList

  vi.stubGlobal('matchMedia', vi.fn(() => mediaQuery))

  return {
    mediaQuery,
    setMatches(nextMatches: boolean) {
      ;(mediaQuery as { matches: boolean }).matches = nextMatches
      const event = { matches: nextMatches } as MediaQueryListEvent
      listeners.forEach((listener) => listener(event))
    },
  }
}

describe('theme', () => {
  afterEach(() => {
    document.documentElement.classList.remove('dark')
    localStorage.clear()
    vi.unstubAllGlobals()
  })

  it('applies dark class from browser preference and ignores stored theme', () => {
    mockMatchMedia(true)
    localStorage.setItem('theme', 'light')

    expect(applySystemTheme()).toBe(true)
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(localStorage.getItem('theme')).toBeNull()
  })

  it('applies light class from browser preference and ignores stored theme', () => {
    document.documentElement.classList.add('dark')
    mockMatchMedia(false)
    localStorage.setItem('theme', 'dark')

    expect(applySystemTheme()).toBe(false)
    expect(document.documentElement.classList.contains('dark')).toBe(false)
    expect(localStorage.getItem('theme')).toBeNull()
  })

  it('updates when browser preference changes', () => {
    const { mediaQuery, setMatches } = mockMatchMedia(false)
    const onChange = vi.fn()

    const stop = watchSystemTheme(onChange)
    expect(onChange).toHaveBeenLastCalledWith(false)
    expect(document.documentElement.classList.contains('dark')).toBe(false)

    setMatches(true)
    expect(onChange).toHaveBeenLastCalledWith(true)
    expect(document.documentElement.classList.contains('dark')).toBe(true)

    stop()
    expect(mediaQuery.removeEventListener).toHaveBeenCalled()
  })

  it('falls back to light when matchMedia is unavailable', () => {
    vi.stubGlobal('matchMedia', undefined)

    expect(isSystemDark()).toBe(false)
    expect(applySystemTheme()).toBe(false)
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })
})
